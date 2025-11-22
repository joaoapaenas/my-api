package main

import (
	"database/sql"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	// Import docs without underscore to access docs.SwaggerInfo
	"github.com/joaoapaenas/my-api/docs"
	"github.com/joaoapaenas/my-api/internal/config"
	"github.com/joaoapaenas/my-api/internal/database"
	"github.com/joaoapaenas/my-api/internal/handler"
	"github.com/joaoapaenas/my-api/internal/service"

	// Pure Go SQLite driver (No CGO/GCC required)
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
	// Use "sqlite" as the driver name for glebarez/go-sqlite
	db, err := sql.Open("sqlite", cfg.DBUrl)
	if err != nil {
		log.Fatalf("Failed to open db: %v", err)
	}
	defer db.Close()

	// Verify connection
	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping db: %v", err)
	}

	// 2. Layer Initialization
	// A. Data Access Layer (sqlc)
	queries := database.New(db)

	// B. Service Layer (Business Logic)
	// This wraps the data layer and handles UUIDs, validation, etc.
	userService := service.NewUserManager(queries)

	// C. Handler Layer (HTTP)
	// This depends on the Service, not the raw database queries
	userHandler := handler.NewUserHandler(userService)

	// 3. Router Setup
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	// 4. Documentation Setup
	// A. Explicitly serve the generated JSON to avoid 500 errors
	r.Get("/swagger/doc.json", func(w http.ResponseWriter, r *http.Request) {
		doc := docs.SwaggerInfo.ReadDoc()
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(doc))
	})

	// B. Mount the Swagger UI pointing to the JSON above
	r.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("http://localhost:8080/swagger/doc.json"),
	))

	// 5. Application Routes
	r.Route("/users", func(r chi.Router) {
		r.Post("/", userHandler.CreateUser)
		r.Get("/{email}", userHandler.GetUser)
	})

	// 6. Start Server
	log.Printf("Server starting on :%s", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, r); err != nil {
		log.Fatal(err)
	}
}
