package main

import (
	"context"
	"github.com/hramov/dbouncer/internal"
	v1 "github.com/hramov/dbouncer/internal/adapter/v1"
	"github.com/hramov/dbouncer/internal/config"
	"github.com/hramov/dbouncer/internal/worker"
	"github.com/hramov/dbouncer/pkg/metrics"
	"github.com/hramov/dbouncer/pkg/storage"
	"github.com/joho/godotenv"
	"log"
	_ "net/http/pprof"
	"os"
	"os/signal"
)

func main() {
	if os.Getenv("ENV") == "" {
		err := godotenv.Load()
		if err != nil {
			log.Fatal("error loading .env file")
		}
	}

	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		log.Fatal("config path env is not set")
	}

	cfg := config.Config{}
	err := config.LoadConfig(configPath, &cfg)
	if err != nil {
		log.Fatalf("cannot parse config file: %v\n", err)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)

	queryCh := make(chan *internal.QueryRequest)
	errCh := make(chan error)
	respCh := make(chan *internal.QueryResponse)
	statCh := make(chan map[string]int)

	go func() {
		var ok bool
		for {
			select {
			case <-ctx.Done():
				return
			case err, ok = <-errCh:
				if !ok {
					return
				}
				log.Printf("error: %v\n", err)
			}
		}
	}()

	var server *v1.Server
	server, err = v1.NewServer(cfg.Port, cfg.Timeout, queryCh, respCh, errCh)

	for _, st := range cfg.Storages {
		var db internal.Storage
		db, err = storage.New(st.Name, st.Dsn, st.PoolMax, st.IdleMax, st.IdleTimeout, st.LifeTime)
		for i := 0; i < st.Workers; i++ {
			var w *worker.Worker
			w, err = worker.NewWorker(i, queryCh, errCh, respCh, statCh, st.Name, db)
			if err != nil {
				log.Fatalf("cannot create worker: %v\n", err)
			}
			go w.Process(ctx)
		}
	}

	go server.Response(ctx)
	go server.Serve(ctx)

	go metrics.Collect(ctx, cfg.MetricsPort, statCh)

	<-ctx.Done()
	cancel()

	os.Exit(0)
}
