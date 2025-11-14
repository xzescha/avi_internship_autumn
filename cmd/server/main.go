package main

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/lib/pq"

	"avi_internship_autumn/internal/app"
	"avi_internship_autumn/internal/config"
	apihttp "avi_internship_autumn/internal/http"
	"avi_internship_autumn/internal/service"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}
	dsn := cfg.DB.DSNString()

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			log.Fatalf("failed to close database: %v", err)
		}
	}(db)

	db.SetMaxOpenConns(cfg.DB.MaxOpenConns)
	db.SetMaxIdleConns(cfg.DB.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.DB.ConnMaxLifetime)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	if err := db.PingContext(ctx); err != nil {
		cancel()
		log.Fatalf("failed to ping database: %v", err)
	}
	cancel()

	// TODO: run migrations here

	repos := app.NewRepositories(db)

	teamSvc := service.NewTeamService(repos.Teams, repos.Users)
	userSvc := service.NewUserService(repos.Users, repos.PRs)
	prSvc := service.NewPRService(repos.PRs, repos.Users)

	handler := apihttp.NewRouter(teamSvc, userSvc, prSvc)
	application := app.NewApp(handler, teamSvc, userSvc, prSvc)

	srv := &http.Server{
		Addr:         ":" + cfg.HTTP.Port,
		Handler:      application.Handler,
		ReadTimeout:  cfg.HTTP.ReadTimeout,
		WriteTimeout: cfg.HTTP.WriteTimeout,
		IdleTimeout:  cfg.HTTP.IdleTimeout,
	}

	go func() {
		log.Printf("HTTP server listening on :%s", cfg.HTTP.Port)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("server error: %v", err)
		}
	}()

	// ---- graceful shutdown ----
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	<-stop
	log.Println("shutting down...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("graceful shutdown failed: %v", err)
		if err := srv.Close(); err != nil {
			log.Printf("server close failed: %v", err)
		}
	}

	log.Println("server stopped")
}
