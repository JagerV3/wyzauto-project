package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/raymond/wyzauto-project/internal/config"
	"github.com/raymond/wyzauto-project/internal/handler"
	"github.com/raymond/wyzauto-project/internal/repository"
	"github.com/raymond/wyzauto-project/internal/service"
)

func main() {
	cfg := config.Load()
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	pool, err := pgxpool.New(ctx, cfg.DBURL)
	if err != nil {
		slog.Error("connect database", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	productRepo := repository.NewProductPostgresRepository(pool)
	translationRepo := repository.NewTranslationPostgresRepository(pool)
	translationLoader := service.NewTranslationLoader(translationRepo, cfg.CacheTTL)
	builder := service.NewProductDocumentBuilder(productRepo, translationLoader, []string{"en", "th"})

	mux := http.NewServeMux()
	handler.NewProductHandler(builder).Register(mux)

	server := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		slog.Info("server started", "addr", cfg.HTTPAddr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server failed", "error", err)
			stop()
		}
	}()

	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		slog.Error("shutdown server", "error", err)
	}
}
