package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"wish1/config"
	"wish1/internal/api"
	"wish1/internal/consumer"
	"wish1/internal/stoplist"
	"wish1/internal/window"
)

func main() {
	cfg := config.Load()

	win := window.New(cfg.BucketCount, cfg.BucketDuration)
	sl := stoplist.New()
	
	go func() {
		ticker := time.NewTicker(cfg.CacheRefresh)
		defer ticker.Stop()
		for t := range ticker.C {
			win.RebuildCache(t)
		}
	}()

	ctx, cancel := context.WithCancel(context.Background())

	cons := consumer.New(cfg.KafkaBrokers, cfg.KafkaTopic, cfg.KafkaGroupID, win, sl)
	go func() {
		cons.Run(ctx)
	}()

	handler := api.NewHandler(win, sl)
	srv := api.NewServer(cfg.HTTPAddr, handler)

	go func() {
		log.Printf("listening on %s", cfg.HTTPAddr)
		if err := srv.ListenAndServe(); err != nil {
			log.Printf("server stopped: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("shutting down...")
	cancel()

	shutCtx, shutCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutCancel()
	srv.Shutdown(shutCtx)
	cons.Close()
}
