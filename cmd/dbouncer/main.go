package main

import (
	"context"
	"github.com/hramov/dbouncer/internal"
	v1 "github.com/hramov/dbouncer/internal/adapter/v1"
	"github.com/hramov/dbouncer/internal/config"
	"github.com/hramov/dbouncer/internal/worker"
	"github.com/joho/godotenv"
	"log"
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
	respCh := make(chan *internal.QueryResponse)
	errCh := make(chan error)

	var server *v1.Server
	server, err = v1.NewServer(cfg.Port, cfg.Timeout, queryCh, respCh, errCh)

	for i := 0; i < len(cfg.Storages); i++ {
		var w *worker.Worker
		w, err = worker.NewWorker(i, queryCh, errCh, respCh, cfg.Storages)
		if err != nil {
			log.Fatalf("cannot create worker: %v\n", err)
		}
		go w.Process(ctx)
	}

	go server.Response(ctx)
	go server.Serve(ctx)

	<-ctx.Done()
	cancel()

	os.Exit(0)
}
