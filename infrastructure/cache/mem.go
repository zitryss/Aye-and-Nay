package cache

import (
	"context"
	"sync"
	"time"

	"github.com/emirpasic/gods/sets/linkedhashset"
	"github.com/emirpasic/gods/trees/binaryheap"
	"golang.org/x/time/rate"

	"github.com/zitryss/aye-and-nay/domain/model"
	"github.com/zitryss/aye-and-nay/pkg/errors"
)

func NewMem(opts ...options) *Mem {
	conf := newMemConfig()
	m := &Mem{
		conf:         conf,
		syncVisitors: syncVisitors{visitors: map[uint64]*visitorTime{}},
		syncQueues:   syncQueues{queues: map[uint64]*linkedhashset.Set{}},
		syncPQueues:  syncPQueues{pqueues: map[uint64]*binaryheap.Heap{}},
		syncPairs:    syncPairs{pairs: map[uint64]*pairsTime{}},
		syncTokens:   syncTokens{tokens: map[uint64]*tokenTime{}},
	}
	for _, opt := range opts {
		opt(m)
	}
	return m
}

type options func(*Mem)

func WithHeartbeatPair(ch chan<- interface{}) options {
	return func(m *Mem) {
		m.heartbeat.pair = ch
	}
}

func WithHeartbeatToken(ch chan<- interface{}) options {
	return func(m *Mem) {
		m.heartbeat.token = ch
	}
}

type Mem struct {
	conf memConfig
	syncVisitors
	syncQueues
	syncPQueues
	syncPairs
	syncTokens
	heartbeat struct {
		pair  chan<- interface{}
		token chan<- interface{}
	}
}

type syncVisitors struct {
	sync.Mutex
	visitors map[uint64]*visitorTime
}

type visitorTime struct {
	limiter *rate.Limiter
	seen    time.Time
}

type syncQueues struct {
	sync.Mutex
	queues map[uint64]*linkedhashset.Set
}

type syncPQueues struct {
	sync.Mutex
	pqueues map[uint64]*binaryheap.Heap
}

type syncPairs struct {
	sync.Mutex
	pairs map[uint64]*pairsTime
}

type pairsTime struct {
	pairs [][2]uint64
	seen  time.Time
}

type syncTokens struct {
	sync.Mutex
	tokens map[uint64]*tokenTime
}

type tokenTime struct {
	token uint64
	seen  time.Time
}

type elem struct {
	album   uint64
	expires time.Time
}

func timeComparator(a, b interface{}) int {
	tA := a.(elem).expires
	tB := b.(elem).expires
	switch {
	case tA.After(tB):
		return 1
	case tA.Before(tB):
		return -1
	default:
		return 0
	}
}

func (m *Mem) Monitor() {
	go func() {
		for {
			now := time.Now()
			m.syncVisitors.Lock()
			for k, v := range m.visitors {
				if now.Sub(v.seen) >= m.conf.timeToLive {
					delete(m.visitors, k)
				}
			}
			m.syncVisitors.Unlock()
			time.Sleep(m.conf.cleanupInterval)
		}
	}()
	go func() {
		for {
			if m.heartbeat.pair != nil {
				m.heartbeat.pair <- struct{}{}
			}
			now := time.Now()
			m.syncPairs.Lock()
			for k, v := range m.pairs {
				if now.Sub(v.seen) >= m.conf.timeToLive {
					delete(m.pairs, k)
				}
			}
			m.syncPairs.Unlock()
			time.Sleep(m.conf.cleanupInterval)
			if m.heartbeat.pair != nil {
				m.heartbeat.pair <- struct{}{}
			}
		}
	}()
	go func() {
		for {
			if m.heartbeat.token != nil {
				m.heartbeat.token <- struct{}{}
			}
			now := time.Now()
			m.syncTokens.Lock()
			for k, v := range m.tokens {
				if now.Sub(v.seen) >= m.conf.timeToLive {
					delete(m.tokens, k)
				}
			}
			m.syncTokens.Unlock()
			time.Sleep(m.conf.cleanupInterval)
			if m.heartbeat.token != nil {
				m.heartbeat.token <- struct{}{}
			}
		}
	}()
}

func (m *Mem) Allow(_ context.Context, ip uint64) (bool, error) {
	m.syncVisitors.Lock()
	defer m.syncVisitors.Unlock()
	v, ok := m.visitors[ip]
	if !ok {
		l := rate.NewLimiter(rate.Limit(m.conf.limiterRequestsPerSecond), m.conf.limiterBurst)
		v = &visitorTime{limiter: l}
		m.visitors[ip] = v
	}
	v.seen = time.Now()
	return v.limiter.Allow(), nil
}

func (m *Mem) Add(_ context.Context, queue uint64, album uint64) error {
	m.syncQueues.Lock()
	defer m.syncQueues.Unlock()
	q, ok := m.queues[queue]
	if !ok {
		q = linkedhashset.New()
		m.queues[queue] = q
	}
	q.Add(album)
	return nil
}

func (m *Mem) Poll(_ context.Context, queue uint64) (uint64, error) {
	m.syncQueues.Lock()
	defer m.syncQueues.Unlock()
	q, ok := m.queues[queue]
	if !ok {
		return 0x0, errors.Wrap(model.ErrUnknown)
	}
	it := q.Iterator()
	if !it.Next() {
		return 0x0, errors.Wrap(model.ErrUnknown)
	}
	album := it.Value().(uint64)
	q.Remove(album)
	return album, nil
}

func (m *Mem) Size(_ context.Context, queue uint64) (int, error) {
	m.syncQueues.Lock()
	defer m.syncQueues.Unlock()
	q, ok := m.queues[queue]
	if !ok {
		return 0, nil
	}
	n := q.Size()
	return n, nil
}

func (m *Mem) PAdd(_ context.Context, pqueue uint64, album uint64, expires time.Time) error {
	m.syncPQueues.Lock()
	defer m.syncPQueues.Unlock()
	pq, ok := m.pqueues[pqueue]
	if !ok {
		pq = binaryheap.NewWith(timeComparator)
		m.pqueues[pqueue] = pq
	}
	pq.Push(elem{album, expires})
	return nil
}

func (m *Mem) PPoll(_ context.Context, pqueue uint64) (uint64, time.Time, error) {
	m.syncPQueues.Lock()
	defer m.syncPQueues.Unlock()
	pq, ok := m.pqueues[pqueue]
	if !ok {
		return 0x0, time.Time{}, errors.Wrap(model.ErrUnknown)
	}
	e, ok := pq.Pop()
	if !ok {
		return 0x0, time.Time{}, errors.Wrap(model.ErrUnknown)
	}
	return e.(elem).album, e.(elem).expires, nil
}

func (m *Mem) PSize(_ context.Context, pqueue uint64) (int, error) {
	m.syncPQueues.Lock()
	defer m.syncPQueues.Unlock()
	pq, ok := m.pqueues[pqueue]
	if !ok {
		return 0, nil
	}
	n := pq.Size()
	return n, nil
}

func (m *Mem) Push(_ context.Context, album uint64, pairs [][2]uint64) error {
	m.syncPairs.Lock()
	defer m.syncPairs.Unlock()
	p, ok := m.pairs[album]
	if !ok {
		p = &pairsTime{}
		p.pairs = make([][2]uint64, 0, len(pairs))
		m.pairs[album] = p
	}
	for _, images := range pairs {
		p.pairs = append(p.pairs, [2]uint64{images[0], images[1]})
	}
	p.seen = time.Now()
	return nil
}

func (m *Mem) Pop(_ context.Context, album uint64) (uint64, uint64, error) {
	m.syncPairs.Lock()
	defer m.syncPairs.Unlock()
	p, ok := m.pairs[album]
	if !ok {
		return 0x0, 0x0, errors.Wrap(model.ErrPairNotFound)
	}
	if len(p.pairs) == 0 {
		return 0x0, 0x0, errors.Wrap(model.ErrPairNotFound)
	}
	images := (p.pairs)[0]
	p.pairs = (p.pairs)[1:]
	p.seen = time.Now()
	return images[0], images[1], nil
}

func (m *Mem) Set(_ context.Context, _ uint64, token uint64, image uint64) error {
	m.syncTokens.Lock()
	defer m.syncTokens.Unlock()
	_, ok := m.tokens[token]
	if ok {
		return errors.Wrap(model.ErrTokenAlreadyExists)
	}
	t := &tokenTime{}
	t.token = image
	t.seen = time.Now()
	m.tokens[token] = t
	return nil
}

func (m *Mem) Get(_ context.Context, _ uint64, token uint64) (uint64, error) {
	m.syncTokens.Lock()
	defer m.syncTokens.Unlock()
	image, ok := m.tokens[token]
	if !ok {
		return 0x0, errors.Wrap(model.ErrTokenNotFound)
	}
	delete(m.tokens, token)
	return image.token, nil
}
