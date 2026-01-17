package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"

	"github.com/ZygmaCore/kids_planet/services/api/internal/clients"
	"github.com/ZygmaCore/kids_planet/services/api/internal/repos"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	db, err := clients.NewPostgres(ctx)
	if err != nil {
		log.Fatalf("startup failed (postgres): %v", err)
	}
	defer func() { _ = db.Close() }()

	log.Println("postgres connected")

	userRepo := repos.NewUserRepo(db)
	gameRepo := repos.NewGameRepo(db)
	_ = userRepo
	_ = gameRepo

	app := fiber.New(fiber.Config{
		AppName:      "game-portal-api",
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	})

	app.Use(recover.New())
	app.Use(logger.New())

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"ok":      true,
			"service": "api",
			"time":    time.Now().UTC().Format(time.RFC3339),
		})
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	addr := "0.0.0.0:" + port
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
