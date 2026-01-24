package handlers

import (
	"database/sql"

	"github.com/gofiber/fiber/v2"

	"github.com/ZygmaCore/kids_planet/services/api/internal/config"
	admin "github.com/ZygmaCore/kids_planet/services/api/internal/handlers/admin"
	public "github.com/ZygmaCore/kids_planet/services/api/internal/handlers/public"
	"github.com/ZygmaCore/kids_planet/services/api/internal/middleware"
	"github.com/ZygmaCore/kids_planet/services/api/internal/repos"
	"github.com/ZygmaCore/kids_planet/services/api/internal/services"
)

type Deps struct {
	Cfg config.Config
	DB  *sql.DB
}

func Register(app *fiber.App, deps Deps) {
	api := app.Group("/api")

	healthHandler := NewHealthHandler(deps.Cfg)
	api.Get("/health", healthHandler.Get)

	gameRepo := repos.NewGameRepo(deps.DB)
	gameSvc := services.NewGameService(gameRepo)
	gamesHandler := public.NewGamesHandler(gameSvc)
	api.Get("/games", gamesHandler.List)
	api.Get("/games/:id", gamesHandler.Get)

	sessionSvc := services.NewSessionService(deps.Cfg, gameRepo)
	sessionsHandler := public.NewSessionsHandler(sessionSvc)
	api.Post("/sessions/start", sessionsHandler.Start)

	userRepo := repos.NewUserRepo(deps.DB)

	authHandler := admin.NewAuthHandler(deps.Cfg, userRepo)
	api.Post("/auth/admin/login", authHandler.Login)

	adminGroup := api.Group(
		"/admin",
		middleware.AuthJWT(deps.Cfg),
		middleware.RequireAdmin(),
	)

	adminPing := admin.NewPingHandler()
	adminGroup.Get("/ping", adminPing.Get)

	adminMe := admin.NewMeHandler(userRepo)
	adminGroup.Get("/me", adminMe.Get)
}
