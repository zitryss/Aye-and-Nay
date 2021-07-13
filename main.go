package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path"
	"strings"
	"syscall"

	"github.com/spf13/viper"
	"golang.org/x/sync/errgroup"

	"github.com/zitryss/aye-and-nay/delivery/http"
	"github.com/zitryss/aye-and-nay/domain/service"
	"github.com/zitryss/aye-and-nay/infrastructure/cache"
	"github.com/zitryss/aye-and-nay/infrastructure/compressor"
	"github.com/zitryss/aye-and-nay/infrastructure/database"
	"github.com/zitryss/aye-and-nay/infrastructure/storage"
	"github.com/zitryss/aye-and-nay/pkg/errors"
	"github.com/zitryss/aye-and-nay/pkg/log"
)

var (
	ballast []byte
)

func main() {
	conf := ""
	flag.StringVar(&conf, "config", "./config.yml", "relative filepath to a config file")
	flag.Parse()
	dir, file := path.Split(conf)
	base := strings.TrimSuffix(file, path.Ext(file))

	viper.SetConfigName(base)
	viper.AddConfigPath(dir)
	err := viper.ReadInConfig()
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "critical:", err)
		os.Exit(1)
	}

	ballast = make([]byte, viper.GetInt64("app.ballast"))

	lvl := viper.GetString("app.log")
	log.SetOutput(os.Stderr)
	log.SetLevel(lvl)
	log.Info("logging initialized")

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	cach, err := cache.New(viper.GetString("cache.use"))
	if err != nil {
		log.Critical(err)
		os.Exit(1)
	}

	comp, err := compressor.New(viper.GetString("compressor.use"))
	if err != nil {
		log.Critical(err)
		os.Exit(1)
	}

	data, err := database.New(viper.GetString("database.use"))
	if err != nil {
		log.Critical(err)
		os.Exit(1)
	}

	stor, err := storage.New(viper.GetString("storage.use"))
	if err != nil {
		log.Critical(err)
		os.Exit(1)
	}

	qCalc := service.NewQueueCalc(cach)
	qCalc.Monitor(ctx)

	qComp := &service.QueueComp{}
	if viper.GetString("compressor.use") != "mock" {
		qComp = service.NewQueueComp(cach)
		qComp.Monitor(ctx)
	}

	qDel := service.NewQueueDel(cach)
	qDel.Monitor(ctx)

	serv := service.New(comp, stor, data, cach, qCalc, qComp, qDel)
	err = serv.CleanUp(ctx)
	if err != nil {
		log.Critical(err)
		os.Exit(1)
	}

	gCalc, ctxCalc := errgroup.WithContext(ctx)
	log.Info("starting calculation worker pool")
	serv.StartWorkingPoolCalc(ctxCalc, gCalc)

	gComp := (*errgroup.Group)(nil)
	ctxComp := context.Context(nil)
	if viper.GetString("compressor.use") != "mock" {
		gComp, ctxComp = errgroup.WithContext(ctx)
		log.Info("starting compression worker pool")
		serv.StartWorkingPoolComp(ctxComp, gComp)
	}

	gDel, ctxDel := errgroup.WithContext(ctx)
	log.Info("starting deletion worker pool")
	serv.StartWorkingPoolDel(ctxDel, gDel)

	middle := http.NewMiddleware(cach)
	srvWait := make(chan error, 1)
	srv := http.NewServer(middle.Chain, serv, srvWait)
	srv.Monitor(ctx)
	log.Info("starting web server")
	err = srv.Start()

	log.Info("stopping web server")
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Error(err)
	}
	err = <-srvWait
	if err != nil {
		log.Error(err)
	}

	log.Info("stopping deletion worker pool")
	err = gDel.Wait()
	if err != nil {
		log.Error(err)
	}

	if viper.GetString("compressor.use") != "mock" {
		log.Info("stopping compression worker pool")
		err = gComp.Wait()
		if err != nil {
			log.Error(err)
		}
	}

	log.Info("stopping calculation worker pool")
	err = gCalc.Wait()
	if err != nil {
		log.Error(err)
	}

	r, ok := cach.(*cache.Redis)
	if ok {
		err = r.Close()
		if err != nil {
			log.Error(err)
		}
	}

	m, ok := data.(*database.Mongo)
	if ok {
		err = m.Close()
		if err != nil {
			log.Error(err)
		}
	}

	b, ok := data.(*database.Badger)
	if ok {
		err = b.Close()
		if err != nil {
			log.Error(err)
		}
	}
}
