package handlers

import (
	"database/sql"

	"github.com/gofiber/fiber/v2"

	"github.com/ZygmaCore/kids_planet/services/api/internal/config"
)

type Deps struct {
	Cfg config.Config
	DB  *sql.DB
}

func Register(app *fiber.App, deps Deps) {
	healthHandler := NewHealthHandler(deps.Cfg)

	app.Get("/health", healthHandler.Get)

	// publicGroup := app.Group("/public")
	// app.Get("/games", ...)
}
