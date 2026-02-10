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
	MinIO  *clients.MinIO
}

func Register(app *fiber.App, deps Deps) {
	api := app.Group("/api")

	healthHandler := NewHealthHandler(deps.Cfg)
	api.Get("/health", healthHandler.Get)

	gameRepo := repos.NewGameRepo(deps.DB)
	userRepo := repos.NewUserRepo(deps.DB)
	submissionRepo := repos.NewSubmissionRepo(deps.DB)
	analyticsRepo := repos.NewAnalyticsRepo(deps.DB)
	dashboardRepo := repos.NewDashboardRepo(deps.DB)
	sessionRepo := repos.NewSessionRepo(deps.DB)

	ageCategoryRepo := repos.NewAgeCategoryRepo(deps.DB)
	educationCategoryRepo := repos.NewEducationCategoryRepo(deps.DB)

	gameSvc := services.NewGameService(
		gameRepo,
		deps.MinIO,
		deps.Cfg.MinIO.Bucket,
		deps.Cfg.Upload.ZipMaxBytes,
	)

	sessionSvc := services.NewSessionService(deps.Cfg, gameRepo, sessionRepo)
	leaderboardSvc := services.NewLeaderboardService(deps.Valkey, submissionRepo)

	categorySvc := services.NewCategoryService(ageCategoryRepo, educationCategoryRepo)
	dashboardSvc := services.NewDashboardService(dashboardRepo)

	gamesHandler := public.NewGamesHandler(gameSvc)
	api.Get("/games", gamesHandler.List)
	api.Get("/games/:id", gamesHandler.Get)

	sessionsHandler := public.NewSessionsHandler(sessionSvc)
	api.Post("/sessions/start", sessionsHandler.Start)

	analyticsHandler := public.NewAnalyticsHandler(deps.Cfg, analyticsRepo)
	api.Post("/analytics/event", analyticsHandler.TrackEvent)

	leaderboardHandler := public.NewLeaderboardHandler(leaderboardSvc)
	api.Get("/leaderboard/:game_id<int>", leaderboardHandler.GetTop)
	api.Post(
		"/leaderboard/submit",
		middleware.PlayToken(deps.Cfg),
		middleware.RateLimitLeaderboardSubmit(deps.Valkey),
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

	adminDashboard := admin.NewDashboardHandler(dashboardSvc)
	adminGroup.Get("/dashboard/overview", adminDashboard.Overview)

	adminGames := admin.NewGamesHandler(gameSvc)
	adminGroup.Get("/games", adminGames.List)
	adminGroup.Post("/games", adminGames.Create)
	adminGroup.Put("/games/:id<int>", adminGames.Update)
	adminGroup.Post("/games/:id<int>/publish", adminGames.Publish)
	adminGroup.Post("/games/:id<int>/unpublish", adminGames.Unpublish)
	adminGroup.Post("/games/:id<int>/upload", adminGames.Upload)

	adminCategories := admin.NewCategoriesHandler(categorySvc)

	adminGroup.Get("/age-categories", adminCategories.ListAge)
	adminGroup.Post("/age-categories", adminCategories.CreateAge)
	adminGroup.Put("/age-categories/:id<int>", adminCategories.UpdateAge)
	adminGroup.Delete("/age-categories/:id<int>", adminCategories.DeleteAge)

	adminGroup.Get("/education-categories", adminCategories.ListEducation)
	adminGroup.Post("/education-categories", adminCategories.CreateEducation)
	adminGroup.Put("/education-categories/:id<int>", adminCategories.UpdateEducation)
	adminGroup.Delete("/education-categories/:id<int>", adminCategories.DeleteEducation)
}
