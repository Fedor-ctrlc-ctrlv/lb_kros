package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"web-api/back/api"
	"web-api/back/api/handlers"
	"web-api/back/api/middleware"
	"web-api/back/database"
	"web-api/back/repository"
	"web-api/back/services"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println(".env not loaded; using environment variables")
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if len(jwtSecret) < 32 {
		log.Fatal("JWT_SECRET must contain at least 32 characters")
	}

	ctx := context.Background()

	db, err := database.Connect(ctx)
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
	authService := services.NewAuthService(authRepo, jwtSecret)
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
		Addr:              "127.0.0.1:8080",
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      60 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	stopCtx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	go func() {
		log.Println("HTTP server listening on http://127.0.0.1:8080")

		if err := server.ListenAndServe(); err != nil &&
			!errors.Is(err, http.ErrServerClosed) {
			log.Printf("HTTP server failed: %v", err)
			stop()
		}
	}()

	<-stopCtx.Done()

	shutdownCtx, cancel := context.WithTimeout(
		context.Background(),
		10*time.Second,
	)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("server shutdown failed: %v", err)
	}

	log.Println("HTTP server stopped")
}
