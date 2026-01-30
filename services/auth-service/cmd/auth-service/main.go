package main

import (
	"context"
	"io"
	"log"
	"os"
	"path/filepath"

	"html2pdf-auth-service/internal/tokens"
	"html2pdf-auth-service/internal/config"
	"html2pdf-auth-service/internal/infra/postgres"
	"html2pdf-auth-service/internal/http/server"
)

func main() {
	logFile := setupLogging()
	if logFile != nil {
		defer logFile.Close()
	}

	cfg := config.Load()

	// Token cache + repository
	cache := tokens.NewCache()
	db := postgres.NewDB()
	repo := postgres.NewTokenRepository(db, cfg.PostgresDSN)

	// Initial token load (may fail if DB not ready yet).
	if err := tokens.NewReloader(repo, cache, cfg.TokenReloadInterval).LoadOnce(context.Background()); err != nil {
		log.Printf("initial token load failed: %v", err)
	} else {
		log.Printf("token store ready")
	}

	// Periodic reload.
	reloader := tokens.NewReloader(repo, cache, cfg.TokenReloadInterval)
	reloader.Start(context.Background())

	store := server.NewRateLimitStore(cfg)

	if err := server.Run(context.Background(), server.Deps{
		Config:     cfg,
		TokenCache: cache,
		Store:      store,
	}); err != nil {
		log.Fatal(err)
	}
}

func setupLogging() *os.File {
	logDir := "/logs"
	if err := os.MkdirAll(logDir, 0o755); err != nil {
		log.Printf("failed to create log dir %s: %v", logDir, err)
		return nil
	}
	path := filepath.Join(logDir, "auth-service.log")
	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		log.Printf("failed to open log file %s: %v", path, err)
		return nil
	}
	log.SetOutput(io.MultiWriter(os.Stdout, file))
	log.SetFlags(log.LstdFlags | log.LUTC)
	log.Printf("logging to %s", path)
	return file
}
