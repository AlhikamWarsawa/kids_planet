package handlers

import (
	"database/sql"

	"github.com/gofiber/fiber/v2"

	"github.com/ZygmaCore/kids_planet/services/api/internal/clients"
	"github.com/ZygmaCore/kids_planet/services/api/internal/config"
	admin "github.com/ZygmaCore/kids_planet/services/api/internal/handlers/admin"
	public "github.com/ZygmaCore/kids_planet/services/api/internal/handlers/public"
	"github.com/ZygmaCore/kids_planet/services/api/internal/middleware"
	"github.com/ZygmaCore/kids_planet/services/api/internal/repos"
	"github.com/ZygmaCore/kids_planet/services/api/internal/services"
)

type Deps struct {
	Cfg    config.Config
	DB     *sql.DB
	Valkey *clients.Valkey
}

func Register(app *fiber.App, deps Deps) {
	api := app.Group("/api")

	healthHandler := NewHealthHandler(deps.Cfg)
	api.Get("/health", healthHandler.Get)

	gameRepo := repos.NewGameRepo(deps.DB)
	userRepo := repos.NewUserRepo(deps.DB)
	submissionRepo := repos.NewSubmissionRepo(deps.DB)

	gameSvc := services.NewGameService(gameRepo)
	sessionSvc := services.NewSessionService(deps.Cfg, gameRepo)
	leaderboardSvc := services.NewLeaderboardService(deps.Valkey, submissionRepo)

	gamesHandler := public.NewGamesHandler(gameSvc)
	api.Get("/games", gamesHandler.List)
	api.Get("/games/:id", gamesHandler.Get)

	sessionsHandler := public.NewSessionsHandler(sessionSvc)
	api.Post("/sessions/start", sessionsHandler.Start)

	leaderboardHandler := public.NewLeaderboardHandler(leaderboardSvc)
	api.Get("/leaderboard/:game_id<int>", leaderboardHandler.GetTop)

	api.Post(
		"/leaderboard/submit",
		middleware.PlayToken(deps.Cfg),
		leaderboardHandler.Submit,
	)

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

	adminGames := admin.NewGamesHandler(gameSvc)
	adminGroup.Get("/games", adminGames.List)
	adminGroup.Post("/games", adminGames.Create)
	adminGroup.Put("/games/:id<int>", adminGames.Update)
	adminGroup.Post("/games/:id<int>/publish", adminGames.Publish)
	adminGroup.Post("/games/:id<int>/unpublish", adminGames.Unpublish)
}
