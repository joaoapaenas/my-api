package main

import (
	"context"
	"database/sql"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/joaoapaenas/my-api/docs"
	"github.com/joaoapaenas/my-api/internal/config"
	"github.com/joaoapaenas/my-api/internal/database"
	"github.com/joaoapaenas/my-api/internal/handler"
	"github.com/joaoapaenas/my-api/internal/logger"
	"github.com/joaoapaenas/my-api/internal/repository"
	"github.com/joaoapaenas/my-api/internal/service"

	_ "github.com/glebarez/go-sqlite"
	httpSwagger "github.com/swaggo/http-swagger/v2"
)

// @title My Go API
// @version 1.0
// @description Production ready starter guide.
// @host localhost:8080
// @BasePath /
func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	logger.Init(cfg.Env)
	slog.Info("Starting application", "env", cfg.Env, "port", cfg.Port)

	// 1. Open Database
	// using the Safe Mode URL from config
	db, err := sql.Open("sqlite", cfg.DBUrl)
	if err != nil {
		slog.Error("Failed to initialize db driver", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	// 2. Verify Connection
	// This triggers the actual file open and pragma application
	if err := db.Ping(); err != nil {
		slog.Error("Failed to connect to database", "error", err)
		os.Exit(1)
	}
	slog.Info("Database connected", "url", cfg.DBUrl)

	// 3. Wiring Layers
	queries := database.New(db)

	// Repositories
	userRepo := repository.NewSQLUserRepository(queries)
	subjectRepo := repository.NewSQLSubjectRepository(queries)
	topicRepo := repository.NewSQLTopicRepository(queries)
	studyCycleRepo := repository.NewSQLStudyCycleRepository(queries)
	cycleItemRepo := repository.NewSQLCycleItemRepository(queries)
	studySessionRepo := repository.NewSQLStudySessionRepository(queries)
	sessionPauseRepo := repository.NewSQLSessionPauseRepository(queries)
	exerciseLogRepo := repository.NewSQLExerciseLogRepository(queries)

	// Services
	userService := service.NewUserManager(userRepo)
	subjectService := service.NewSubjectManager(subjectRepo)
	topicService := service.NewTopicManager(topicRepo)
	studyCycleService := service.NewStudyCycleManager(studyCycleRepo)
	cycleItemService := service.NewCycleItemManager(cycleItemRepo)
	studySessionService := service.NewStudySessionManager(studySessionRepo)
	sessionPauseService := service.NewSessionPauseManager(sessionPauseRepo)
	exerciseLogService := service.NewExerciseLogManager(exerciseLogRepo)

	// Handlers
	userHandler := handler.NewUserHandler(userService)
	subjectHandler := handler.NewSubjectHandler(subjectService)
	topicHandler := handler.NewTopicHandler(topicService)
	studyCycleHandler := handler.NewStudyCycleHandler(studyCycleService)
	cycleItemHandler := handler.NewCycleItemHandler(cycleItemService)
	studySessionHandler := handler.NewStudySessionHandler(studySessionService)
	sessionPauseHandler := handler.NewSessionPauseHandler(sessionPauseService)
	exerciseLogHandler := handler.NewExerciseLogHandler(exerciseLogService)

	// 4. Router Setup
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	r.Get("/swagger/doc.json", func(w http.ResponseWriter, r *http.Request) {
		doc := docs.SwaggerInfo.ReadDoc()
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(doc))
	})
	r.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("http://localhost:8080/swagger/doc.json"),
	))

	r.Route("/users", func(r chi.Router) {
		r.Post("/", userHandler.CreateUser)
		r.Get("/{email}", userHandler.GetUser)
	})

	r.Route("/subjects", func(r chi.Router) {
		r.Post("/", subjectHandler.CreateSubject)
		r.Get("/", subjectHandler.ListSubjects)
		r.Post("/{id}/topics", topicHandler.CreateTopic)
		r.Get("/{id}/topics", topicHandler.ListTopics)
	})

	r.Route("/study-cycles", func(r chi.Router) {
		r.Post("/", studyCycleHandler.CreateStudyCycle)
		r.Get("/active", studyCycleHandler.GetActiveStudyCycle)
		r.Post("/{id}/items", cycleItemHandler.CreateCycleItem)
		r.Get("/{id}/items", cycleItemHandler.ListCycleItems)
	})

	r.Route("/study-sessions", func(r chi.Router) {
		r.Post("/", studySessionHandler.CreateStudySession)
		r.Put("/{id}", studySessionHandler.UpdateSessionDuration)
	})

	r.Route("/session-pauses", func(r chi.Router) {
		r.Post("/", sessionPauseHandler.CreateSessionPause)
		r.Put("/{id}/end", sessionPauseHandler.EndSessionPause)
	})

	r.Route("/exercise-logs", func(r chi.Router) {
		r.Post("/", exerciseLogHandler.CreateExerciseLog)
	})

	// 5. Server
	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: r,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("Server failed to start", "error", err)
			os.Exit(1)
		}
	}()
	slog.Info("Server is ready to handle requests")

	// 6. Graceful Shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	sig := <-quit
	slog.Info("Shutting down server...", "signal", sig.String())

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("Server forced to shutdown", "error", err)
	}

	slog.Info("Server exited properly")
}
