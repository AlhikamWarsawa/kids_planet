package handlers

import (
	"database/sql"

	"github.com/gofiber/fiber/v2"

	"github.com/ZygmaCore/kids_planet/services/api/internal/config"
	admin "github.com/ZygmaCore/kids_planet/services/api/internal/handlers/admin"
	"github.com/ZygmaCore/kids_planet/services/api/internal/middleware"
	"github.com/ZygmaCore/kids_planet/services/api/internal/repos"
)

type Deps struct {
	Cfg config.Config
	DB  *sql.DB
}

func Register(app *fiber.App, deps Deps) {
	api := app.Group("/api")

	healthHandler := NewHealthHandler(deps.Cfg)
	api.Get("/health", healthHandler.Get)

	userRepo := repos.NewUserRepo(deps.DB)

	authHandler := admin.NewAuthHandler(deps.Cfg, userRepo)
	api.Post("/auth/admin/login", authHandler.Login)

	adminPing := admin.NewPingHandler()
	adminGroup := api.Group("/admin", middleware.AuthJWT(deps.Cfg), middleware.RequireAdmin())
	adminGroup.Get("/ping", adminPing.Get)
}
