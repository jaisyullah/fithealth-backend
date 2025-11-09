package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/jaisyullah/fithealth-backend/internal/config"
	"github.com/jaisyullah/fithealth-backend/internal/logger"
	"github.com/jaisyullah/fithealth-backend/internal/oauth"
	"github.com/jaisyullah/fithealth-backend/internal/server"
	"github.com/jaisyullah/fithealth-backend/internal/store"
	"github.com/jaisyullah/fithealth-backend/internal/worker"

	"github.com/redis/go-redis/v9"
)

func main() {
	cfg := config.LoadFromEnv()
	logger.Init(cfg.LogLevel)

	db, err := store.NewGorm(cfg.DatabaseDSN)
	if err != nil {
		log.Fatalf("failed connect db: %v", err)
	}

	redisClient := redis.NewClient(&redis.Options{
		Addr: cfg.RedisAddr,
	})

	tokenMgr := oauth.NewTokenManager(cfg.SatusehatTokenURL, cfg.SatusehatClientID, cfg.SatusehatClientSecret, 10*time.Second)

	// Start worker
	w := worker.NewSenderWorker(db, redisClient, tokenMgr, cfg)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go w.Start(ctx)

	// Start HTTP server
	e := server.NewServer(db, redisClient, tokenMgr, cfg)

	// graceful shutdown
	go func() {
		if err := e.Start(":" + cfg.Port); err != nil {
			logger.Log.Fatalf("server stopped: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	logger.Log.Printf("shutting down server...")
	ctxShut, cancel2 := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel2()
	if err := e.Shutdown(ctxShut); err != nil {
		logger.Log.Fatalf("shutdown failed: %v", err)
	}
}
