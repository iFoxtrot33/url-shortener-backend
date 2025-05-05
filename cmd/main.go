// @title URL Shortener API
// @version 2.0
// @description API for shortening URLs and managing shortened links

// @BasePath /
package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	configs "UrlShortenerBackend/config"

	_ "UrlShortenerBackend/docs"
	"UrlShortenerBackend/internal/link"
	"UrlShortenerBackend/pkg/db"
	"UrlShortenerBackend/pkg/logger"
	"UrlShortenerBackend/pkg/middleware"
	"UrlShortenerBackend/pkg/swagger"

	"github.com/rs/zerolog"
)

func main() {
	//Configs
	cfg := configs.Init()

	// Setting up router
	router := http.NewServeMux()

	//Logger
	log := logger.NewLogger(cfg)
	log.Info().Msg("Application started")
	log.Info().Msg("Environment: " + cfg.Env)

	//Database
	database := db.NewDb(cfg)
	log.Info().Msg("Database connected")

	// Run auto-migration
	log.Info().Msg("Starting auto migration...")
	log.Info().Msg("Running migration for Link model...")
	err := database.AutoMigrate(&link.Link{})
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to run migrations")
	}
	log.Info().Msg("Auto migration completed successfully!")

	// Repositories
	linkRepository := link.NewLinkRepository(database)

	//Services
	linkService := link.NewLinkService(linkRepository, log)
	linkService.Start()
	defer linkService.Stop()

	//Handlers
	link.NewLinkHandler(router, &link.LinkHandlerDeps{
		LinkRepository: linkRepository,
		Config:         cfg,
		Logger:         log,
	})

	// Swagger
	swagger.SetupSwagger(router)

	//Middlewares
	stack := middleware.Chain(
		middleware.Logging(log),
		middleware.CORS(cfg.CORS.AllowedOrigins),
	)

	server := &http.Server{
		Addr:         cfg.HTTPServer.Address,
		Handler:      stack(router),
		ReadTimeout:  cfg.HTTPServer.Timeout,
		WriteTimeout: cfg.HTTPServer.Timeout,
		IdleTimeout:  cfg.HTTPServer.IdleTimeout,
	}

	// Starting server
	go runServer(server, log)

	// Server shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info().Msg("Shutting down application...")

	// Completing remaining requests
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Stopping server
	if err := server.Shutdown(ctx); err != nil {
		log.Error().Err(err).Msg("Server forced to shutdown")
	}

	// Closing database connection
	sqlDB, _ := database.DB.DB()
	sqlDB.Close()

	log.Info().Msg("Application stopped successfully")
}

func runServer(server *http.Server, log *zerolog.Logger) {
	log.Info().Str("address", server.Addr).Msg("Starting HTTP server")
	err := server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		log.Fatal().Err(err).Msg("Could not start HTTP server")
	}
}
