package main

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"

	"github.com/neves144/goledger-challenge-besu/app/internal/config"
	"github.com/neves144/goledger-challenge-besu/app/internal/simplestorage"
)

func run(cfg *config.Config) error {
	db, err := openDB(cfg)
	if err != nil {
		return fmt.Errorf("connect to database: %w", err)
	}
	defer db.Close()

	besu, err := simplestorage.NewBesuAdapter(cfg.BesuRPCURL, cfg.ContractAddress, cfg.PrivateKey)
	if err != nil {
		return fmt.Errorf("connect to besu: %w", err)
	}
	defer besu.Close()

	repo := simplestorage.NewPostgresRepo(db)
	svc := simplestorage.NewService(besu, besu, repo, cfg.ContractAddress)
	h := simplestorage.NewHandler(svc)

	srv := &http.Server{
		Addr:    cfg.ServerAddr,
		Handler: buildRouter(h),
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		slog.Info("server starting", "addr", cfg.ServerAddr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("listener error", "err", err)
			quit <- syscall.SIGTERM
		}
	}()

	<-quit
	slog.Info("shutting down")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	return srv.Shutdown(ctx)
}

func buildRouter(h *simplestorage.Handler) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(requestLogger)
	r.Use(middleware.Recoverer)

	r.Post("/api/storage", h.SetHandler)
	r.Get("/api/storage", h.GetHandler)
	r.Post("/api/storage/sync", h.SyncHandler)
	r.Get("/api/storage/check", h.CheckHandler)

	return r
}

func requestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
		start := time.Now()
		next.ServeHTTP(ww, r)
		slog.Info("request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", ww.Status(),
			"duration_ms", time.Since(start).Milliseconds(),
			"request_id", middleware.GetReqID(r.Context()),
		)
	})
}

func openDB(cfg *config.Config) (*sql.DB, error) {
	connConfig, err := pgx.ParseConfig("")
	if err != nil {
		return nil, err
	}
	connConfig.Host = cfg.DBHost
	connConfig.Port = cfg.DBPort
	connConfig.User = cfg.DBUser
	connConfig.Password = cfg.DBPass
	connConfig.Database = cfg.DBName

	db := stdlib.OpenDB(*connConfig)
	if err := db.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}
