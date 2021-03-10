package service

import (
	"context"
	"math/rand"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/zitryss/aye-and-nay/domain/model"
	"github.com/zitryss/aye-and-nay/pkg/errors"
	"github.com/zitryss/aye-and-nay/pkg/linalg"
	myrand "github.com/zitryss/aye-and-nay/pkg/rand"
)

func New(
	comp model.Compresser,
	stor model.Storager,
	pers model.Databaser,
	temp model.Cacher,
	qCalc *QueueCalc,
	qComp *QueueComp,
	qDel *QueueDel,
	opts ...options,
) *Service {
	conf := newServiceConfig()
	s := &Service{
		conf:  conf,
		comp:  comp,
		stor:  stor,
		pers:  pers,
		pair:  temp,
		token: temp,
		queue: struct {
			calc *QueueCalc
			comp *QueueComp
			del  *QueueDel
		}{
			qCalc,
			qComp,
			qDel,
		},
		rand: struct {
			id      func(length int) (string, error)
			shuffle func(n int, swap func(i int, j int))
			now     func() time.Time
		}{
			myrand.Id,
			rand.Shuffle,
			time.Now,
		},
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

func NewQueueCalc(q model.Queuer) *QueueCalc {
	return &QueueCalc{newQueue("calculation", q)}
}

type QueueCalc struct {
	*queue
}

func NewQueueComp(q model.Queuer) *QueueComp {
	return &QueueComp{newQueue("compression", q)}
}

type QueueComp struct {
	*queue
}

func NewQueueDel(q model.PQueuer) *QueueDel {
	return &QueueDel{newPQueue("deletion", q)}
}

type QueueDel struct {
	*pqueue
}

type options func(*Service)

func WithRandId(fn func(int) (string, error)) options {
	return func(s *Service) {
		s.rand.id = fn
	}
}

func WithRandShuffle(fn func(int, func(int, int))) options {
	return func(s *Service) {
		s.rand.shuffle = fn
	}
}

func WithRandNow(fn func() time.Time) options {
	return func(s *Service) {
		s.rand.now = fn
	}
}

func WithHeartbeatCalc(ch chan<- interface{}) options {
	return func(s *Service) {
		s.heartbeat.calc = ch
	}
}

func WithHeartbeatComp(ch chan<- interface{}) options {
	return func(s *Service) {
		s.heartbeat.comp = ch
	}
}

func WithHeartbeatDel(ch chan<- interface{}) options {
	return func(s *Service) {
		s.heartbeat.del = ch
	}
}

type Service struct {
	conf  serviceConfig
	comp  model.Compresser
	stor  model.Storager
	pers  model.Databaser
	pair  model.Stacker
	token model.Tokener
	queue struct {
		calc *QueueCalc
		comp *QueueComp
		del  *QueueDel
	}
	rand struct {
		id      func(length int) (string, error)
		shuffle func(n int, swap func(i, j int))
		now     func() time.Time
	}
	heartbeat struct {
		calc chan<- interface{}
		comp chan<- interface{}
		del  chan<- interface{}
	}
}

func (s *Service) StartWorkingPoolCalc(ctx context.Context, g *errgroup.Group) {
	go func() {
		sem := make(chan struct{}, s.conf.numberOfWorkersCalc)
		for {
			select {
			case sem <- struct{}{}:
			case <-ctx.Done():
				return
			}
			g.Go(func() (e error) {
				defer func() { <-sem }()
				defer func() {
					v := recover()
					if v == nil {
						return
					}
					err, ok := v.(error)
					if ok {
						e = errors.Wrap(err)
					} else {
						e = errors.Wrapf(model.ErrUnknown, "%v", v)
					}
				}()
				for {
					select {
					case <-ctx.Done():
						return
					default:
					}
					album, err := s.queue.calc.poll(ctx)
					if err != nil {
						err = errors.Wrap(err)
						handleError(err)
						e = err
						continue
					}
					select {
					case <-ctx.Done():
						return
					default:
					}
					edgs, err := s.pers.GetEdges(ctx, album)
					if err != nil {
						err = errors.Wrap(err)
						handleError(err)
						e = err
						continue
					}
					vect := linalg.PageRank(edgs)
					err = s.pers.UpdateRatings(ctx, album, vect)
					if err != nil {
						err = errors.Wrap(err)
						handleError(err)
						e = err
						continue
					}
					if s.heartbeat.calc != nil {
						s.heartbeat.calc <- struct{}{}
					}
				}
			})
		}
	}()
}

func (s *Service) StartWorkingPoolComp(ctx context.Context, g *errgroup.Group) {
	go func() {
		sem := make(chan struct{}, s.conf.numberOfWorkersComp)
		for {
			select {
			case sem <- struct{}{}:
			case <-ctx.Done():
				return
			}
			g.Go(func() (e error) {
				defer func() { <-sem }()
				defer func() {
					v := recover()
					if v == nil {
						return
					}
					err, ok := v.(error)
					if ok {
						e = errors.Wrap(err)
					} else {
						e = errors.Wrapf(model.ErrUnknown, "%v", v)
					}
				}()
			outer:
				for {
					select {
					case <-ctx.Done():
						return
					default:
					}
					album, err := s.queue.comp.poll(ctx)
					if err != nil {
						err = errors.Wrap(err)
						handleError(err)
						e = err
						continue
					}
					select {
					case <-ctx.Done():
						return
					default:
					}
					images, err := s.pers.GetImages(ctx, album)
					if err != nil {
						err = errors.Wrap(err)
						handleError(err)
						e = err
						continue
					}
					for _, image := range images {
						f, err := s.stor.Get(ctx, album, image)
						if err != nil {
							err = errors.Wrap(err)
							handleError(err)
							e = err
							continue outer
						}
						f, err = s.comp.Compress(ctx, f)
						if errors.Is(err, model.ErrThirdPartyUnavailable) {
							if s.heartbeat.comp != nil {
								s.heartbeat.comp <- err
							}
						}
						if err != nil {
							err = errors.Wrap(err)
							handleError(err)
							e = err
							continue outer
						}
						err = s.stor.Remove(ctx, album, image)
						if err != nil {
							err = errors.Wrap(err)
							handleError(err)
							e = err
							continue outer
						}
						_, err = s.stor.Put(ctx, album, image, f)
						if err != nil {
							err = errors.Wrap(err)
							handleError(err)
							e = err
							continue outer
						}
						err = s.pers.UpdateCompressionStatus(ctx, album, image)
						if err != nil {
							err = errors.Wrap(err)
							handleError(err)
							e = err
							continue outer
						}
						if s.heartbeat.comp != nil {
							p, _ := s.Progress(ctx, album)
							s.heartbeat.comp <- p
						}
					}
				}
			})
		}
	}()
}

func (s *Service) StartWorkingPoolDel(ctx context.Context, g *errgroup.Group) {
	g.Go(func() (e error) {
		defer func() {
			v := recover()
			if v == nil {
				return
			}
			err, ok := v.(error)
			if ok {
				e = errors.Wrap(err)
			} else {
				e = errors.Wrapf(model.ErrUnknown, "%v", v)
			}
		}()
		for {
			select {
			case <-ctx.Done():
				return
			default:
			}
			album, err := s.queue.del.poll(ctx)
			if err != nil {
				err = errors.Wrap(err)
				handleError(err)
				e = err
				continue
			}
			select {
			case <-ctx.Done():
				return
			default:
			}
			images, err := s.pers.GetImages(ctx, album)
			if err != nil {
				err = errors.Wrap(err)
				handleError(err)
				e = err
				continue
			}
			err = s.pers.DeleteAlbum(ctx, album)
			if err != nil {
				err = errors.Wrap(err)
				handleError(err)
				e = err
				continue
			}
			for _, image := range images {
				err = s.stor.Remove(ctx, album, image)
				if err != nil {
					err = errors.Wrap(err)
					handleError(err)
					e = err
					continue
				}
			}
			if s.heartbeat.del != nil {
				s.heartbeat.del <- struct{}{}
			}
		}
	})
}

func (s *Service) Album(ctx context.Context, ff []model.File, dur time.Duration) (uint64, error) {
	album, err := s.rand.id(s.conf.albumIdLength)
	if err != nil {
		return "", errors.Wrap(err)
	}
	imgs := make([]model.Image, 0, len(ff))
	for _, f := range ff {
		image, err := s.rand.id(s.conf.imageIdLength)
		if err != nil {
			return "", errors.Wrap(err)
		}
		src, err := s.stor.Put(ctx, album, image, f)
		if err != nil {
			return "", errors.Wrap(err)
		}
		img := model.Image{}
		img.Id = image
		img.Src = src
		imgs = append(imgs, img)
	}
	edgs := map[string]map[string]int(nil)
	alb := model.Album{album, imgs, edgs}
	err = s.pers.SaveAlbum(ctx, alb)
	if err != nil {
		return "", errors.Wrap(err)
	}
	err = s.queue.comp.add(ctx, album)
	if err != nil {
		return "", errors.Wrap(err)
	}
	err = s.queue.del.add(ctx, album, s.rand.now().Add(dur))
	if err != nil {
		return "", errors.Wrap(err)
	}
	return alb.Id, nil
}

func (s *Service) Progress(ctx context.Context, album uint64) (float64, error) {
	all, err := s.pers.CountImages(ctx, album)
	if err != nil {
		return 0, errors.Wrap(err)
	}
	n, err := s.pers.CountImagesCompressed(ctx, album)
	if err != nil {
		return 0, errors.Wrap(err)
	}
	return float64(n) / float64(all), nil
}

func (s *Service) Pair(ctx context.Context, album uint64) (model.Image, model.Image, error) {
	image1, image2, err := s.pair.Pop(ctx, album)
	if errors.Is(err, model.ErrPairNotFound) {
		err = s.genPairs(ctx, album)
		if err != nil {
			return model.Image{}, model.Image{}, errors.Wrap(err)
		}
		image1, image2, err = s.pair.Pop(ctx, album)
	}
	if err != nil {
		return model.Image{}, model.Image{}, errors.Wrap(err)
	}
	img1, err := s.pers.GetImage(ctx, album, image1)
	if err != nil {
		return model.Image{}, model.Image{}, errors.Wrap(err)
	}
	img2, err := s.pers.GetImage(ctx, album, image2)
	if err != nil {
		return model.Image{}, model.Image{}, errors.Wrap(err)
	}
	token1, err := s.rand.id(s.conf.tokenIdLength)
	if err != nil {
		return model.Image{}, model.Image{}, errors.Wrap(err)
	}
	err = s.token.Set(ctx, album, token1, img1.Id)
	if err != nil {
		return model.Image{}, model.Image{}, errors.Wrap(err)
	}
	img1.Token = token1
	token2, err := s.rand.id(s.conf.tokenIdLength)
	if err != nil {
		return model.Image{}, model.Image{}, errors.Wrap(err)
	}
	err = s.token.Set(ctx, album, token2, img2.Id)
	if err != nil {
		return model.Image{}, model.Image{}, errors.Wrap(err)
	}
	img2.Token = token2
	return img1, img2, nil
}

func (s *Service) genPairs(ctx context.Context, album string) error {
	images, err := s.pers.GetImages(ctx, album)
	if err != nil {
		return errors.Wrap(err)
	}
	s.rand.shuffle(len(images), func(i, j int) { images[i], images[j] = images[j], images[i] })
	images = append(images, images[0])
	pairs := make([][2]string, 0, len(images)-1)
	for i := 0; i < len(images)-1; i++ {
		image1 := images[i]
		image2 := images[i+1]
		pairs = append(pairs, [2]string{image1, image2})
	}
	s.rand.shuffle(len(pairs), func(i, j int) { pairs[i], pairs[j] = pairs[j], pairs[i] })
	err = s.pair.Push(ctx, album, pairs)
	if err != nil {
		return errors.Wrap(err)
	}
	return nil
}

func (s *Service) Vote(ctx context.Context, album uint64, tokenFrom uint64, tokenTo uint64) error {
	imageFrom, err := s.token.Get(ctx, album, tokenFrom)
	if err != nil {
		return errors.Wrap(err)
	}
	imageTo, err := s.token.Get(ctx, album, tokenTo)
	if err != nil {
		return errors.Wrap(err)
	}
	err = s.pers.SaveVote(ctx, album, imageFrom, imageTo)
	if err != nil {
		return errors.Wrap(err)
	}
	err = s.queue.calc.add(ctx, album)
	if err != nil {
		return errors.Wrap(err)
	}
	return nil
}

func (s *Service) Top(ctx context.Context, album uint64) ([]model.Image, error) {
	imgs, err := s.pers.GetImagesOrdered(ctx, album)
	if err != nil {
		return nil, errors.Wrap(err)
	}
	return imgs, nil
}
