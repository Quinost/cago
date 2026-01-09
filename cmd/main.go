package main

import (
	"cago/internal"
	http_s "cago/internal/http"
	"cago/internal/resp2"
	"context"
	"fmt"
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

	respServer := resp2.NewRESP2Server(cfg, cachesrv, ctx)
	httpServer := http_s.NewHttpServer(cfg, cachesrv, ctx)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		if err := respServer.Run(); err != nil {
			log.Fatal("RESPServer internal error:", err)
		}
	}()

	go func() {
		if err := httpServer.Run(); err != nil {
			log.Fatal("HttpServer internal error:", err)
		}
	}()

	go worker.Run(ctx)
	fmt.Printf("Default TTL: %v, Cleanup interval: %v\n", cfg.DefaultTTL, cfg.CleanupInterval)

	<-sigChan
	cancel()
	respServer.Shutdown()
}
