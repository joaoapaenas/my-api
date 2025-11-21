package main

import (
	"database/sql"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/joaoapaenas/my-api/internal/config"
	"github.com/joaoapaenas/my-api/internal/database"
	"github.com/joaoapaenas/my-api/internal/handler"

	// USE THIS PURE GO DRIVER:
	_ "github.com/glebarez/go-sqlite"

	httpSwagger "github.com/swaggo/http-swagger/v2"
)

// @title My Go API
// @version 1.0
// @description Production ready starter guide.
// @host localhost:8080
// @BasePath /
func main() {
	cfg := config.Load()

	// 1. Database Connection
	// IMPORTANT: Driver name is "sqlite", not "sqlite3"
	db, err := sql.Open("sqlite", cfg.DBUrl)
	if err != nil {
		log.Fatalf("Failed to open db: %v", err)
	}
	defer db.Close()

	// Basic check
	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping db: %v", err)
	}

	queries := database.New(db)
	userHandler := handler.NewUserHandler(queries)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	r.Get("/swagger/*", httpSwagger.WrapHandler)

	r.Route("/users", func(r chi.Router) {
		r.Post("/", userHandler.CreateUser)
		r.Get("/{email}", userHandler.GetUser)
	})

	log.Printf("Server starting on :%s", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, r); err != nil {
		log.Fatal(err)
	}
}
