package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/ZygmaCore/kids_planet/services/api/internal/clients"
	"github.com/ZygmaCore/kids_planet/services/api/internal/config"
	"github.com/ZygmaCore/kids_planet/services/api/internal/handlers"
	"github.com/ZygmaCore/kids_planet/services/api/internal/middleware"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg := config.MustLoad()

	db, err := clients.NewPostgres(ctx, cfg.Postgres)
	if err != nil {
		log.Fatalf("startup failed (postgres): %v", err)
	}
	defer func() { _ = db.Close() }()
	log.Println("postgres connected")

	vk, err := clients.NewValkey(cfg.Valkey)
	if err != nil {
		log.Fatalf("startup failed (valkey): %v", err)
	}
	defer func() { _ = vk.Close() }()
	log.Println("valkey connected")

	mo, err := clients.NewMinIO(ctx, cfg.MinIO)
	if err != nil {
		log.Fatalf("startup failed (minio): %v", err)
	}
	log.Println("minio connected")

	app := fiber.New(fiber.Config{
		AppName:      "game-portal-api",
		ReadTimeout:  20 * time.Second,
		WriteTimeout: 20 * time.Second,
		BodyLimit:    middleware.MaxRequestBodyBytes,
	})

	app.Use(middleware.RequestID())
	app.Use(middleware.Recover())
	app.Use(middleware.Logging())
	app.Use(middleware.CORS())
	app.Use(middleware.SizeLimit())

	handlers.Register(app, handlers.Deps{
		Cfg:    cfg,
		DB:     db,
		Valkey: vk,
		MinIO:  mo,
	})

	if cfg.Env != "prod" {
		app.Get("/api/panic", func(c *fiber.Ctx) error { panic("test") })
	}

	addr := "0.0.0.0:" + cfg.Port
	log.Printf("API listening on %s", addr)

	go func() {
		if err := app.Listen(addr); err != nil {
			log.Printf("server stopped: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("shutdown signal received")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := app.ShutdownWithContext(shutdownCtx); err != nil {
		log.Printf("shutdown error: %v", err)
	}

	log.Println("shutdown complete")
}
