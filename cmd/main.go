package main

import (
	"cago/internal"
	"cago/internal/resp2"
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())

	cfg := internal.LoadConfig()
	storage := internal.NewStorage()
	cachesrv := internal.NewCacheService(storage, cfg.DefaultTTL)
	worker := internal.NewCleanupWorker(cfg, storage)

	//server := simple.NewSimpleProtocolServer(cfg, cachesrv, ctx)
	server := resp2.NewRESP2Server(cfg, cachesrv, ctx)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		if err := server.Run(); err != nil {
			log.Fatal("Server internal error:", err)
		}
	}()

	go worker.Run(ctx)

	<-sigChan
	cancel()
	server.Shutdown()
}
