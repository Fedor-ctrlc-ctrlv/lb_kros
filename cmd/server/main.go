package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"web-api/internal/api"
	"web-api/internal/api/handlers"
	"web-api/internal/api/middleware"
	"web-api/internal/config"
	"web-api/internal/database"
	"web-api/internal/repository"
	"web-api/internal/services"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("configuration error: ", err)
	}

	ctx := context.Background()

	db, err := database.Connect(ctx, cfg.Database)
	if err != nil {
		log.Fatal("database connection failed: ", err)
	}
	defer db.Close()

	// Hosts
	hostRepo := repository.NewHostRepository(db)
	hostService := services.NewHostService(hostRepo)
	hostHandler := handlers.NewHostHandler(hostService)

	// Security policies
	policyRepo := repository.NewPolicyRepository(db)
	policyService := services.NewPolicyService(policyRepo)
	policyHandler := handlers.NewPolicyHandler(policyService)

	// Audits
	auditRepo := repository.NewAuditRepository(db)
	auditService := services.NewAuditService(auditRepo)
	auditHandler := handlers.NewAuditHandler(auditService)

	// Baselines
	baselineRepo := repository.NewBaselineRepository(db)
	baselineService := services.NewBaselineService(baselineRepo)
	baselineHandler := handlers.NewBaselineHandler(baselineService)

	// Reports
	reportRepo := repository.NewReportRepository(db)
	reportService := services.NewReportService(reportRepo)
	reportHandler := handlers.NewReportHandler(reportService)

	// JWT authentication
	authRepo := repository.NewAuthRepository(db)
	authService := services.NewAuthService(
		authRepo,
		cfg.JWT.Secret,
	)
	authHandler := handlers.NewAuthHandler(authService)
	authMiddleware := middleware.NewAuthMiddleware(authService)

	router := api.NewRouter(
		hostHandler,
		policyHandler,
		auditHandler,
		baselineHandler,
		reportHandler,
		authHandler,
		authMiddleware,
	)

	server := &http.Server{
		Addr:              cfg.HTTPAddress(),
		Handler:           router,
		ReadHeaderTimeout: cfg.HTTP.ReadHeaderTimeout,
		ReadTimeout:       cfg.HTTP.ReadTimeout,
		WriteTimeout:      cfg.HTTP.WriteTimeout,
		IdleTimeout:       cfg.HTTP.IdleTimeout,
	}

	stopCtx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	go func() {
		log.Printf(
			"HTTP server listening on http://%s",
			cfg.HTTPAddress(),
		)

		if err := server.ListenAndServe(); err != nil &&
			!errors.Is(err, http.ErrServerClosed) {
			log.Printf("HTTP server failed: %v", err)
			stop()
		}
	}()

	<-stopCtx.Done()

	shutdownCtx, cancel := context.WithTimeout(
		context.Background(),
		cfg.HTTP.ShutdownTimeout,
	)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("server shutdown failed: %v", err)
	}

	log.Println("HTTP server stopped")
}
