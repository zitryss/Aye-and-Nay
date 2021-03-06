package service

import (
	"context"
	"reflect"
	"testing"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/zitryss/aye-and-nay/domain/model"
	"github.com/zitryss/aye-and-nay/infrastructure/cache"
	"github.com/zitryss/aye-and-nay/infrastructure/compressor"
	"github.com/zitryss/aye-and-nay/infrastructure/database"
	"github.com/zitryss/aye-and-nay/infrastructure/storage"
	_ "github.com/zitryss/aye-and-nay/internal/config"
	. "github.com/zitryss/aye-and-nay/internal/testing"
	"github.com/zitryss/aye-and-nay/pkg/errors"
)

func TestServiceAlbum(t *testing.T) {
	t.Run("Positive", func(t *testing.T) {
		fn1 := func() func() (uint64, error) {
			i := uint64(0)
			return func() (uint64, error) {
				i++
				return 0x463E + i, nil
			}
		}()
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		comp := compressor.NewMock()
		stor := storage.NewMock()
		mDb := database.NewMem()
		mCache := cache.NewMem()
		qCalc := &QueueCalc{}
		qCalc.Monitor(ctx)
		qComp := &QueueComp{newQueue(0xB273, mCache)}
		qComp.Monitor(ctx)
		qDel := &QueueDel{}
		qDel.Monitor(ctx)
		heartbeatComp := make(chan interface{})
		serv := New(comp, stor, mDb, mCache, qCalc, qComp, qDel, WithRandId(fn1), WithHeartbeatComp(heartbeatComp))
		gComp, ctxComp := errgroup.WithContext(ctx)
		serv.StartWorkingPoolComp(ctxComp, gComp)
		files := []model.File{Png(), Png()}
		_, err := serv.Album(ctx, files, 0*time.Millisecond)
		if err != nil {
			t.Error(err)
		}
		v := CheckChannel(t, heartbeatComp)
		p, ok := v.(float64)
		if !ok {
			t.Error("v.(type) != float64")
		}
		if !EqualFloat(p, 0.5) {
			t.Error("p != 0.5")
		}
		v = CheckChannel(t, heartbeatComp)
		p, ok = v.(float64)
		if !ok {
			t.Error("v.(type) != float64")
		}
		if !EqualFloat(p, 1) {
			t.Error("p != 1")
		}
	})
	t.Run("Negative", func(t *testing.T) {
		fn1 := func() func() (uint64, error) {
			i := uint64(0)
			return func() (uint64, error) {
				i++
				return 0x915C + i, nil
			}
		}()
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		heartbeatRestart := make(chan interface{})
		comp := compressor.NewShortPixel(compressor.WithHeartbeatRestart(heartbeatRestart))
		comp.Monitor()
		stor := storage.NewMock()
		mDb := database.NewMem()
		mCache := cache.NewMem()
		qCalc := &QueueCalc{}
		qCalc.Monitor(ctx)
		qComp := &QueueComp{newQueue(0x88AB, mCache)}
		qComp.Monitor(ctx)
		qDel := &QueueDel{}
		qDel.Monitor(ctx)
		heartbeatComp := make(chan interface{})
		serv := New(comp, stor, mDb, mCache, qCalc, qComp, qDel, WithRandId(fn1), WithHeartbeatComp(heartbeatComp))
		gComp, ctxComp := errgroup.WithContext(ctx)
		serv.StartWorkingPoolComp(ctxComp, gComp)
		files := []model.File{Png(), Png()}
		_, err := serv.Album(ctx, files, 0*time.Millisecond)
		if err != nil {
			t.Error(err)
		}
		v := CheckChannel(t, heartbeatComp)
		_ = CheckChannel(t, heartbeatComp)
		err, ok := v.(error)
		if !ok {
			t.Error("v.(type) != error")
		}
		if !errors.Is(err, model.ErrThirdPartyUnavailable) {
			t.Error(err)
		}
		files = []model.File{Png(), Png()}
		_, err = serv.Album(ctx, files, 0*time.Millisecond)
		if err != nil {
			t.Error(err)
		}
		v = CheckChannel(t, heartbeatComp)
		p, ok := v.(float64)
		if !ok {
			t.Error("v.(type) != float64")
		}
		if !EqualFloat(p, 0.5) {
			t.Error("p != 0.5")
		}
		v = CheckChannel(t, heartbeatComp)
		p, ok = v.(float64)
		if !ok {
			t.Error("v.(type) != float64")
		}
		if !EqualFloat(p, 1) {
			t.Error("p != 1")
		}
		CheckChannel(t, heartbeatRestart)
		CheckChannel(t, heartbeatRestart)
		files = []model.File{Png(), Png()}
		_, err = serv.Album(ctx, files, 0*time.Millisecond)
		if err != nil {
			t.Error(err)
		}
		v = CheckChannel(t, heartbeatComp)
		_ = CheckChannel(t, heartbeatComp)
		err, ok = v.(error)
		if !ok {
			t.Error("v.(type) != error")
		}
		if !errors.Is(err, model.ErrThirdPartyUnavailable) {
			t.Error(err)
		}
		files = []model.File{Png(), Png()}
		_, err = serv.Album(ctx, files, 0*time.Millisecond)
		if err != nil {
			t.Error(err)
		}
		v = CheckChannel(t, heartbeatComp)
		p, ok = v.(float64)
		if !ok {
			t.Error("v.(type) != float64")
		}
		if !EqualFloat(p, 0.5) {
			t.Error("p != 0.5")
		}
		v = CheckChannel(t, heartbeatComp)
		p, ok = v.(float64)
		if !ok {
			t.Error("v.(type) != float64")
		}
		if !EqualFloat(p, 1) {
			t.Error("p != 1")
		}
	})
}

func TestServicePair(t *testing.T) {
	t.Run("Positive", func(t *testing.T) {
		fn1 := func() func() (uint64, error) {
			i := uint64(0)
			return func() (uint64, error) {
				i++
				return 0x3BC5 + i, nil
			}
		}()
		fn2 := func(n int, swap func(i int, j int)) {
		}
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		comp := compressor.NewMock()
		stor := storage.NewMock()
		mDb := database.NewMem()
		mCache := cache.NewMem()
		qCalc := &QueueCalc{}
		qCalc.Monitor(ctx)
		qComp := &QueueComp{}
		qComp.Monitor(ctx)
		qDel := &QueueDel{}
		qDel.Monitor(ctx)
		serv := New(comp, stor, mDb, mCache, qCalc, qComp, qDel, WithRandId(fn1), WithRandShuffle(fn2))
		files := []model.File{Png(), Png()}
		album, err := serv.Album(ctx, files, 0*time.Millisecond)
		if err != nil {
			t.Error(err)
		}
		img7, img8, err := serv.Pair(ctx, album)
		if err != nil {
			t.Error(err)
		}
		img1 := model.Image{Id: 0x3BC7, Token: 0x3BC9, Src: "/aye-and-nay/albums/xjsAAAAAAAA/images/xzsAAAAAAAA"}
		img2 := model.Image{Id: 0x3BC8, Token: 0x3BCA, Src: "/aye-and-nay/albums/xjsAAAAAAAA/images/yDsAAAAAAAA"}
		imgs1 := []model.Image{img1, img2}
		if reflect.DeepEqual(img7, img8) {
			t.Error("img7 == img8")
		}
		if !IsIn(img7, imgs1) {
			t.Error("img7 is not in imgs")
		}
		if !IsIn(img8, imgs1) {
			t.Error("img8 is not in imgs")
		}
		img9, img10, err := serv.Pair(ctx, album)
		if err != nil {
			t.Error(err)
		}
		img3 := model.Image{Id: 0x3BC8, Token: 0x3BCB, Src: "/aye-and-nay/albums/xjsAAAAAAAA/images/yDsAAAAAAAA"}
		img4 := model.Image{Id: 0x3BC7, Token: 0x3BCC, Src: "/aye-and-nay/albums/xjsAAAAAAAA/images/xzsAAAAAAAA"}
		imgs2 := []model.Image{img3, img4}
		if reflect.DeepEqual(img9, img10) {
			t.Error("img9 == img10")
		}
		if !IsIn(img9, imgs2) {
			t.Error("img9 is not in imgs")
		}
		if !IsIn(img10, imgs2) {
			t.Error("img10 is not in imgs")
		}
		img11, img12, err := serv.Pair(ctx, album)
		if err != nil {
			t.Error(err)
		}
		img5 := model.Image{Id: 0x3BC7, Token: 0x3BCD, Src: "/aye-and-nay/albums/xjsAAAAAAAA/images/xzsAAAAAAAA"}
		img6 := model.Image{Id: 0x3BC8, Token: 0x3BCE, Src: "/aye-and-nay/albums/xjsAAAAAAAA/images/yDsAAAAAAAA"}
		imgs3 := []model.Image{img5, img6}
		if reflect.DeepEqual(img11, img12) {
			t.Error("img11 == img12")
		}
		if !IsIn(img11, imgs3) {
			t.Error("img11 is not in imgs")
		}
		if !IsIn(img12, imgs3) {
			t.Error("img12 is not in imgs")
		}
	})
	t.Run("Negative", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		comp := compressor.NewMock()
		stor := storage.NewMock()
		mDb := database.NewMem()
		mCache := cache.NewMem()
		qCalc := &QueueCalc{}
		qCalc.Monitor(ctx)
		qComp := &QueueComp{}
		qComp.Monitor(ctx)
		qDel := &QueueDel{}
		qDel.Monitor(ctx)
		serv := New(comp, stor, mDb, mCache, qCalc, qComp, qDel)
		_, _, err := serv.Pair(ctx, 0xEB46)
		if !errors.Is(err, model.ErrAlbumNotFound) {
			t.Error(err)
		}
	})
}

func TestServiceVote(t *testing.T) {
	t.Run("Positive", func(t *testing.T) {
		fn1 := func() func() (uint64, error) {
			i := uint64(0)
			return func() (uint64, error) {
				i++
				return 0xC389 + i, nil
			}
		}()
		fn2 := func(n int, swap func(i int, j int)) {
		}
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		comp := compressor.NewMock()
		stor := storage.NewMock()
		mDb := database.NewMem()
		mCache := cache.NewMem()
		qCalc := &QueueCalc{}
		qCalc.Monitor(ctx)
		qComp := &QueueComp{}
		qComp.Monitor(ctx)
		qDel := &QueueDel{}
		qDel.Monitor(ctx)
		serv := New(comp, stor, mDb, mCache, qCalc, qComp, qDel, WithRandId(fn1), WithRandShuffle(fn2))
		files := []model.File{Png(), Png()}
		album, err := serv.Album(ctx, files, 0*time.Millisecond)
		if err != nil {
			t.Error(err)
		}
		img1, img2, err := serv.Pair(ctx, album)
		if err != nil {
			t.Error(err)
		}
		err = serv.Vote(ctx, album, img1.Token, img2.Token)
		if err != nil {
			t.Error(err)
		}
	})
	t.Run("Negative1", func(t *testing.T) {
		fn1 := func() func() (uint64, error) {
			i := uint64(0)
			return func() (uint64, error) {
				i++
				return 0xE24F + i, nil
			}
		}()
		fn2 := func(n int, swap func(i int, j int)) {
		}
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		comp := compressor.NewMock()
		stor := storage.NewMock()
		mDb := database.NewMem()
		mCache := cache.NewMem()
		qCalc := &QueueCalc{}
		qCalc.Monitor(ctx)
		qComp := &QueueComp{}
		qComp.Monitor(ctx)
		qDel := &QueueDel{}
		qDel.Monitor(ctx)
		serv := New(comp, stor, mDb, mCache, qCalc, qComp, qDel, WithRandId(fn1), WithRandShuffle(fn2))
		files := []model.File{Png(), Png()}
		album, err := serv.Album(ctx, files, 0*time.Millisecond)
		if err != nil {
			t.Error(err)
		}
		img1, img2, err := serv.Pair(ctx, album)
		if err != nil {
			t.Error(err)
		}
		err = serv.Vote(ctx, 0x12E6, img1.Token, img2.Token)
		if !errors.Is(err, model.ErrAlbumNotFound) {
			t.Error(err)
		}
	})
	t.Run("Negative2", func(t *testing.T) {
		fn1 := func() func() (uint64, error) {
			i := uint64(0)
			return func() (uint64, error) {
				i++
				return 0xBC43 + i, nil
			}
		}()
		fn2 := func(n int, swap func(i int, j int)) {
		}
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		comp := compressor.NewMock()
		stor := storage.NewMock()
		mDb := database.NewMem()
		mCache := cache.NewMem()
		qCalc := &QueueCalc{}
		qCalc.Monitor(ctx)
		qComp := &QueueComp{}
		qComp.Monitor(ctx)
		qDel := &QueueDel{}
		qDel.Monitor(ctx)
		serv := New(comp, stor, mDb, mCache, qCalc, qComp, qDel, WithRandId(fn1), WithRandShuffle(fn2))
		files := []model.File{Png(), Png()}
		album, err := serv.Album(ctx, files, 0*time.Millisecond)
		if err != nil {
			t.Error(err)
		}
		_, _, err = serv.Pair(ctx, album)
		if err != nil {
			t.Error(err)
		}
		err = serv.Vote(ctx, album, 0x1CC1, 0xF83C)
		if !errors.Is(err, model.ErrTokenNotFound) {
			t.Error(err)
		}
	})
}

func TestServiceTop(t *testing.T) {
	t.Run("Positive", func(t *testing.T) {
		fn1 := func() func() (uint64, error) {
			i := uint64(0)
			return func() (uint64, error) {
				i++
				return 0x4DB8 + i, nil
			}
		}()
		fn2 := func(n int, swap func(i int, j int)) {
		}
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		comp := compressor.NewMock()
		stor := storage.NewMock()
		mDb := database.NewMem()
		mCache := cache.NewMem()
		qCalc := &QueueCalc{newQueue(0x1A01, mCache)}
		qCalc.Monitor(ctx)
		qComp := &QueueComp{}
		qComp.Monitor(ctx)
		qDel := &QueueDel{}
		qDel.Monitor(ctx)
		heartbeatCalc := make(chan interface{})
		serv := New(comp, stor, mDb, mCache, qCalc, qComp, qDel, WithRandId(fn1), WithRandShuffle(fn2), WithHeartbeatCalc(heartbeatCalc))
		gCalc, ctxCalc := errgroup.WithContext(ctx)
		serv.StartWorkingPoolCalc(ctxCalc, gCalc)
		files := []model.File{Png(), Png()}
		album, err := serv.Album(ctx, files, 0*time.Millisecond)
		if err != nil {
			t.Error(err)
		}
		img1, img2, err := serv.Pair(ctx, album)
		if err != nil {
			t.Error(err)
		}
		err = serv.Vote(ctx, album, img1.Token, img2.Token)
		if err != nil {
			t.Error(err)
		}
		CheckChannel(t, heartbeatCalc)
		img3, img4, err := serv.Pair(ctx, album)
		if err != nil {
			t.Error(err)
		}
		err = serv.Vote(ctx, album, img3.Token, img4.Token)
		if err != nil {
			t.Error(err)
		}
		CheckChannel(t, heartbeatCalc)
		imgs1, err := serv.Top(ctx, album)
		if err != nil {
			t.Error(err)
		}
		img5 := model.Image{Id: 0x4DBA, Src: "/aye-and-nay/albums/uU0AAAAAAAA/images/uk0AAAAAAAA", Rating: 0.5, Compressed: false}
		img6 := model.Image{Id: 0x4DBB, Src: "/aye-and-nay/albums/uU0AAAAAAAA/images/u00AAAAAAAA", Rating: 0.5, Compressed: false}
		imgs2 := []model.Image{img5, img6}
		if !reflect.DeepEqual(imgs1, imgs2) {
			t.Error("imgs1 != imgs2")
		}
	})
	t.Run("Negative", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		comp := compressor.NewMock()
		stor := storage.NewMock()
		mDb := database.NewMem()
		mCache := cache.NewMem()
		qCalc := &QueueCalc{}
		qCalc.Monitor(ctx)
		qComp := &QueueComp{}
		qComp.Monitor(ctx)
		qDel := &QueueDel{}
		qDel.Monitor(ctx)
		serv := New(comp, stor, mDb, mCache, qCalc, qComp, qDel)
		_, err := serv.Top(ctx, 0x83CD)
		if !errors.Is(err, model.ErrAlbumNotFound) {
			t.Error(err)
		}
	})
}

func TestServiceDelete(t *testing.T) {
	t.Run("Positive1", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		comp := compressor.NewMock()
		stor := storage.NewMock()
		mDb := database.NewMem()
		mCache := cache.NewMem()
		qCalc := &QueueCalc{}
		qCalc.Monitor(ctx)
		qComp := &QueueComp{}
		qComp.Monitor(ctx)
		qDel := &QueueDel{newPQueue(0xE3FF, mCache)}
		qDel.Monitor(ctx)
		alb1 := AlbumEmptyFactory(0x101F)
		alb1.Expires = time.Now().Add(-1 * time.Hour)
		err := mDb.SaveAlbum(ctx, alb1)
		if err != nil {
			t.Error(err)
		}
		alb2 := AlbumEmptyFactory(0xFFBB)
		alb2.Expires = time.Now().Add(1 * time.Hour)
		err = mDb.SaveAlbum(ctx, alb2)
		if err != nil {
			t.Error(err)
		}
		heartbeatDel := make(chan interface{})
		serv := New(comp, stor, mDb, mCache, qCalc, qComp, qDel, WithHeartbeatDel(heartbeatDel))
		err = serv.CleanUp(ctx)
		if err != nil {
			t.Error(err)
		}
		gDel, ctxDel := errgroup.WithContext(ctx)
		serv.StartWorkingPoolDel(ctxDel, gDel)
		v := CheckChannel(t, heartbeatDel)
		album, ok := v.(uint64)
		if !ok {
			t.Error("v.(type) != uint64")
		}
		if album != 0x101F {
			t.Error("album != 0x101F")
		}
	})
	t.Run("Positive2", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		comp := compressor.NewMock()
		stor := storage.NewMock()
		mDb := database.NewMem()
		mCache := cache.NewMem()
		qCalc := &QueueCalc{}
		qCalc.Monitor(ctx)
		qComp := &QueueComp{}
		qComp.Monitor(ctx)
		qDel := &QueueDel{newPQueue(0xEF3F, mCache)}
		qDel.Monitor(ctx)
		heartbeatDel := make(chan interface{})
		serv := New(comp, stor, mDb, mCache, qCalc, qComp, qDel, WithHeartbeatDel(heartbeatDel))
		gDel, ctxDel := errgroup.WithContext(ctx)
		serv.StartWorkingPoolDel(ctxDel, gDel)
		files := []model.File{Png(), Png()}
		dur := 100 * time.Millisecond
		album, err := serv.Album(ctx, files, dur)
		if err != nil {
			t.Error(err)
		}
		CheckChannel(t, heartbeatDel)
		_, err = serv.Top(ctx, album)
		if !errors.Is(err, model.ErrAlbumNotFound) {
			t.Error(err)
		}
	})
	t.Run("Negative", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		comp := compressor.NewMock()
		stor := storage.NewMock()
		mDb := database.NewMem()
		mCache := cache.NewMem()
		qCalc := &QueueCalc{}
		qCalc.Monitor(ctx)
		qComp := &QueueComp{}
		qComp.Monitor(ctx)
		qDel := &QueueDel{newPQueue(0xEF3F, mCache)}
		qDel.Monitor(ctx)
		heartbeatDel := make(chan interface{})
		serv := New(comp, stor, mDb, mCache, qCalc, qComp, qDel, WithHeartbeatDel(heartbeatDel))
		gDel, ctxDel := errgroup.WithContext(ctx)
		serv.StartWorkingPoolDel(ctxDel, gDel)
		files := []model.File{Png(), Png()}
		dur := 0 * time.Second
		album, err := serv.Album(ctx, files, dur)
		if err != nil {
			t.Error(err)
		}
		select {
		case <-heartbeatDel:
			t.Error("<-heartbeatDel")
		case <-time.After(1 * time.Second):
		}
		_, err = serv.Top(ctx, album)
		if err != nil {
			t.Error(err)
		}
	})
}
