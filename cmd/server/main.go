package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"polling-service/internal/config"
	"polling-service/internal/handler"
	"polling-service/internal/middleware"
	"polling-service/internal/service"
	"polling-service/internal/storage/postgres"
	redisstorage "polling-service/internal/storage/redis"
	resultsync "polling-service/internal/sync"
)

func main() {
	ctx := context.Background()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	pgPool, err := postgres.New(ctx, cfg.PostgresDSN)
	if err != nil {
		log.Fatalf("failed to connect to postgres: %v", err)
	}
	defer pgPool.Close()

	redisClient, err := redisstorage.New(
		ctx,
		cfg.RedisAddr,
		cfg.RedisDB,
	)
	if err != nil {
		log.Fatalf("failed to connect to redis: %v", err)
	}
	defer func() {
		if err := redisClient.Close(); err != nil {
			log.Printf("redis close failed: %v", err)
		}
	}()

	pollRepository := postgres.NewPollRepository(pgPool)
	pollCache := redisstorage.NewPollStorage(redisClient)

	voteStorage := redisstorage.NewVoteStorage(redisClient)
	resultRepository := postgres.NewResultRepository(pgPool)

	pollService := service.NewPollService(
		pollRepository,
		pollCache,
	)

	voteService := service.NewVoteService(
		pollRepository,
		pollCache,
		voteStorage,
	)

	resultService := service.NewResultService(
		pollRepository,
		voteStorage,
	)

	resultSync := resultsync.NewResultSync(
		pollRepository,
		voteStorage,
		resultRepository,
		5*time.Second,
	)

	publicHandler := handler.NewPublicHandler(
		pollService,
		voteService,
	)

	adminHandler := handler.NewAdminHandler(
		pollService,
		resultService,
	)

	rateLimiter := middleware.NewRateLimiter(
		cfg.RateLimitRPS,
		time.Second,
	)

	router := handler.NewRouter(
		publicHandler,
		adminHandler,
		cfg.AdminToken,
		rateLimiter,
	)

	server := &http.Server{
		Addr:    ":" + cfg.HTTPPort,
		Handler: router,
	}

	go func() {
		log.Printf("server started on :%s", cfg.HTTPPort)

		if err := server.ListenAndServe(); err != nil &&
			err != http.ErrServerClosed {
			log.Fatalf("server failed: %v", err)
		}
	}()

	go resultSync.Start(ctx)

	waitForShutdown(server)
}

func waitForShutdown(server *http.Server) {
	stop := make(chan os.Signal, 1)

	signal.Notify(
		stop,
		syscall.SIGINT,
		syscall.SIGTERM,
	)

	<-stop

	log.Println("shutting down server...")

	ctx, cancel := context.WithTimeout(
		context.Background(),
		5*time.Second,
	)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("server shutdown failed: %v", err)
		return
	}

	log.Println("server stopped")
}
