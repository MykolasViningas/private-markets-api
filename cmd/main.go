package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"

	"private-markets-api/internal/handler"
	"private-markets-api/internal/repository"
	"private-markets-api/internal/server"
	"private-markets-api/internal/service"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		logger.Error("DATABASE_URL environment variable is required")
		os.Exit(1)
	}

	pool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		logger.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	if err := pool.Ping(context.Background()); err != nil {
		logger.Error("failed to ping database", "error", err)
		os.Exit(1)
	}

	// Initialize repository
	repo := repository.New(pool)

	// Initialize services
	fundService := service.NewFundService(repo)
	investorService := service.NewInvestorService(repo)
	investmentService := service.NewInvestmentService(repo)

	// Initialize handler
	h := handler.NewHandler(fundService, investorService, investmentService, logger)

	router := server.NewRouter(h, logger)

	srv := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	logger.Info("server started", "addr", ":8080")

	log.Fatal(srv.ListenAndServe())
}
