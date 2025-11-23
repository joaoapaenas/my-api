--- 

**File:** `.air.toml`

```typescript
#:schema https://json.schemastore.org/any.json

root = "."
testdata_dir = "testdata"
tmp_dir = "tmp"

[build]
  args_bin = []
  bin = "./tmp/main.exe"
  cmd = "go build -o ./tmp/main.exe ./cmd/api"
  delay = 1000
  exclude_dir = ["assets", "tmp", "vendor", "testdata", "docs"]
  exclude_file = []
  exclude_regex = ["_test.go"]
  exclude_unchanged = false
  follow_symlink = false
  full_bin = ""
  include_dir = []
  include_ext = ["go", "tpl", "tmpl", "html"]
  include_file = []
  kill_delay = "0s"
  log = "build-errors.log"
  poll = false
  poll_interval = 0
  post_cmd = []
  pre_cmd = []
  rerun = false
  rerun_delay = 500
  send_interrupt = false
  stop_on_error = false

[color]
  app = ""
  build = "yellow"
  main = "magenta"
  runner = "green"
  watcher = "cyan"

[log]
  main_only = false
  silent = false
  time = false

[misc]
  clean_on_exit = false

[proxy]
  app_port = 0
  enabled = false
  proxy_port = 0

[screen]
  clear_on_rebuild = false
  keep_scroll = true

```

--- 

**File:** `collect_code.py`

```typescript
import os
import sys

# --- Configuration ---

# 1. The root directory of your project.
#    os.getcwd() assumes you run the script from your project's root folder.
PROJECT_ROOT = os.getcwd()

# 2. The name of the file to save all the code into.
OUTPUT_FILE = "output.md"

# 3. File extensions to look for.
FILE_EXTENSIONS = (".py", ".go", ".json", ".yaml", ".toml")

# 4. Directories to exclude.
#    This is crucial to avoid including thousands of files from node_modules.
EXCLUDE_DIRS = {
    "node_modules",
    ".git",
    ".idea",
    ".pytest_cache",
    ".venv",
    "tests",
    ".docs",
    ".__pycache__",
    ".expo",
    "assets",
    "web-build",
    "dist",
    "build",
    # Add any other directories you want to ignore here
}

# --- End of Configuration ---


def is_excluded(path, root_path):
    """Check if a path is in one of the excluded directories."""
    relative_path = os.path.relpath(path, root_path)
    parts = relative_path.split(os.path.sep)
    return any(part in EXCLUDE_DIRS for part in parts)


def collect_files():
    """Walks through the project, finds relevant files, and writes them to the output file."""
    file_count = 0

    # Open the output file in write mode, which will overwrite it if it exists.
    # We use utf-8 encoding for compatibility with all source code characters.
    try:
        with open(OUTPUT_FILE, "w", encoding="utf-8") as outfile:
            print(f"Starting file collection in '{PROJECT_ROOT}'...")
            print(f"Output will be saved to '{OUTPUT_FILE}'")

            # os.walk is the perfect tool for traversing a directory tree.
            for root, dirs, files in os.walk(PROJECT_ROOT, topdown=True):
                # Efficiently skip excluded directories
                dirs[:] = [d for d in dirs if d not in EXCLUDE_DIRS]

                for filename in files:
                    if filename.endswith(FILE_EXTENSIONS):
                        file_path = os.path.join(root, filename)

                        try:
                            with open(
                                file_path, "r", encoding="utf-8", errors="ignore"
                            ) as infile:
                                content = infile.read()

                                # Get a clean, relative path for the header
                                relative_path = os.path.relpath(file_path, PROJECT_ROOT)

                                # Use forward slashes for cross-platform consistency in the output
                                header_path = relative_path.replace(os.path.sep, "/")

                                print(f"  -> Adding {header_path}")

                                # Write the formatted content to the output file
                                outfile.write(f"--- \n\n")
                                outfile.write(f"**File:** `{header_path}`\n\n")
                                outfile.write("```typescript\n")
                                outfile.write(content)
                                outfile.write("\n```\n\n")

                                file_count += 1
                        except Exception as e:
                            print(f"    [!] Error reading file {file_path}: {e}")

    except IOError as e:
        print(f"Error: Could not write to output file {OUTPUT_FILE}: {e}")
        sys.exit(1)

    return file_count


if __name__ == "__main__":
    total_files = collect_files()
    if total_files > 0:
        print(f"\n✅ Success! Combined {total_files} files into '{OUTPUT_FILE}'.")
    else:
        print("\n⚠️ Warning: No files with the specified extensions were found.")

```

--- 

**File:** `sqlc.yaml`

```typescript
version: "2"
sql:
  - engine: "sqlite"
    queries: "sql/queries"
    schema: "sql/schema"
    gen:
      go:
        package: "database"
        out: "internal/database"
        emit_json_tags: true
        emit_interface: true # Essential for mocking in unit tests

```

--- 

**File:** `cmd/api/main.go`

```typescript
package main

import (
	"context"
	"database/sql"
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
	customMiddleware "github.com/joaoapaenas/my-api/internal/middleware"
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
		slog.Error("Failed to load config", "error", err)
		os.Exit(1)
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
	analyticsRepo := repository.NewSQLAnalyticsRepository(queries)

	// Services
	userService := service.NewUserManager(userRepo)
	subjectService := service.NewSubjectManager(subjectRepo)
	topicService := service.NewTopicManager(topicRepo)
	studyCycleService := service.NewStudyCycleManager(studyCycleRepo)
	cycleItemService := service.NewCycleItemManager(cycleItemRepo)
	studySessionService := service.NewStudySessionManager(studySessionRepo)
	sessionPauseService := service.NewSessionPauseManager(sessionPauseRepo)
	exerciseLogService := service.NewExerciseLogManager(exerciseLogRepo)
	analyticsService := service.NewAnalyticsManager(analyticsRepo)

	// Handlers
	userHandler := handler.NewUserHandler(userService)
	subjectHandler := handler.NewSubjectHandler(subjectService)
	topicHandler := handler.NewTopicHandler(topicService)
	studyCycleHandler := handler.NewStudyCycleHandler(studyCycleService)
	cycleItemHandler := handler.NewCycleItemHandler(cycleItemService)
	studySessionHandler := handler.NewStudySessionHandler(studySessionService)
	sessionPauseHandler := handler.NewSessionPauseHandler(sessionPauseService)
	exerciseLogHandler := handler.NewExerciseLogHandler(exerciseLogService)
	analyticsHandler := handler.NewAnalyticsHandler(analyticsService)

	// 4. Router Setup
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(customMiddleware.RequestLogger)
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

	// Middleware
	basicAuth := customMiddleware.NewBasicAuthMiddleware(userService)

	r.Route("/users", func(r chi.Router) {
		r.Post("/", userHandler.CreateUser)
		r.Get("/{email}", userHandler.GetUser)
	})

	r.Route("/subjects", func(r chi.Router) {
		r.Use(basicAuth.BasicAuth)
		r.Post("/", subjectHandler.CreateSubject)
		r.Get("/", subjectHandler.ListSubjects)
		r.Get("/{id}", subjectHandler.GetSubject)
		r.Put("/{id}", subjectHandler.UpdateSubject)
		r.Delete("/{id}", subjectHandler.DeleteSubject)
		r.Post("/{id}/topics", topicHandler.CreateTopic)
		r.Get("/{id}/topics", topicHandler.ListTopics)
	})

	r.Route("/topics", func(r chi.Router) {
		r.Use(basicAuth.BasicAuth)
		r.Get("/{id}", topicHandler.GetTopic)
		r.Put("/{id}", topicHandler.UpdateTopic)
		r.Delete("/{id}", topicHandler.DeleteTopic)
	})

	r.Route("/study-cycles", func(r chi.Router) {
		r.Use(basicAuth.BasicAuth)
		r.Post("/", studyCycleHandler.CreateStudyCycle)
		r.Get("/active", studyCycleHandler.GetActiveStudyCycle)
		r.Get("/active/items", studyCycleHandler.GetActiveCycleWithItems) // Round-robin
		r.Get("/{id}", studyCycleHandler.GetStudyCycle)
		r.Put("/{id}", studyCycleHandler.UpdateStudyCycle)
		r.Delete("/{id}", studyCycleHandler.DeleteStudyCycle)
		r.Post("/{id}/items", cycleItemHandler.CreateCycleItem)
		r.Get("/{id}/items", cycleItemHandler.ListCycleItems)
	})

	r.Route("/cycle-items", func(r chi.Router) {
		r.Use(basicAuth.BasicAuth)
		r.Get("/{id}", cycleItemHandler.GetCycleItem)
		r.Put("/{id}", cycleItemHandler.UpdateCycleItem)
		r.Delete("/{id}", cycleItemHandler.DeleteCycleItem)
	})

	r.Route("/study-sessions", func(r chi.Router) {
		r.Use(basicAuth.BasicAuth)
		r.Post("/", studySessionHandler.CreateStudySession)
		r.Get("/open", studySessionHandler.GetOpenSession) // Crash recovery
		r.Get("/{id}", studySessionHandler.GetStudySession)
		r.Put("/{id}", studySessionHandler.UpdateSessionDuration)
		r.Delete("/{id}", studySessionHandler.DeleteStudySession)
	})

	r.Route("/session-pauses", func(r chi.Router) {
		r.Use(basicAuth.BasicAuth)
		r.Post("/", sessionPauseHandler.CreateSessionPause)
		r.Get("/{id}", sessionPauseHandler.GetSessionPause)
		r.Put("/{id}/end", sessionPauseHandler.EndSessionPause)
		r.Delete("/{id}", sessionPauseHandler.DeleteSessionPause)
	})

	r.Route("/exercise-logs", func(r chi.Router) {
		r.Use(basicAuth.BasicAuth)
		r.Post("/", exerciseLogHandler.CreateExerciseLog)
		r.Get("/{id}", exerciseLogHandler.GetExerciseLog)
		r.Delete("/{id}", exerciseLogHandler.DeleteExerciseLog)
	})

	// Analytics routes
	r.Route("/analytics", func(r chi.Router) {
		r.Use(basicAuth.BasicAuth)
		r.Get("/time-by-subject", analyticsHandler.GetTimeReportBySubject)
		r.Get("/accuracy-by-subject", analyticsHandler.GetAccuracyBySubject)
		r.Get("/accuracy-by-topic/{subject_id}", analyticsHandler.GetAccuracyByTopic)
		r.Get("/heatmap", analyticsHandler.GetActivityHeatmap)
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

```

--- 

**File:** `cmd/migrate/main.go`

```typescript
package main

import (
	"database/sql"
	"log"
	"os"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	db, err := sql.Open("sqlite3", "dev.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	driver, err := sqlite3.WithInstance(db, &sqlite3.Config{})
	if err != nil {
		log.Fatal(err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://sql/schema",
		"sqlite3",
		driver,
	)
	if err != nil {
		log.Fatal(err)
	}

	if len(os.Args) > 1 && os.Args[1] == "down" {
		if err := m.Down(); err != nil && err != migrate.ErrNoChange {
			log.Fatal(err)
		}
		log.Println("Migration down successful")
	} else {
		if err := m.Up(); err != nil && err != migrate.ErrNoChange {
			log.Fatal(err)
		}
		log.Println("Migration up successful")
	}
}

```

--- 

**File:** `docs/docs.go`

```typescript
// Package docs Code generated by swaggo/swag. DO NOT EDIT
package docs

import "github.com/swaggo/swag"

const docTemplate = `{
    "schemes": {{ marshal .Schemes }},
    "swagger": "2.0",
    "info": {
        "description": "{{escape .Description}}",
        "title": "{{.Title}}",
        "contact": {},
        "version": "{{.Version}}"
    },
    "host": "{{.Host}}",
    "basePath": "{{.BasePath}}",
    "paths": {
        "/analytics/accuracy-by-subject": {
            "get": {
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "analytics"
                ],
                "summary": "Get accuracy report by subject",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "array",
                            "items": {
                                "$ref": "#/definitions/handler.AccuracyReportResponse"
                            }
                        }
                    }
                }
            }
        },
        "/analytics/accuracy-by-topic/{subject_id}": {
            "get": {
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "analytics"
                ],
                "summary": "Get accuracy report by topic for a subject",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Subject ID",
                        "name": "subject_id",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "array",
                            "items": {
                                "$ref": "#/definitions/handler.TopicAccuracyResponse"
                            }
                        }
                    }
                }
            }
        },
        "/analytics/heatmap": {
            "get": {
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "analytics"
                ],
                "summary": "Get activity heatmap data",
                "parameters": [
                    {
                        "type": "integer",
                        "default": 30,
                        "description": "Number of days",
                        "name": "days",
                        "in": "query"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "array",
                            "items": {
                                "$ref": "#/definitions/handler.HeatmapDayResponse"
                            }
                        }
                    }
                }
            }
        },
        "/analytics/time-by-subject": {
            "get": {
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "analytics"
                ],
                "summary": "Get time tracking report by subject",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Start date (YYYY-MM-DD)",
                        "name": "start_date",
                        "in": "query"
                    },
                    {
                        "type": "string",
                        "description": "End date (YYYY-MM-DD)",
                        "name": "end_date",
                        "in": "query"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "array",
                            "items": {
                                "$ref": "#/definitions/handler.TimeReportResponse"
                            }
                        }
                    }
                }
            }
        },
        "/cycle-items/{id}": {
            "get": {
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "cycle_items"
                ],
                "summary": "Get a cycle item by ID",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Item ID",
                        "name": "id",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/handler.CycleItemResponse"
                        }
                    }
                }
            },
            "put": {
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "cycle_items"
                ],
                "summary": "Update a cycle item",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Item ID",
                        "name": "id",
                        "in": "path",
                        "required": true
                    },
                    {
                        "description": "Cycle item info",
                        "name": "input",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/handler.UpdateCycleItemRequest"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            },
            "delete": {
                "tags": [
                    "cycle_items"
                ],
                "summary": "Delete a cycle item",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Item ID",
                        "name": "id",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "204": {
                        "description": "No Content"
                    }
                }
            }
        },
        "/exercise-logs": {
            "post": {
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "exercise_logs"
                ],
                "summary": "Create a new exercise log",
                "parameters": [
                    {
                        "description": "Exercise log info",
                        "name": "input",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/handler.CreateExerciseLogRequest"
                        }
                    }
                ],
                "responses": {
                    "201": {
                        "description": "Created",
                        "schema": {
                            "$ref": "#/definitions/handler.ExerciseLogResponse"
                        }
                    }
                }
            }
        },
        "/exercise-logs/{id}": {
            "get": {
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "exercise_logs"
                ],
                "summary": "Get an exercise log by ID",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Log ID",
                        "name": "id",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/handler.ExerciseLogResponse"
                        }
                    }
                }
            },
            "delete": {
                "tags": [
                    "exercise_logs"
                ],
                "summary": "Delete an exercise log",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Log ID",
                        "name": "id",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "204": {
                        "description": "No Content"
                    }
                }
            }
        },
        "/session-pauses": {
            "post": {
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "session_pauses"
                ],
                "summary": "Create a new session pause",
                "parameters": [
                    {
                        "description": "Session pause info",
                        "name": "input",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/handler.CreateSessionPauseRequest"
                        }
                    }
                ],
                "responses": {
                    "201": {
                        "description": "Created",
                        "schema": {
                            "$ref": "#/definitions/handler.SessionPauseResponse"
                        }
                    }
                }
            }
        },
        "/session-pauses/{id}": {
            "get": {
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "session_pauses"
                ],
                "summary": "Get a session pause by ID",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Pause ID",
                        "name": "id",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/handler.SessionPauseResponse"
                        }
                    }
                }
            },
            "delete": {
                "tags": [
                    "session_pauses"
                ],
                "summary": "Delete a session pause",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Pause ID",
                        "name": "id",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "204": {
                        "description": "No Content"
                    }
                }
            }
        },
        "/session-pauses/{id}/end": {
            "put": {
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "session_pauses"
                ],
                "summary": "End a session pause",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Pause ID",
                        "name": "id",
                        "in": "path",
                        "required": true
                    },
                    {
                        "description": "End pause info",
                        "name": "input",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/handler.EndSessionPauseRequest"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            }
        },
        "/study-cycles": {
            "post": {
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "study_cycles"
                ],
                "summary": "Create a new study cycle",
                "parameters": [
                    {
                        "description": "Study cycle info",
                        "name": "input",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/handler.CreateStudyCycleRequest"
                        }
                    }
                ],
                "responses": {
                    "201": {
                        "description": "Created",
                        "schema": {
                            "$ref": "#/definitions/handler.StudyCycleResponse"
                        }
                    }
                }
            }
        },
        "/study-cycles/active": {
            "get": {
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "study_cycles"
                ],
                "summary": "Get the active study cycle",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/handler.StudyCycleResponse"
                        }
                    }
                }
            }
        },
        "/study-cycles/active/items": {
            "get": {
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "study_cycles"
                ],
                "summary": "Get active cycle with all items (round-robin)",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "array",
                            "items": {
                                "$ref": "#/definitions/handler.CycleItemWithSubjectResponse"
                            }
                        }
                    }
                }
            }
        },
        "/study-cycles/{id}": {
            "get": {
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "study_cycles"
                ],
                "summary": "Get a study cycle by ID",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Cycle ID",
                        "name": "id",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/handler.StudyCycleResponse"
                        }
                    }
                }
            },
            "put": {
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "study_cycles"
                ],
                "summary": "Update a study cycle",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Cycle ID",
                        "name": "id",
                        "in": "path",
                        "required": true
                    },
                    {
                        "description": "Study cycle info",
                        "name": "input",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/handler.UpdateStudyCycleRequest"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            },
            "delete": {
                "tags": [
                    "study_cycles"
                ],
                "summary": "Delete a study cycle",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Cycle ID",
                        "name": "id",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "204": {
                        "description": "No Content"
                    }
                }
            }
        },
        "/study-cycles/{id}/items": {
            "get": {
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "cycle_items"
                ],
                "summary": "List all items for a cycle",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Cycle ID",
                        "name": "id",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "array",
                            "items": {
                                "$ref": "#/definitions/handler.CycleItemResponse"
                            }
                        }
                    }
                }
            },
            "post": {
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "cycle_items"
                ],
                "summary": "Create a new cycle item",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Cycle ID",
                        "name": "id",
                        "in": "path",
                        "required": true
                    },
                    {
                        "description": "Cycle item info",
                        "name": "input",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/handler.CreateCycleItemRequest"
                        }
                    }
                ],
                "responses": {
                    "201": {
                        "description": "Created",
                        "schema": {
                            "$ref": "#/definitions/handler.CycleItemResponse"
                        }
                    }
                }
            }
        },
        "/study-sessions": {
            "post": {
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "study_sessions"
                ],
                "summary": "Create a new study session",
                "parameters": [
                    {
                        "description": "Study session info",
                        "name": "input",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/handler.CreateStudySessionRequest"
                        }
                    }
                ],
                "responses": {
                    "201": {
                        "description": "Created",
                        "schema": {
                            "$ref": "#/definitions/handler.StudySessionResponse"
                        }
                    }
                }
            }
        },
        "/study-sessions/open": {
            "get": {
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "study_sessions"
                ],
                "summary": "Get open/unfinished session (crash recovery)",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/handler.OpenSessionResponse"
                        }
                    }
                }
            }
        },
        "/study-sessions/{id}": {
            "get": {
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "study_sessions"
                ],
                "summary": "Get a study session by ID",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Session ID",
                        "name": "id",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/handler.StudySessionResponse"
                        }
                    }
                }
            },
            "put": {
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "study_sessions"
                ],
                "summary": "Update study session duration",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Session ID",
                        "name": "id",
                        "in": "path",
                        "required": true
                    },
                    {
                        "description": "Session duration info",
                        "name": "input",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/handler.UpdateSessionDurationRequest"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            },
            "delete": {
                "tags": [
                    "study_sessions"
                ],
                "summary": "Delete a study session",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Session ID",
                        "name": "id",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "204": {
                        "description": "No Content"
                    }
                }
            }
        },
        "/subjects": {
            "get": {
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "subjects"
                ],
                "summary": "List all subjects",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "array",
                            "items": {
                                "$ref": "#/definitions/handler.SubjectResponse"
                            }
                        }
                    }
                }
            },
            "post": {
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "subjects"
                ],
                "summary": "Create a new subject",
                "parameters": [
                    {
                        "description": "Subject info",
                        "name": "input",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/handler.CreateSubjectRequest"
                        }
                    }
                ],
                "responses": {
                    "201": {
                        "description": "Created",
                        "schema": {
                            "$ref": "#/definitions/handler.SubjectResponse"
                        }
                    }
                }
            }
        },
        "/subjects/{id}": {
            "get": {
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "subjects"
                ],
                "summary": "Get a subject by ID",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Subject ID",
                        "name": "id",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/handler.SubjectResponse"
                        }
                    }
                }
            },
            "put": {
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "subjects"
                ],
                "summary": "Update a subject",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Subject ID",
                        "name": "id",
                        "in": "path",
                        "required": true
                    },
                    {
                        "description": "Subject info",
                        "name": "input",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/handler.UpdateSubjectRequest"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            },
            "delete": {
                "tags": [
                    "subjects"
                ],
                "summary": "Delete a subject",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Subject ID",
                        "name": "id",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "204": {
                        "description": "No Content"
                    }
                }
            }
        },
        "/subjects/{id}/topics": {
            "get": {
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "topics"
                ],
                "summary": "List all topics for a subject",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Subject ID",
                        "name": "id",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "array",
                            "items": {
                                "$ref": "#/definitions/handler.TopicResponse"
                            }
                        }
                    }
                }
            },
            "post": {
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "topics"
                ],
                "summary": "Create a new topic for a subject",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Subject ID",
                        "name": "id",
                        "in": "path",
                        "required": true
                    },
                    {
                        "description": "Topic info",
                        "name": "input",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/handler.CreateTopicRequest"
                        }
                    }
                ],
                "responses": {
                    "201": {
                        "description": "Created",
                        "schema": {
                            "$ref": "#/definitions/handler.TopicResponse"
                        }
                    }
                }
            }
        },
        "/topics/{id}": {
            "get": {
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "topics"
                ],
                "summary": "Get a topic by ID",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Topic ID",
                        "name": "id",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/handler.TopicResponse"
                        }
                    }
                }
            },
            "put": {
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "topics"
                ],
                "summary": "Update a topic",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Topic ID",
                        "name": "id",
                        "in": "path",
                        "required": true
                    },
                    {
                        "description": "Topic info",
                        "name": "input",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/handler.UpdateTopicRequest"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            },
            "delete": {
                "tags": [
                    "topics"
                ],
                "summary": "Delete a topic",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Topic ID",
                        "name": "id",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "204": {
                        "description": "No Content"
                    }
                }
            }
        },
        "/users": {
            "post": {
                "description": "Create a user with email and name",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "users"
                ],
                "summary": "Create a new user",
                "parameters": [
                    {
                        "description": "User info",
                        "name": "input",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/handler.CreateUserRequest"
                        }
                    }
                ],
                "responses": {
                    "201": {
                        "description": "Created",
                        "schema": {
                            "$ref": "#/definitions/database.User"
                        }
                    },
                    "400": {
                        "description": "Invalid request",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "500": {
                        "description": "Internal server error",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            }
        },
        "/users/{email}": {
            "get": {
                "tags": [
                    "users"
                ],
                "summary": "Get user by Email",
                "parameters": [
                    {
                        "type": "string",
                        "description": "User Email",
                        "name": "email",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/database.User"
                        }
                    },
                    "404": {
                        "description": "User not found",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            }
        }
    },
    "definitions": {
        "database.User": {
            "type": "object",
            "properties": {
                "created_at": {
                    "type": "string"
                },
                "email": {
                    "type": "string"
                },
                "id": {
                    "type": "string"
                },
                "name": {
                    "type": "string"
                }
            }
        },
        "handler.AccuracyReportResponse": {
            "type": "object",
            "properties": {
                "accuracy_percentage": {
                    "type": "number"
                },
                "color_hex": {
                    "type": "string"
                },
                "subject_id": {
                    "type": "string"
                },
                "subject_name": {
                    "type": "string"
                },
                "total_correct": {
                    "type": "integer"
                },
                "total_questions": {
                    "type": "integer"
                }
            }
        },
        "handler.CreateCycleItemRequest": {
            "type": "object",
            "required": [
                "order_index",
                "subject_id"
            ],
            "properties": {
                "order_index": {
                    "type": "integer",
                    "minimum": 1
                },
                "planned_duration_minutes": {
                    "type": "integer",
                    "minimum": 1
                },
                "subject_id": {
                    "type": "string"
                }
            }
        },
        "handler.CreateExerciseLogRequest": {
            "type": "object",
            "required": [
                "correct_count",
                "questions_count",
                "subject_id"
            ],
            "properties": {
                "correct_count": {
                    "type": "integer",
                    "minimum": 0
                },
                "questions_count": {
                    "type": "integer",
                    "minimum": 0
                },
                "session_id": {
                    "type": "string"
                },
                "subject_id": {
                    "type": "string"
                },
                "topic_id": {
                    "type": "string"
                }
            }
        },
        "handler.CreateSessionPauseRequest": {
            "type": "object",
            "required": [
                "session_id",
                "started_at"
            ],
            "properties": {
                "session_id": {
                    "type": "string"
                },
                "started_at": {
                    "type": "string"
                }
            }
        },
        "handler.CreateStudyCycleRequest": {
            "type": "object",
            "required": [
                "name"
            ],
            "properties": {
                "description": {
                    "type": "string"
                },
                "is_active": {
                    "type": "boolean"
                },
                "name": {
                    "type": "string",
                    "minLength": 2
                }
            }
        },
        "handler.CreateStudySessionRequest": {
            "type": "object",
            "required": [
                "started_at",
                "subject_id"
            ],
            "properties": {
                "cycle_item_id": {
                    "type": "string"
                },
                "started_at": {
                    "type": "string"
                },
                "subject_id": {
                    "type": "string"
                }
            }
        },
        "handler.CreateSubjectRequest": {
            "type": "object",
            "required": [
                "name"
            ],
            "properties": {
                "color_hex": {
                    "type": "string"
                },
                "name": {
                    "type": "string",
                    "minLength": 2
                }
            }
        },
        "handler.CreateTopicRequest": {
            "type": "object",
            "required": [
                "name"
            ],
            "properties": {
                "name": {
                    "type": "string",
                    "minLength": 2
                }
            }
        },
        "handler.CreateUserRequest": {
            "type": "object",
            "required": [
                "email",
                "name"
            ],
            "properties": {
                "email": {
                    "description": "required: cannot be empty\nemail: must be a valid email format",
                    "type": "string"
                },
                "name": {
                    "description": "min=2: must be at least 2 chars",
                    "type": "string",
                    "minLength": 2
                }
            }
        },
        "handler.CycleItemResponse": {
            "type": "object",
            "properties": {
                "created_at": {
                    "type": "string"
                },
                "cycle_id": {
                    "type": "string"
                },
                "id": {
                    "type": "string"
                },
                "order_index": {
                    "type": "integer"
                },
                "planned_duration_minutes": {
                    "type": "integer"
                },
                "subject_id": {
                    "type": "string"
                },
                "updated_at": {
                    "type": "string"
                }
            }
        },
        "handler.CycleItemWithSubjectResponse": {
            "type": "object",
            "properties": {
                "color_hex": {
                    "type": "string"
                },
                "cycle_item_id": {
                    "type": "string"
                },
                "order_index": {
                    "type": "integer"
                },
                "planned_duration_minutes": {
                    "type": "integer"
                },
                "subject_id": {
                    "type": "string"
                },
                "subject_name": {
                    "type": "string"
                }
            }
        },
        "handler.EndSessionPauseRequest": {
            "type": "object",
            "required": [
                "ended_at"
            ],
            "properties": {
                "ended_at": {
                    "type": "string"
                }
            }
        },
        "handler.ExerciseLogResponse": {
            "type": "object",
            "properties": {
                "correct_count": {
                    "type": "integer"
                },
                "created_at": {
                    "type": "string"
                },
                "id": {
                    "type": "string"
                },
                "questions_count": {
                    "type": "integer"
                },
                "session_id": {
                    "type": "string"
                },
                "subject_id": {
                    "type": "string"
                },
                "topic_id": {
                    "type": "string"
                }
            }
        },
        "handler.HeatmapDayResponse": {
            "type": "object",
            "properties": {
                "sessions_count": {
                    "type": "integer"
                },
                "study_date": {
                    "type": "string"
                },
                "total_seconds": {
                    "type": "integer"
                }
            }
        },
        "handler.OpenSessionResponse": {
            "type": "object",
            "properties": {
                "color_hex": {
                    "type": "string"
                },
                "cycle_item_id": {
                    "type": "string"
                },
                "id": {
                    "type": "string"
                },
                "started_at": {
                    "type": "string"
                },
                "subject_id": {
                    "type": "string"
                },
                "subject_name": {
                    "type": "string"
                }
            }
        },
        "handler.SessionPauseResponse": {
            "type": "object",
            "properties": {
                "ended_at": {
                    "type": "string"
                },
                "id": {
                    "type": "string"
                },
                "session_id": {
                    "type": "string"
                },
                "started_at": {
                    "type": "string"
                }
            }
        },
        "handler.StudyCycleResponse": {
            "type": "object",
            "properties": {
                "created_at": {
                    "type": "string"
                },
                "deleted_at": {
                    "type": "string"
                },
                "description": {
                    "type": "string"
                },
                "id": {
                    "type": "string"
                },
                "is_active": {
                    "type": "integer"
                },
                "name": {
                    "type": "string"
                },
                "updated_at": {
                    "type": "string"
                }
            }
        },
        "handler.StudySessionResponse": {
            "type": "object",
            "properties": {
                "cycle_item_id": {
                    "type": "string"
                },
                "finished_at": {
                    "type": "string"
                },
                "gross_duration_seconds": {
                    "type": "integer"
                },
                "id": {
                    "type": "string"
                },
                "net_duration_seconds": {
                    "type": "integer"
                },
                "notes": {
                    "type": "string"
                },
                "started_at": {
                    "type": "string"
                },
                "subject_id": {
                    "type": "string"
                }
            }
        },
        "handler.SubjectResponse": {
            "type": "object",
            "properties": {
                "color_hex": {
                    "type": "string"
                },
                "created_at": {
                    "type": "string"
                },
                "deleted_at": {
                    "type": "string"
                },
                "id": {
                    "type": "string"
                },
                "name": {
                    "type": "string"
                },
                "updated_at": {
                    "type": "string"
                }
            }
        },
        "handler.TimeReportResponse": {
            "type": "object",
            "properties": {
                "color_hex": {
                    "type": "string"
                },
                "sessions_count": {
                    "type": "integer"
                },
                "subject_id": {
                    "type": "string"
                },
                "subject_name": {
                    "type": "string"
                },
                "total_hours_net": {
                    "type": "number"
                }
            }
        },
        "handler.TopicAccuracyResponse": {
            "type": "object",
            "properties": {
                "accuracy_percentage": {
                    "type": "number"
                },
                "topic_id": {
                    "type": "string"
                },
                "topic_name": {
                    "type": "string"
                },
                "total_correct": {
                    "type": "integer"
                },
                "total_questions": {
                    "type": "integer"
                }
            }
        },
        "handler.TopicResponse": {
            "type": "object",
            "properties": {
                "created_at": {
                    "type": "string"
                },
                "deleted_at": {
                    "type": "string"
                },
                "id": {
                    "type": "string"
                },
                "name": {
                    "type": "string"
                },
                "subject_id": {
                    "type": "string"
                },
                "updated_at": {
                    "type": "string"
                }
            }
        },
        "handler.UpdateCycleItemRequest": {
            "type": "object",
            "required": [
                "order_index",
                "subject_id"
            ],
            "properties": {
                "order_index": {
                    "type": "integer",
                    "minimum": 1
                },
                "planned_duration_minutes": {
                    "type": "integer",
                    "minimum": 1
                },
                "subject_id": {
                    "type": "string"
                }
            }
        },
        "handler.UpdateSessionDurationRequest": {
            "type": "object",
            "properties": {
                "finished_at": {
                    "type": "string"
                },
                "gross_duration_seconds": {
                    "type": "integer"
                },
                "net_duration_seconds": {
                    "type": "integer"
                },
                "notes": {
                    "type": "string"
                }
            }
        },
        "handler.UpdateStudyCycleRequest": {
            "type": "object",
            "required": [
                "name"
            ],
            "properties": {
                "description": {
                    "type": "string"
                },
                "is_active": {
                    "type": "boolean"
                },
                "name": {
                    "type": "string",
                    "minLength": 2
                }
            }
        },
        "handler.UpdateSubjectRequest": {
            "type": "object",
            "required": [
                "name"
            ],
            "properties": {
                "color_hex": {
                    "type": "string"
                },
                "name": {
                    "type": "string",
                    "minLength": 2
                }
            }
        },
        "handler.UpdateTopicRequest": {
            "type": "object",
            "required": [
                "name"
            ],
            "properties": {
                "name": {
                    "type": "string",
                    "minLength": 2
                }
            }
        }
    }
}`

// SwaggerInfo holds exported Swagger Info so clients can modify it
var SwaggerInfo = &swag.Spec{
	Version:          "1.0",
	Host:             "localhost:8080",
	BasePath:         "/",
	Schemes:          []string{},
	Title:            "My Go API",
	Description:      "Production ready starter guide.",
	InfoInstanceName: "swagger",
	SwaggerTemplate:  docTemplate,
	LeftDelim:        "{{",
	RightDelim:       "}}",
}

func init() {
	swag.Register(SwaggerInfo.InstanceName(), SwaggerInfo)
}

```

--- 

**File:** `docs/swagger.json`

```typescript
{
    "swagger": "2.0",
    "info": {
        "description": "Production ready starter guide.",
        "title": "My Go API",
        "contact": {},
        "version": "1.0"
    },
    "host": "localhost:8080",
    "basePath": "/",
    "paths": {
        "/analytics/accuracy-by-subject": {
            "get": {
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "analytics"
                ],
                "summary": "Get accuracy report by subject",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "array",
                            "items": {
                                "$ref": "#/definitions/handler.AccuracyReportResponse"
                            }
                        }
                    }
                }
            }
        },
        "/analytics/accuracy-by-topic/{subject_id}": {
            "get": {
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "analytics"
                ],
                "summary": "Get accuracy report by topic for a subject",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Subject ID",
                        "name": "subject_id",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "array",
                            "items": {
                                "$ref": "#/definitions/handler.TopicAccuracyResponse"
                            }
                        }
                    }
                }
            }
        },
        "/analytics/heatmap": {
            "get": {
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "analytics"
                ],
                "summary": "Get activity heatmap data",
                "parameters": [
                    {
                        "type": "integer",
                        "default": 30,
                        "description": "Number of days",
                        "name": "days",
                        "in": "query"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "array",
                            "items": {
                                "$ref": "#/definitions/handler.HeatmapDayResponse"
                            }
                        }
                    }
                }
            }
        },
        "/analytics/time-by-subject": {
            "get": {
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "analytics"
                ],
                "summary": "Get time tracking report by subject",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Start date (YYYY-MM-DD)",
                        "name": "start_date",
                        "in": "query"
                    },
                    {
                        "type": "string",
                        "description": "End date (YYYY-MM-DD)",
                        "name": "end_date",
                        "in": "query"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "array",
                            "items": {
                                "$ref": "#/definitions/handler.TimeReportResponse"
                            }
                        }
                    }
                }
            }
        },
        "/cycle-items/{id}": {
            "get": {
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "cycle_items"
                ],
                "summary": "Get a cycle item by ID",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Item ID",
                        "name": "id",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/handler.CycleItemResponse"
                        }
                    }
                }
            },
            "put": {
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "cycle_items"
                ],
                "summary": "Update a cycle item",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Item ID",
                        "name": "id",
                        "in": "path",
                        "required": true
                    },
                    {
                        "description": "Cycle item info",
                        "name": "input",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/handler.UpdateCycleItemRequest"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            },
            "delete": {
                "tags": [
                    "cycle_items"
                ],
                "summary": "Delete a cycle item",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Item ID",
                        "name": "id",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "204": {
                        "description": "No Content"
                    }
                }
            }
        },
        "/exercise-logs": {
            "post": {
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "exercise_logs"
                ],
                "summary": "Create a new exercise log",
                "parameters": [
                    {
                        "description": "Exercise log info",
                        "name": "input",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/handler.CreateExerciseLogRequest"
                        }
                    }
                ],
                "responses": {
                    "201": {
                        "description": "Created",
                        "schema": {
                            "$ref": "#/definitions/handler.ExerciseLogResponse"
                        }
                    }
                }
            }
        },
        "/exercise-logs/{id}": {
            "get": {
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "exercise_logs"
                ],
                "summary": "Get an exercise log by ID",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Log ID",
                        "name": "id",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/handler.ExerciseLogResponse"
                        }
                    }
                }
            },
            "delete": {
                "tags": [
                    "exercise_logs"
                ],
                "summary": "Delete an exercise log",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Log ID",
                        "name": "id",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "204": {
                        "description": "No Content"
                    }
                }
            }
        },
        "/session-pauses": {
            "post": {
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "session_pauses"
                ],
                "summary": "Create a new session pause",
                "parameters": [
                    {
                        "description": "Session pause info",
                        "name": "input",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/handler.CreateSessionPauseRequest"
                        }
                    }
                ],
                "responses": {
                    "201": {
                        "description": "Created",
                        "schema": {
                            "$ref": "#/definitions/handler.SessionPauseResponse"
                        }
                    }
                }
            }
        },
        "/session-pauses/{id}": {
            "get": {
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "session_pauses"
                ],
                "summary": "Get a session pause by ID",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Pause ID",
                        "name": "id",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/handler.SessionPauseResponse"
                        }
                    }
                }
            },
            "delete": {
                "tags": [
                    "session_pauses"
                ],
                "summary": "Delete a session pause",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Pause ID",
                        "name": "id",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "204": {
                        "description": "No Content"
                    }
                }
            }
        },
        "/session-pauses/{id}/end": {
            "put": {
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "session_pauses"
                ],
                "summary": "End a session pause",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Pause ID",
                        "name": "id",
                        "in": "path",
                        "required": true
                    },
                    {
                        "description": "End pause info",
                        "name": "input",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/handler.EndSessionPauseRequest"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            }
        },
        "/study-cycles": {
            "post": {
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "study_cycles"
                ],
                "summary": "Create a new study cycle",
                "parameters": [
                    {
                        "description": "Study cycle info",
                        "name": "input",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/handler.CreateStudyCycleRequest"
                        }
                    }
                ],
                "responses": {
                    "201": {
                        "description": "Created",
                        "schema": {
                            "$ref": "#/definitions/handler.StudyCycleResponse"
                        }
                    }
                }
            }
        },
        "/study-cycles/active": {
            "get": {
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "study_cycles"
                ],
                "summary": "Get the active study cycle",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/handler.StudyCycleResponse"
                        }
                    }
                }
            }
        },
        "/study-cycles/active/items": {
            "get": {
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "study_cycles"
                ],
                "summary": "Get active cycle with all items (round-robin)",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "array",
                            "items": {
                                "$ref": "#/definitions/handler.CycleItemWithSubjectResponse"
                            }
                        }
                    }
                }
            }
        },
        "/study-cycles/{id}": {
            "get": {
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "study_cycles"
                ],
                "summary": "Get a study cycle by ID",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Cycle ID",
                        "name": "id",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/handler.StudyCycleResponse"
                        }
                    }
                }
            },
            "put": {
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "study_cycles"
                ],
                "summary": "Update a study cycle",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Cycle ID",
                        "name": "id",
                        "in": "path",
                        "required": true
                    },
                    {
                        "description": "Study cycle info",
                        "name": "input",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/handler.UpdateStudyCycleRequest"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            },
            "delete": {
                "tags": [
                    "study_cycles"
                ],
                "summary": "Delete a study cycle",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Cycle ID",
                        "name": "id",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "204": {
                        "description": "No Content"
                    }
                }
            }
        },
        "/study-cycles/{id}/items": {
            "get": {
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "cycle_items"
                ],
                "summary": "List all items for a cycle",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Cycle ID",
                        "name": "id",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "array",
                            "items": {
                                "$ref": "#/definitions/handler.CycleItemResponse"
                            }
                        }
                    }
                }
            },
            "post": {
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "cycle_items"
                ],
                "summary": "Create a new cycle item",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Cycle ID",
                        "name": "id",
                        "in": "path",
                        "required": true
                    },
                    {
                        "description": "Cycle item info",
                        "name": "input",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/handler.CreateCycleItemRequest"
                        }
                    }
                ],
                "responses": {
                    "201": {
                        "description": "Created",
                        "schema": {
                            "$ref": "#/definitions/handler.CycleItemResponse"
                        }
                    }
                }
            }
        },
        "/study-sessions": {
            "post": {
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "study_sessions"
                ],
                "summary": "Create a new study session",
                "parameters": [
                    {
                        "description": "Study session info",
                        "name": "input",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/handler.CreateStudySessionRequest"
                        }
                    }
                ],
                "responses": {
                    "201": {
                        "description": "Created",
                        "schema": {
                            "$ref": "#/definitions/handler.StudySessionResponse"
                        }
                    }
                }
            }
        },
        "/study-sessions/open": {
            "get": {
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "study_sessions"
                ],
                "summary": "Get open/unfinished session (crash recovery)",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/handler.OpenSessionResponse"
                        }
                    }
                }
            }
        },
        "/study-sessions/{id}": {
            "get": {
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "study_sessions"
                ],
                "summary": "Get a study session by ID",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Session ID",
                        "name": "id",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/handler.StudySessionResponse"
                        }
                    }
                }
            },
            "put": {
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "study_sessions"
                ],
                "summary": "Update study session duration",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Session ID",
                        "name": "id",
                        "in": "path",
                        "required": true
                    },
                    {
                        "description": "Session duration info",
                        "name": "input",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/handler.UpdateSessionDurationRequest"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            },
            "delete": {
                "tags": [
                    "study_sessions"
                ],
                "summary": "Delete a study session",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Session ID",
                        "name": "id",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "204": {
                        "description": "No Content"
                    }
                }
            }
        },
        "/subjects": {
            "get": {
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "subjects"
                ],
                "summary": "List all subjects",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "array",
                            "items": {
                                "$ref": "#/definitions/handler.SubjectResponse"
                            }
                        }
                    }
                }
            },
            "post": {
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "subjects"
                ],
                "summary": "Create a new subject",
                "parameters": [
                    {
                        "description": "Subject info",
                        "name": "input",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/handler.CreateSubjectRequest"
                        }
                    }
                ],
                "responses": {
                    "201": {
                        "description": "Created",
                        "schema": {
                            "$ref": "#/definitions/handler.SubjectResponse"
                        }
                    }
                }
            }
        },
        "/subjects/{id}": {
            "get": {
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "subjects"
                ],
                "summary": "Get a subject by ID",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Subject ID",
                        "name": "id",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/handler.SubjectResponse"
                        }
                    }
                }
            },
            "put": {
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "subjects"
                ],
                "summary": "Update a subject",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Subject ID",
                        "name": "id",
                        "in": "path",
                        "required": true
                    },
                    {
                        "description": "Subject info",
                        "name": "input",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/handler.UpdateSubjectRequest"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            },
            "delete": {
                "tags": [
                    "subjects"
                ],
                "summary": "Delete a subject",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Subject ID",
                        "name": "id",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "204": {
                        "description": "No Content"
                    }
                }
            }
        },
        "/subjects/{id}/topics": {
            "get": {
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "topics"
                ],
                "summary": "List all topics for a subject",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Subject ID",
                        "name": "id",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "array",
                            "items": {
                                "$ref": "#/definitions/handler.TopicResponse"
                            }
                        }
                    }
                }
            },
            "post": {
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "topics"
                ],
                "summary": "Create a new topic for a subject",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Subject ID",
                        "name": "id",
                        "in": "path",
                        "required": true
                    },
                    {
                        "description": "Topic info",
                        "name": "input",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/handler.CreateTopicRequest"
                        }
                    }
                ],
                "responses": {
                    "201": {
                        "description": "Created",
                        "schema": {
                            "$ref": "#/definitions/handler.TopicResponse"
                        }
                    }
                }
            }
        },
        "/topics/{id}": {
            "get": {
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "topics"
                ],
                "summary": "Get a topic by ID",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Topic ID",
                        "name": "id",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/handler.TopicResponse"
                        }
                    }
                }
            },
            "put": {
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "topics"
                ],
                "summary": "Update a topic",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Topic ID",
                        "name": "id",
                        "in": "path",
                        "required": true
                    },
                    {
                        "description": "Topic info",
                        "name": "input",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/handler.UpdateTopicRequest"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            },
            "delete": {
                "tags": [
                    "topics"
                ],
                "summary": "Delete a topic",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Topic ID",
                        "name": "id",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "204": {
                        "description": "No Content"
                    }
                }
            }
        },
        "/users": {
            "post": {
                "description": "Create a user with email and name",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "users"
                ],
                "summary": "Create a new user",
                "parameters": [
                    {
                        "description": "User info",
                        "name": "input",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/handler.CreateUserRequest"
                        }
                    }
                ],
                "responses": {
                    "201": {
                        "description": "Created",
                        "schema": {
                            "$ref": "#/definitions/database.User"
                        }
                    },
                    "400": {
                        "description": "Invalid request",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "500": {
                        "description": "Internal server error",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            }
        },
        "/users/{email}": {
            "get": {
                "tags": [
                    "users"
                ],
                "summary": "Get user by Email",
                "parameters": [
                    {
                        "type": "string",
                        "description": "User Email",
                        "name": "email",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/database.User"
                        }
                    },
                    "404": {
                        "description": "User not found",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            }
        }
    },
    "definitions": {
        "database.User": {
            "type": "object",
            "properties": {
                "created_at": {
                    "type": "string"
                },
                "email": {
                    "type": "string"
                },
                "id": {
                    "type": "string"
                },
                "name": {
                    "type": "string"
                }
            }
        },
        "handler.AccuracyReportResponse": {
            "type": "object",
            "properties": {
                "accuracy_percentage": {
                    "type": "number"
                },
                "color_hex": {
                    "type": "string"
                },
                "subject_id": {
                    "type": "string"
                },
                "subject_name": {
                    "type": "string"
                },
                "total_correct": {
                    "type": "integer"
                },
                "total_questions": {
                    "type": "integer"
                }
            }
        },
        "handler.CreateCycleItemRequest": {
            "type": "object",
            "required": [
                "order_index",
                "subject_id"
            ],
            "properties": {
                "order_index": {
                    "type": "integer",
                    "minimum": 1
                },
                "planned_duration_minutes": {
                    "type": "integer",
                    "minimum": 1
                },
                "subject_id": {
                    "type": "string"
                }
            }
        },
        "handler.CreateExerciseLogRequest": {
            "type": "object",
            "required": [
                "correct_count",
                "questions_count",
                "subject_id"
            ],
            "properties": {
                "correct_count": {
                    "type": "integer",
                    "minimum": 0
                },
                "questions_count": {
                    "type": "integer",
                    "minimum": 0
                },
                "session_id": {
                    "type": "string"
                },
                "subject_id": {
                    "type": "string"
                },
                "topic_id": {
                    "type": "string"
                }
            }
        },
        "handler.CreateSessionPauseRequest": {
            "type": "object",
            "required": [
                "session_id",
                "started_at"
            ],
            "properties": {
                "session_id": {
                    "type": "string"
                },
                "started_at": {
                    "type": "string"
                }
            }
        },
        "handler.CreateStudyCycleRequest": {
            "type": "object",
            "required": [
                "name"
            ],
            "properties": {
                "description": {
                    "type": "string"
                },
                "is_active": {
                    "type": "boolean"
                },
                "name": {
                    "type": "string",
                    "minLength": 2
                }
            }
        },
        "handler.CreateStudySessionRequest": {
            "type": "object",
            "required": [
                "started_at",
                "subject_id"
            ],
            "properties": {
                "cycle_item_id": {
                    "type": "string"
                },
                "started_at": {
                    "type": "string"
                },
                "subject_id": {
                    "type": "string"
                }
            }
        },
        "handler.CreateSubjectRequest": {
            "type": "object",
            "required": [
                "name"
            ],
            "properties": {
                "color_hex": {
                    "type": "string"
                },
                "name": {
                    "type": "string",
                    "minLength": 2
                }
            }
        },
        "handler.CreateTopicRequest": {
            "type": "object",
            "required": [
                "name"
            ],
            "properties": {
                "name": {
                    "type": "string",
                    "minLength": 2
                }
            }
        },
        "handler.CreateUserRequest": {
            "type": "object",
            "required": [
                "email",
                "name"
            ],
            "properties": {
                "email": {
                    "description": "required: cannot be empty\nemail: must be a valid email format",
                    "type": "string"
                },
                "name": {
                    "description": "min=2: must be at least 2 chars",
                    "type": "string",
                    "minLength": 2
                }
            }
        },
        "handler.CycleItemResponse": {
            "type": "object",
            "properties": {
                "created_at": {
                    "type": "string"
                },
                "cycle_id": {
                    "type": "string"
                },
                "id": {
                    "type": "string"
                },
                "order_index": {
                    "type": "integer"
                },
                "planned_duration_minutes": {
                    "type": "integer"
                },
                "subject_id": {
                    "type": "string"
                },
                "updated_at": {
                    "type": "string"
                }
            }
        },
        "handler.CycleItemWithSubjectResponse": {
            "type": "object",
            "properties": {
                "color_hex": {
                    "type": "string"
                },
                "cycle_item_id": {
                    "type": "string"
                },
                "order_index": {
                    "type": "integer"
                },
                "planned_duration_minutes": {
                    "type": "integer"
                },
                "subject_id": {
                    "type": "string"
                },
                "subject_name": {
                    "type": "string"
                }
            }
        },
        "handler.EndSessionPauseRequest": {
            "type": "object",
            "required": [
                "ended_at"
            ],
            "properties": {
                "ended_at": {
                    "type": "string"
                }
            }
        },
        "handler.ExerciseLogResponse": {
            "type": "object",
            "properties": {
                "correct_count": {
                    "type": "integer"
                },
                "created_at": {
                    "type": "string"
                },
                "id": {
                    "type": "string"
                },
                "questions_count": {
                    "type": "integer"
                },
                "session_id": {
                    "type": "string"
                },
                "subject_id": {
                    "type": "string"
                },
                "topic_id": {
                    "type": "string"
                }
            }
        },
        "handler.HeatmapDayResponse": {
            "type": "object",
            "properties": {
                "sessions_count": {
                    "type": "integer"
                },
                "study_date": {
                    "type": "string"
                },
                "total_seconds": {
                    "type": "integer"
                }
            }
        },
        "handler.OpenSessionResponse": {
            "type": "object",
            "properties": {
                "color_hex": {
                    "type": "string"
                },
                "cycle_item_id": {
                    "type": "string"
                },
                "id": {
                    "type": "string"
                },
                "started_at": {
                    "type": "string"
                },
                "subject_id": {
                    "type": "string"
                },
                "subject_name": {
                    "type": "string"
                }
            }
        },
        "handler.SessionPauseResponse": {
            "type": "object",
            "properties": {
                "ended_at": {
                    "type": "string"
                },
                "id": {
                    "type": "string"
                },
                "session_id": {
                    "type": "string"
                },
                "started_at": {
                    "type": "string"
                }
            }
        },
        "handler.StudyCycleResponse": {
            "type": "object",
            "properties": {
                "created_at": {
                    "type": "string"
                },
                "deleted_at": {
                    "type": "string"
                },
                "description": {
                    "type": "string"
                },
                "id": {
                    "type": "string"
                },
                "is_active": {
                    "type": "integer"
                },
                "name": {
                    "type": "string"
                },
                "updated_at": {
                    "type": "string"
                }
            }
        },
        "handler.StudySessionResponse": {
            "type": "object",
            "properties": {
                "cycle_item_id": {
                    "type": "string"
                },
                "finished_at": {
                    "type": "string"
                },
                "gross_duration_seconds": {
                    "type": "integer"
                },
                "id": {
                    "type": "string"
                },
                "net_duration_seconds": {
                    "type": "integer"
                },
                "notes": {
                    "type": "string"
                },
                "started_at": {
                    "type": "string"
                },
                "subject_id": {
                    "type": "string"
                }
            }
        },
        "handler.SubjectResponse": {
            "type": "object",
            "properties": {
                "color_hex": {
                    "type": "string"
                },
                "created_at": {
                    "type": "string"
                },
                "deleted_at": {
                    "type": "string"
                },
                "id": {
                    "type": "string"
                },
                "name": {
                    "type": "string"
                },
                "updated_at": {
                    "type": "string"
                }
            }
        },
        "handler.TimeReportResponse": {
            "type": "object",
            "properties": {
                "color_hex": {
                    "type": "string"
                },
                "sessions_count": {
                    "type": "integer"
                },
                "subject_id": {
                    "type": "string"
                },
                "subject_name": {
                    "type": "string"
                },
                "total_hours_net": {
                    "type": "number"
                }
            }
        },
        "handler.TopicAccuracyResponse": {
            "type": "object",
            "properties": {
                "accuracy_percentage": {
                    "type": "number"
                },
                "topic_id": {
                    "type": "string"
                },
                "topic_name": {
                    "type": "string"
                },
                "total_correct": {
                    "type": "integer"
                },
                "total_questions": {
                    "type": "integer"
                }
            }
        },
        "handler.TopicResponse": {
            "type": "object",
            "properties": {
                "created_at": {
                    "type": "string"
                },
                "deleted_at": {
                    "type": "string"
                },
                "id": {
                    "type": "string"
                },
                "name": {
                    "type": "string"
                },
                "subject_id": {
                    "type": "string"
                },
                "updated_at": {
                    "type": "string"
                }
            }
        },
        "handler.UpdateCycleItemRequest": {
            "type": "object",
            "required": [
                "order_index",
                "subject_id"
            ],
            "properties": {
                "order_index": {
                    "type": "integer",
                    "minimum": 1
                },
                "planned_duration_minutes": {
                    "type": "integer",
                    "minimum": 1
                },
                "subject_id": {
                    "type": "string"
                }
            }
        },
        "handler.UpdateSessionDurationRequest": {
            "type": "object",
            "properties": {
                "finished_at": {
                    "type": "string"
                },
                "gross_duration_seconds": {
                    "type": "integer"
                },
                "net_duration_seconds": {
                    "type": "integer"
                },
                "notes": {
                    "type": "string"
                }
            }
        },
        "handler.UpdateStudyCycleRequest": {
            "type": "object",
            "required": [
                "name"
            ],
            "properties": {
                "description": {
                    "type": "string"
                },
                "is_active": {
                    "type": "boolean"
                },
                "name": {
                    "type": "string",
                    "minLength": 2
                }
            }
        },
        "handler.UpdateSubjectRequest": {
            "type": "object",
            "required": [
                "name"
            ],
            "properties": {
                "color_hex": {
                    "type": "string"
                },
                "name": {
                    "type": "string",
                    "minLength": 2
                }
            }
        },
        "handler.UpdateTopicRequest": {
            "type": "object",
            "required": [
                "name"
            ],
            "properties": {
                "name": {
                    "type": "string",
                    "minLength": 2
                }
            }
        }
    }
}
```

--- 

**File:** `docs/swagger.yaml`

```typescript
basePath: /
definitions:
  database.User:
    properties:
      created_at:
        type: string
      email:
        type: string
      id:
        type: string
      name:
        type: string
    type: object
  handler.AccuracyReportResponse:
    properties:
      accuracy_percentage:
        type: number
      color_hex:
        type: string
      subject_id:
        type: string
      subject_name:
        type: string
      total_correct:
        type: integer
      total_questions:
        type: integer
    type: object
  handler.CreateCycleItemRequest:
    properties:
      order_index:
        minimum: 1
        type: integer
      planned_duration_minutes:
        minimum: 1
        type: integer
      subject_id:
        type: string
    required:
    - order_index
    - subject_id
    type: object
  handler.CreateExerciseLogRequest:
    properties:
      correct_count:
        minimum: 0
        type: integer
      questions_count:
        minimum: 0
        type: integer
      session_id:
        type: string
      subject_id:
        type: string
      topic_id:
        type: string
    required:
    - correct_count
    - questions_count
    - subject_id
    type: object
  handler.CreateSessionPauseRequest:
    properties:
      session_id:
        type: string
      started_at:
        type: string
    required:
    - session_id
    - started_at
    type: object
  handler.CreateStudyCycleRequest:
    properties:
      description:
        type: string
      is_active:
        type: boolean
      name:
        minLength: 2
        type: string
    required:
    - name
    type: object
  handler.CreateStudySessionRequest:
    properties:
      cycle_item_id:
        type: string
      started_at:
        type: string
      subject_id:
        type: string
    required:
    - started_at
    - subject_id
    type: object
  handler.CreateSubjectRequest:
    properties:
      color_hex:
        type: string
      name:
        minLength: 2
        type: string
    required:
    - name
    type: object
  handler.CreateTopicRequest:
    properties:
      name:
        minLength: 2
        type: string
    required:
    - name
    type: object
  handler.CreateUserRequest:
    properties:
      email:
        description: |-
          required: cannot be empty
          email: must be a valid email format
        type: string
      name:
        description: 'min=2: must be at least 2 chars'
        minLength: 2
        type: string
    required:
    - email
    - name
    type: object
  handler.CycleItemResponse:
    properties:
      created_at:
        type: string
      cycle_id:
        type: string
      id:
        type: string
      order_index:
        type: integer
      planned_duration_minutes:
        type: integer
      subject_id:
        type: string
      updated_at:
        type: string
    type: object
  handler.CycleItemWithSubjectResponse:
    properties:
      color_hex:
        type: string
      cycle_item_id:
        type: string
      order_index:
        type: integer
      planned_duration_minutes:
        type: integer
      subject_id:
        type: string
      subject_name:
        type: string
    type: object
  handler.EndSessionPauseRequest:
    properties:
      ended_at:
        type: string
    required:
    - ended_at
    type: object
  handler.ExerciseLogResponse:
    properties:
      correct_count:
        type: integer
      created_at:
        type: string
      id:
        type: string
      questions_count:
        type: integer
      session_id:
        type: string
      subject_id:
        type: string
      topic_id:
        type: string
    type: object
  handler.HeatmapDayResponse:
    properties:
      sessions_count:
        type: integer
      study_date:
        type: string
      total_seconds:
        type: integer
    type: object
  handler.OpenSessionResponse:
    properties:
      color_hex:
        type: string
      cycle_item_id:
        type: string
      id:
        type: string
      started_at:
        type: string
      subject_id:
        type: string
      subject_name:
        type: string
    type: object
  handler.SessionPauseResponse:
    properties:
      ended_at:
        type: string
      id:
        type: string
      session_id:
        type: string
      started_at:
        type: string
    type: object
  handler.StudyCycleResponse:
    properties:
      created_at:
        type: string
      deleted_at:
        type: string
      description:
        type: string
      id:
        type: string
      is_active:
        type: integer
      name:
        type: string
      updated_at:
        type: string
    type: object
  handler.StudySessionResponse:
    properties:
      cycle_item_id:
        type: string
      finished_at:
        type: string
      gross_duration_seconds:
        type: integer
      id:
        type: string
      net_duration_seconds:
        type: integer
      notes:
        type: string
      started_at:
        type: string
      subject_id:
        type: string
    type: object
  handler.SubjectResponse:
    properties:
      color_hex:
        type: string
      created_at:
        type: string
      deleted_at:
        type: string
      id:
        type: string
      name:
        type: string
      updated_at:
        type: string
    type: object
  handler.TimeReportResponse:
    properties:
      color_hex:
        type: string
      sessions_count:
        type: integer
      subject_id:
        type: string
      subject_name:
        type: string
      total_hours_net:
        type: number
    type: object
  handler.TopicAccuracyResponse:
    properties:
      accuracy_percentage:
        type: number
      topic_id:
        type: string
      topic_name:
        type: string
      total_correct:
        type: integer
      total_questions:
        type: integer
    type: object
  handler.TopicResponse:
    properties:
      created_at:
        type: string
      deleted_at:
        type: string
      id:
        type: string
      name:
        type: string
      subject_id:
        type: string
      updated_at:
        type: string
    type: object
  handler.UpdateCycleItemRequest:
    properties:
      order_index:
        minimum: 1
        type: integer
      planned_duration_minutes:
        minimum: 1
        type: integer
      subject_id:
        type: string
    required:
    - order_index
    - subject_id
    type: object
  handler.UpdateSessionDurationRequest:
    properties:
      finished_at:
        type: string
      gross_duration_seconds:
        type: integer
      net_duration_seconds:
        type: integer
      notes:
        type: string
    type: object
  handler.UpdateStudyCycleRequest:
    properties:
      description:
        type: string
      is_active:
        type: boolean
      name:
        minLength: 2
        type: string
    required:
    - name
    type: object
  handler.UpdateSubjectRequest:
    properties:
      color_hex:
        type: string
      name:
        minLength: 2
        type: string
    required:
    - name
    type: object
  handler.UpdateTopicRequest:
    properties:
      name:
        minLength: 2
        type: string
    required:
    - name
    type: object
host: localhost:8080
info:
  contact: {}
  description: Production ready starter guide.
  title: My Go API
  version: "1.0"
paths:
  /analytics/accuracy-by-subject:
    get:
      produces:
      - application/json
      responses:
        "200":
          description: OK
          schema:
            items:
              $ref: '#/definitions/handler.AccuracyReportResponse'
            type: array
      summary: Get accuracy report by subject
      tags:
      - analytics
  /analytics/accuracy-by-topic/{subject_id}:
    get:
      parameters:
      - description: Subject ID
        in: path
        name: subject_id
        required: true
        type: string
      produces:
      - application/json
      responses:
        "200":
          description: OK
          schema:
            items:
              $ref: '#/definitions/handler.TopicAccuracyResponse'
            type: array
      summary: Get accuracy report by topic for a subject
      tags:
      - analytics
  /analytics/heatmap:
    get:
      parameters:
      - default: 30
        description: Number of days
        in: query
        name: days
        type: integer
      produces:
      - application/json
      responses:
        "200":
          description: OK
          schema:
            items:
              $ref: '#/definitions/handler.HeatmapDayResponse'
            type: array
      summary: Get activity heatmap data
      tags:
      - analytics
  /analytics/time-by-subject:
    get:
      parameters:
      - description: Start date (YYYY-MM-DD)
        in: query
        name: start_date
        type: string
      - description: End date (YYYY-MM-DD)
        in: query
        name: end_date
        type: string
      produces:
      - application/json
      responses:
        "200":
          description: OK
          schema:
            items:
              $ref: '#/definitions/handler.TimeReportResponse'
            type: array
      summary: Get time tracking report by subject
      tags:
      - analytics
  /cycle-items/{id}:
    delete:
      parameters:
      - description: Item ID
        in: path
        name: id
        required: true
        type: string
      responses:
        "204":
          description: No Content
      summary: Delete a cycle item
      tags:
      - cycle_items
    get:
      parameters:
      - description: Item ID
        in: path
        name: id
        required: true
        type: string
      produces:
      - application/json
      responses:
        "200":
          description: OK
          schema:
            $ref: '#/definitions/handler.CycleItemResponse'
      summary: Get a cycle item by ID
      tags:
      - cycle_items
    put:
      consumes:
      - application/json
      parameters:
      - description: Item ID
        in: path
        name: id
        required: true
        type: string
      - description: Cycle item info
        in: body
        name: input
        required: true
        schema:
          $ref: '#/definitions/handler.UpdateCycleItemRequest'
      produces:
      - application/json
      responses:
        "200":
          description: OK
          schema:
            type: string
      summary: Update a cycle item
      tags:
      - cycle_items
  /exercise-logs:
    post:
      consumes:
      - application/json
      parameters:
      - description: Exercise log info
        in: body
        name: input
        required: true
        schema:
          $ref: '#/definitions/handler.CreateExerciseLogRequest'
      produces:
      - application/json
      responses:
        "201":
          description: Created
          schema:
            $ref: '#/definitions/handler.ExerciseLogResponse'
      summary: Create a new exercise log
      tags:
      - exercise_logs
  /exercise-logs/{id}:
    delete:
      parameters:
      - description: Log ID
        in: path
        name: id
        required: true
        type: string
      responses:
        "204":
          description: No Content
      summary: Delete an exercise log
      tags:
      - exercise_logs
    get:
      parameters:
      - description: Log ID
        in: path
        name: id
        required: true
        type: string
      produces:
      - application/json
      responses:
        "200":
          description: OK
          schema:
            $ref: '#/definitions/handler.ExerciseLogResponse'
      summary: Get an exercise log by ID
      tags:
      - exercise_logs
  /session-pauses:
    post:
      consumes:
      - application/json
      parameters:
      - description: Session pause info
        in: body
        name: input
        required: true
        schema:
          $ref: '#/definitions/handler.CreateSessionPauseRequest'
      produces:
      - application/json
      responses:
        "201":
          description: Created
          schema:
            $ref: '#/definitions/handler.SessionPauseResponse'
      summary: Create a new session pause
      tags:
      - session_pauses
  /session-pauses/{id}:
    delete:
      parameters:
      - description: Pause ID
        in: path
        name: id
        required: true
        type: string
      responses:
        "204":
          description: No Content
      summary: Delete a session pause
      tags:
      - session_pauses
    get:
      parameters:
      - description: Pause ID
        in: path
        name: id
        required: true
        type: string
      produces:
      - application/json
      responses:
        "200":
          description: OK
          schema:
            $ref: '#/definitions/handler.SessionPauseResponse'
      summary: Get a session pause by ID
      tags:
      - session_pauses
  /session-pauses/{id}/end:
    put:
      consumes:
      - application/json
      parameters:
      - description: Pause ID
        in: path
        name: id
        required: true
        type: string
      - description: End pause info
        in: body
        name: input
        required: true
        schema:
          $ref: '#/definitions/handler.EndSessionPauseRequest'
      produces:
      - application/json
      responses:
        "200":
          description: OK
          schema:
            type: string
      summary: End a session pause
      tags:
      - session_pauses
  /study-cycles:
    post:
      consumes:
      - application/json
      parameters:
      - description: Study cycle info
        in: body
        name: input
        required: true
        schema:
          $ref: '#/definitions/handler.CreateStudyCycleRequest'
      produces:
      - application/json
      responses:
        "201":
          description: Created
          schema:
            $ref: '#/definitions/handler.StudyCycleResponse'
      summary: Create a new study cycle
      tags:
      - study_cycles
  /study-cycles/{id}:
    delete:
      parameters:
      - description: Cycle ID
        in: path
        name: id
        required: true
        type: string
      responses:
        "204":
          description: No Content
      summary: Delete a study cycle
      tags:
      - study_cycles
    get:
      parameters:
      - description: Cycle ID
        in: path
        name: id
        required: true
        type: string
      produces:
      - application/json
      responses:
        "200":
          description: OK
          schema:
            $ref: '#/definitions/handler.StudyCycleResponse'
      summary: Get a study cycle by ID
      tags:
      - study_cycles
    put:
      consumes:
      - application/json
      parameters:
      - description: Cycle ID
        in: path
        name: id
        required: true
        type: string
      - description: Study cycle info
        in: body
        name: input
        required: true
        schema:
          $ref: '#/definitions/handler.UpdateStudyCycleRequest'
      produces:
      - application/json
      responses:
        "200":
          description: OK
          schema:
            type: string
      summary: Update a study cycle
      tags:
      - study_cycles
  /study-cycles/{id}/items:
    get:
      parameters:
      - description: Cycle ID
        in: path
        name: id
        required: true
        type: string
      produces:
      - application/json
      responses:
        "200":
          description: OK
          schema:
            items:
              $ref: '#/definitions/handler.CycleItemResponse'
            type: array
      summary: List all items for a cycle
      tags:
      - cycle_items
    post:
      consumes:
      - application/json
      parameters:
      - description: Cycle ID
        in: path
        name: id
        required: true
        type: string
      - description: Cycle item info
        in: body
        name: input
        required: true
        schema:
          $ref: '#/definitions/handler.CreateCycleItemRequest'
      produces:
      - application/json
      responses:
        "201":
          description: Created
          schema:
            $ref: '#/definitions/handler.CycleItemResponse'
      summary: Create a new cycle item
      tags:
      - cycle_items
  /study-cycles/active:
    get:
      produces:
      - application/json
      responses:
        "200":
          description: OK
          schema:
            $ref: '#/definitions/handler.StudyCycleResponse'
      summary: Get the active study cycle
      tags:
      - study_cycles
  /study-cycles/active/items:
    get:
      produces:
      - application/json
      responses:
        "200":
          description: OK
          schema:
            items:
              $ref: '#/definitions/handler.CycleItemWithSubjectResponse'
            type: array
      summary: Get active cycle with all items (round-robin)
      tags:
      - study_cycles
  /study-sessions:
    post:
      consumes:
      - application/json
      parameters:
      - description: Study session info
        in: body
        name: input
        required: true
        schema:
          $ref: '#/definitions/handler.CreateStudySessionRequest'
      produces:
      - application/json
      responses:
        "201":
          description: Created
          schema:
            $ref: '#/definitions/handler.StudySessionResponse'
      summary: Create a new study session
      tags:
      - study_sessions
  /study-sessions/{id}:
    delete:
      parameters:
      - description: Session ID
        in: path
        name: id
        required: true
        type: string
      responses:
        "204":
          description: No Content
      summary: Delete a study session
      tags:
      - study_sessions
    get:
      parameters:
      - description: Session ID
        in: path
        name: id
        required: true
        type: string
      produces:
      - application/json
      responses:
        "200":
          description: OK
          schema:
            $ref: '#/definitions/handler.StudySessionResponse'
      summary: Get a study session by ID
      tags:
      - study_sessions
    put:
      consumes:
      - application/json
      parameters:
      - description: Session ID
        in: path
        name: id
        required: true
        type: string
      - description: Session duration info
        in: body
        name: input
        required: true
        schema:
          $ref: '#/definitions/handler.UpdateSessionDurationRequest'
      produces:
      - application/json
      responses:
        "200":
          description: OK
          schema:
            type: string
      summary: Update study session duration
      tags:
      - study_sessions
  /study-sessions/open:
    get:
      produces:
      - application/json
      responses:
        "200":
          description: OK
          schema:
            $ref: '#/definitions/handler.OpenSessionResponse'
      summary: Get open/unfinished session (crash recovery)
      tags:
      - study_sessions
  /subjects:
    get:
      produces:
      - application/json
      responses:
        "200":
          description: OK
          schema:
            items:
              $ref: '#/definitions/handler.SubjectResponse'
            type: array
      summary: List all subjects
      tags:
      - subjects
    post:
      consumes:
      - application/json
      parameters:
      - description: Subject info
        in: body
        name: input
        required: true
        schema:
          $ref: '#/definitions/handler.CreateSubjectRequest'
      produces:
      - application/json
      responses:
        "201":
          description: Created
          schema:
            $ref: '#/definitions/handler.SubjectResponse'
      summary: Create a new subject
      tags:
      - subjects
  /subjects/{id}:
    delete:
      parameters:
      - description: Subject ID
        in: path
        name: id
        required: true
        type: string
      responses:
        "204":
          description: No Content
      summary: Delete a subject
      tags:
      - subjects
    get:
      parameters:
      - description: Subject ID
        in: path
        name: id
        required: true
        type: string
      produces:
      - application/json
      responses:
        "200":
          description: OK
          schema:
            $ref: '#/definitions/handler.SubjectResponse'
      summary: Get a subject by ID
      tags:
      - subjects
    put:
      consumes:
      - application/json
      parameters:
      - description: Subject ID
        in: path
        name: id
        required: true
        type: string
      - description: Subject info
        in: body
        name: input
        required: true
        schema:
          $ref: '#/definitions/handler.UpdateSubjectRequest'
      produces:
      - application/json
      responses:
        "200":
          description: OK
          schema:
            type: string
      summary: Update a subject
      tags:
      - subjects
  /subjects/{id}/topics:
    get:
      parameters:
      - description: Subject ID
        in: path
        name: id
        required: true
        type: string
      produces:
      - application/json
      responses:
        "200":
          description: OK
          schema:
            items:
              $ref: '#/definitions/handler.TopicResponse'
            type: array
      summary: List all topics for a subject
      tags:
      - topics
    post:
      consumes:
      - application/json
      parameters:
      - description: Subject ID
        in: path
        name: id
        required: true
        type: string
      - description: Topic info
        in: body
        name: input
        required: true
        schema:
          $ref: '#/definitions/handler.CreateTopicRequest'
      produces:
      - application/json
      responses:
        "201":
          description: Created
          schema:
            $ref: '#/definitions/handler.TopicResponse'
      summary: Create a new topic for a subject
      tags:
      - topics
  /topics/{id}:
    delete:
      parameters:
      - description: Topic ID
        in: path
        name: id
        required: true
        type: string
      responses:
        "204":
          description: No Content
      summary: Delete a topic
      tags:
      - topics
    get:
      parameters:
      - description: Topic ID
        in: path
        name: id
        required: true
        type: string
      produces:
      - application/json
      responses:
        "200":
          description: OK
          schema:
            $ref: '#/definitions/handler.TopicResponse'
      summary: Get a topic by ID
      tags:
      - topics
    put:
      consumes:
      - application/json
      parameters:
      - description: Topic ID
        in: path
        name: id
        required: true
        type: string
      - description: Topic info
        in: body
        name: input
        required: true
        schema:
          $ref: '#/definitions/handler.UpdateTopicRequest'
      produces:
      - application/json
      responses:
        "200":
          description: OK
          schema:
            type: string
      summary: Update a topic
      tags:
      - topics
  /users:
    post:
      consumes:
      - application/json
      description: Create a user with email and name
      parameters:
      - description: User info
        in: body
        name: input
        required: true
        schema:
          $ref: '#/definitions/handler.CreateUserRequest'
      produces:
      - application/json
      responses:
        "201":
          description: Created
          schema:
            $ref: '#/definitions/database.User'
        "400":
          description: Invalid request
          schema:
            type: string
        "500":
          description: Internal server error
          schema:
            type: string
      summary: Create a new user
      tags:
      - users
  /users/{email}:
    get:
      parameters:
      - description: User Email
        in: path
        name: email
        required: true
        type: string
      responses:
        "200":
          description: OK
          schema:
            $ref: '#/definitions/database.User'
        "404":
          description: User not found
          schema:
            type: string
      summary: Get user by Email
      tags:
      - users
swagger: "2.0"

```

--- 

**File:** `internal/config/config.go`

```typescript
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Port    string
	DBUrl   string
	Env     string
	Timeout time.Duration
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	cfg := &Config{
		Port:    getEnv("PORT", "8080"),
		DBUrl:   getEnv("DB_URL", ""),
		Env:     getEnv("ENV", "development"),
		Timeout: 5 * time.Second,
	}

	if cfg.DBUrl == "" {
		if cfg.Env == "development" {
			// 1. Get Absolute Path
			wd, _ := os.Getwd()
			rawPath := filepath.Join(wd, "dev.db")

			// 2. Force Forward Slashes (Windows "B:\" breaks URI parsing, "B:/" works)
			cleanPath := filepath.ToSlash(rawPath)

			// 3. Construct "Compatibility Mode" DSN
			// file: prefix is required for parameters to work
			// _pragma=journal_mode(DELETE): No WAL/SHM files
			// _pragma=temp_store(MEMORY): No temp files on disk
			// _pragma=mmap_size(0): No memory mapping (fixes "Out of Memory" on some drives)
			cfg.DBUrl = fmt.Sprintf("file:%s?_pragma=journal_mode(DELETE)&_pragma=temp_store(MEMORY)&_pragma=mmap_size(0)", cleanPath)
		} else {
			return nil, fmt.Errorf("DB_URL environment variable is required")
		}
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

```

--- 

**File:** `internal/database/analytics.sql.go`

```typescript
// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.30.0
// source: analytics.sql

package database

import (
	"context"
	"database/sql"
)

const getAccuracyBySubject = `-- name: GetAccuracyBySubject :many
SELECT 
    s.id AS subject_id,
    s.name AS subject_name,
    s.color_hex,
    SUM(el.questions_count) AS total_questions,
    SUM(el.correct_count) AS total_correct,
    ROUND(
        (SUM(el.correct_count) * 100.0) / NULLIF(SUM(el.questions_count), 0), 
        2
    ) AS accuracy_percentage
FROM subjects s
LEFT JOIN exercise_logs el ON s.id = el.subject_id
WHERE s.deleted_at IS NULL
GROUP BY s.id, s.name, s.color_hex
HAVING total_questions > 0
ORDER BY accuracy_percentage ASC
`

type GetAccuracyBySubjectRow struct {
	SubjectID          string          `json:"subject_id"`
	SubjectName        string          `json:"subject_name"`
	ColorHex           sql.NullString  `json:"color_hex"`
	TotalQuestions     sql.NullFloat64 `json:"total_questions"`
	TotalCorrect       sql.NullFloat64 `json:"total_correct"`
	AccuracyPercentage float64         `json:"accuracy_percentage"`
}

func (q *Queries) GetAccuracyBySubject(ctx context.Context) ([]GetAccuracyBySubjectRow, error) {
	rows, err := q.db.QueryContext(ctx, getAccuracyBySubject)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetAccuracyBySubjectRow
	for rows.Next() {
		var i GetAccuracyBySubjectRow
		if err := rows.Scan(
			&i.SubjectID,
			&i.SubjectName,
			&i.ColorHex,
			&i.TotalQuestions,
			&i.TotalCorrect,
			&i.AccuracyPercentage,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getAccuracyByTopic = `-- name: GetAccuracyByTopic :many
SELECT 
    t.id AS topic_id,
    t.name AS topic_name,
    SUM(el.questions_count) AS total_questions,
    SUM(el.correct_count) AS total_correct,
    ROUND(
        (SUM(el.correct_count) * 100.0) / NULLIF(SUM(el.questions_count), 0), 
        2
    ) AS accuracy_percentage
FROM topics t
LEFT JOIN exercise_logs el ON t.id = el.topic_id
WHERE t.subject_id = ?
  AND t.deleted_at IS NULL
GROUP BY t.id, t.name
HAVING total_questions > 0
ORDER BY accuracy_percentage ASC
`

type GetAccuracyByTopicRow struct {
	TopicID            string          `json:"topic_id"`
	TopicName          string          `json:"topic_name"`
	TotalQuestions     sql.NullFloat64 `json:"total_questions"`
	TotalCorrect       sql.NullFloat64 `json:"total_correct"`
	AccuracyPercentage float64         `json:"accuracy_percentage"`
}

func (q *Queries) GetAccuracyByTopic(ctx context.Context, subjectID string) ([]GetAccuracyByTopicRow, error) {
	rows, err := q.db.QueryContext(ctx, getAccuracyByTopic, subjectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetAccuracyByTopicRow
	for rows.Next() {
		var i GetAccuracyByTopicRow
		if err := rows.Scan(
			&i.TopicID,
			&i.TopicName,
			&i.TotalQuestions,
			&i.TotalCorrect,
			&i.AccuracyPercentage,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getActivityHeatmap = `-- name: GetActivityHeatmap :many
SELECT 
    strftime('%Y-%m-%d', started_at) AS study_date,
    COUNT(DISTINCT id) AS sessions_count,
    COALESCE(SUM(net_duration_seconds), 0) AS total_seconds
FROM study_sessions
WHERE finished_at IS NOT NULL
  AND datetime(started_at) >= datetime('now', '-' || CAST(?1 AS TEXT) || ' days')
GROUP BY study_date
ORDER BY study_date DESC
`

type GetActivityHeatmapRow struct {
	StudyDate     interface{} `json:"study_date"`
	SessionsCount int64       `json:"sessions_count"`
	TotalSeconds  interface{} `json:"total_seconds"`
}

func (q *Queries) GetActivityHeatmap(ctx context.Context, daysCount string) ([]GetActivityHeatmapRow, error) {
	rows, err := q.db.QueryContext(ctx, getActivityHeatmap, daysCount)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetActivityHeatmapRow
	for rows.Next() {
		var i GetActivityHeatmapRow
		if err := rows.Scan(&i.StudyDate, &i.SessionsCount, &i.TotalSeconds); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getTimeReportBySubject = `-- name: GetTimeReportBySubject :many

SELECT 
    s.id AS subject_id,
    s.name AS subject_name,
    s.color_hex,
    COUNT(ss.id) AS sessions_count,
    ROUND(COALESCE(SUM(ss.net_duration_seconds), 0) / 3600.0, 2) AS total_hours_net
FROM subjects s
LEFT JOIN study_sessions ss ON s.id = ss.subject_id 
    AND ss.finished_at IS NOT NULL
    AND (?1 = '' OR ss.started_at >= ?1)
    AND (?2 = '' OR ss.started_at <= ?2)
WHERE s.deleted_at IS NULL
GROUP BY s.id, s.name, s.color_hex
HAVING sessions_count > 0
ORDER BY total_hours_net DESC
`

type GetTimeReportBySubjectParams struct {
	StartDateFrom interface{} `json:"start_date_from"`
	StartDateTo   interface{} `json:"start_date_to"`
}

type GetTimeReportBySubjectRow struct {
	SubjectID     string         `json:"subject_id"`
	SubjectName   string         `json:"subject_name"`
	ColorHex      sql.NullString `json:"color_hex"`
	SessionsCount int64          `json:"sessions_count"`
	TotalHoursNet float64        `json:"total_hours_net"`
}

// Analytics Queries for Study App
func (q *Queries) GetTimeReportBySubject(ctx context.Context, arg GetTimeReportBySubjectParams) ([]GetTimeReportBySubjectRow, error) {
	rows, err := q.db.QueryContext(ctx, getTimeReportBySubject, arg.StartDateFrom, arg.StartDateTo)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetTimeReportBySubjectRow
	for rows.Next() {
		var i GetTimeReportBySubjectRow
		if err := rows.Scan(
			&i.SubjectID,
			&i.SubjectName,
			&i.ColorHex,
			&i.SessionsCount,
			&i.TotalHoursNet,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

```

--- 

**File:** `internal/database/cycle_items.sql.go`

```typescript
// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.30.0
// source: cycle_items.sql

package database

import (
	"context"
	"database/sql"
)

const createCycleItem = `-- name: CreateCycleItem :one
INSERT INTO cycle_items (id, cycle_id, subject_id, order_index, planned_duration_minutes)
VALUES (?, ?, ?, ?, ?)
RETURNING id, cycle_id, subject_id, order_index, planned_duration_minutes, created_at, updated_at
`

type CreateCycleItemParams struct {
	ID                     string        `json:"id"`
	CycleID                string        `json:"cycle_id"`
	SubjectID              string        `json:"subject_id"`
	OrderIndex             int64         `json:"order_index"`
	PlannedDurationMinutes sql.NullInt64 `json:"planned_duration_minutes"`
}

func (q *Queries) CreateCycleItem(ctx context.Context, arg CreateCycleItemParams) (CycleItem, error) {
	row := q.db.QueryRowContext(ctx, createCycleItem,
		arg.ID,
		arg.CycleID,
		arg.SubjectID,
		arg.OrderIndex,
		arg.PlannedDurationMinutes,
	)
	var i CycleItem
	err := row.Scan(
		&i.ID,
		&i.CycleID,
		&i.SubjectID,
		&i.OrderIndex,
		&i.PlannedDurationMinutes,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const deleteCycleItem = `-- name: DeleteCycleItem :exec
DELETE FROM cycle_items
WHERE id = ?
`

func (q *Queries) DeleteCycleItem(ctx context.Context, id string) error {
	_, err := q.db.ExecContext(ctx, deleteCycleItem, id)
	return err
}

const getCycleItem = `-- name: GetCycleItem :one
SELECT id, cycle_id, subject_id, order_index, planned_duration_minutes, created_at, updated_at FROM cycle_items
WHERE id = ?
`

func (q *Queries) GetCycleItem(ctx context.Context, id string) (CycleItem, error) {
	row := q.db.QueryRowContext(ctx, getCycleItem, id)
	var i CycleItem
	err := row.Scan(
		&i.ID,
		&i.CycleID,
		&i.SubjectID,
		&i.OrderIndex,
		&i.PlannedDurationMinutes,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const listCycleItems = `-- name: ListCycleItems :many
SELECT id, cycle_id, subject_id, order_index, planned_duration_minutes, created_at, updated_at FROM cycle_items
WHERE cycle_id = ?
ORDER BY order_index
`

func (q *Queries) ListCycleItems(ctx context.Context, cycleID string) ([]CycleItem, error) {
	rows, err := q.db.QueryContext(ctx, listCycleItems, cycleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []CycleItem
	for rows.Next() {
		var i CycleItem
		if err := rows.Scan(
			&i.ID,
			&i.CycleID,
			&i.SubjectID,
			&i.OrderIndex,
			&i.PlannedDurationMinutes,
			&i.CreatedAt,
			&i.UpdatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const updateCycleItem = `-- name: UpdateCycleItem :exec
UPDATE cycle_items
SET subject_id = ?, order_index = ?, planned_duration_minutes = ?, updated_at = datetime('now')
WHERE id = ?
`

type UpdateCycleItemParams struct {
	SubjectID              string        `json:"subject_id"`
	OrderIndex             int64         `json:"order_index"`
	PlannedDurationMinutes sql.NullInt64 `json:"planned_duration_minutes"`
	ID                     string        `json:"id"`
}

func (q *Queries) UpdateCycleItem(ctx context.Context, arg UpdateCycleItemParams) error {
	_, err := q.db.ExecContext(ctx, updateCycleItem,
		arg.SubjectID,
		arg.OrderIndex,
		arg.PlannedDurationMinutes,
		arg.ID,
	)
	return err
}

```

--- 

**File:** `internal/database/db.go`

```typescript
// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.30.0

package database

import (
	"context"
	"database/sql"
)

type DBTX interface {
	ExecContext(context.Context, string, ...interface{}) (sql.Result, error)
	PrepareContext(context.Context, string) (*sql.Stmt, error)
	QueryContext(context.Context, string, ...interface{}) (*sql.Rows, error)
	QueryRowContext(context.Context, string, ...interface{}) *sql.Row
}

func New(db DBTX) *Queries {
	return &Queries{db: db}
}

type Queries struct {
	db DBTX
}

func (q *Queries) WithTx(tx *sql.Tx) *Queries {
	return &Queries{
		db: tx,
	}
}

```

--- 

**File:** `internal/database/exercise_logs.sql.go`

```typescript
// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.30.0
// source: exercise_logs.sql

package database

import (
	"context"
	"database/sql"
)

const createExerciseLog = `-- name: CreateExerciseLog :one
INSERT INTO exercise_logs (id, session_id, subject_id, topic_id, questions_count, correct_count)
VALUES (?, ?, ?, ?, ?, ?)
RETURNING id, session_id, subject_id, topic_id, questions_count, correct_count, created_at
`

type CreateExerciseLogParams struct {
	ID             string         `json:"id"`
	SessionID      sql.NullString `json:"session_id"`
	SubjectID      string         `json:"subject_id"`
	TopicID        sql.NullString `json:"topic_id"`
	QuestionsCount int64          `json:"questions_count"`
	CorrectCount   int64          `json:"correct_count"`
}

func (q *Queries) CreateExerciseLog(ctx context.Context, arg CreateExerciseLogParams) (ExerciseLog, error) {
	row := q.db.QueryRowContext(ctx, createExerciseLog,
		arg.ID,
		arg.SessionID,
		arg.SubjectID,
		arg.TopicID,
		arg.QuestionsCount,
		arg.CorrectCount,
	)
	var i ExerciseLog
	err := row.Scan(
		&i.ID,
		&i.SessionID,
		&i.SubjectID,
		&i.TopicID,
		&i.QuestionsCount,
		&i.CorrectCount,
		&i.CreatedAt,
	)
	return i, err
}

const deleteExerciseLog = `-- name: DeleteExerciseLog :exec
DELETE FROM exercise_logs
WHERE id = ?
`

func (q *Queries) DeleteExerciseLog(ctx context.Context, id string) error {
	_, err := q.db.ExecContext(ctx, deleteExerciseLog, id)
	return err
}

const getExerciseLog = `-- name: GetExerciseLog :one
SELECT id, session_id, subject_id, topic_id, questions_count, correct_count, created_at FROM exercise_logs
WHERE id = ?
`

func (q *Queries) GetExerciseLog(ctx context.Context, id string) (ExerciseLog, error) {
	row := q.db.QueryRowContext(ctx, getExerciseLog, id)
	var i ExerciseLog
	err := row.Scan(
		&i.ID,
		&i.SessionID,
		&i.SubjectID,
		&i.TopicID,
		&i.QuestionsCount,
		&i.CorrectCount,
		&i.CreatedAt,
	)
	return i, err
}

```

--- 

**File:** `internal/database/models.go`

```typescript
// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.30.0

package database

import (
	"database/sql"
	"time"
)

type CycleItem struct {
	ID                     string        `json:"id"`
	CycleID                string        `json:"cycle_id"`
	SubjectID              string        `json:"subject_id"`
	OrderIndex             int64         `json:"order_index"`
	PlannedDurationMinutes sql.NullInt64 `json:"planned_duration_minutes"`
	CreatedAt              string        `json:"created_at"`
	UpdatedAt              string        `json:"updated_at"`
}

type ExerciseLog struct {
	ID             string         `json:"id"`
	SessionID      sql.NullString `json:"session_id"`
	SubjectID      string         `json:"subject_id"`
	TopicID        sql.NullString `json:"topic_id"`
	QuestionsCount int64          `json:"questions_count"`
	CorrectCount   int64          `json:"correct_count"`
	CreatedAt      string         `json:"created_at"`
}

type SessionPause struct {
	ID              string         `json:"id"`
	SessionID       string         `json:"session_id"`
	StartedAt       string         `json:"started_at"`
	EndedAt         sql.NullString `json:"ended_at"`
	DurationSeconds sql.NullInt64  `json:"duration_seconds"`
}

type StudyCycle struct {
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	Description sql.NullString `json:"description"`
	IsActive    sql.NullInt64  `json:"is_active"`
	CreatedAt   string         `json:"created_at"`
	UpdatedAt   string         `json:"updated_at"`
	DeletedAt   sql.NullString `json:"deleted_at"`
}

type StudySession struct {
	ID                   string         `json:"id"`
	SubjectID            string         `json:"subject_id"`
	CycleItemID          sql.NullString `json:"cycle_item_id"`
	StartedAt            string         `json:"started_at"`
	FinishedAt           sql.NullString `json:"finished_at"`
	GrossDurationSeconds sql.NullInt64  `json:"gross_duration_seconds"`
	NetDurationSeconds   sql.NullInt64  `json:"net_duration_seconds"`
	Notes                sql.NullString `json:"notes"`
	CreatedAt            string         `json:"created_at"`
	UpdatedAt            string         `json:"updated_at"`
}

type Subject struct {
	ID        string         `json:"id"`
	Name      string         `json:"name"`
	ColorHex  sql.NullString `json:"color_hex"`
	CreatedAt string         `json:"created_at"`
	UpdatedAt string         `json:"updated_at"`
	DeletedAt sql.NullString `json:"deleted_at"`
}

type Topic struct {
	ID        string         `json:"id"`
	SubjectID string         `json:"subject_id"`
	Name      string         `json:"name"`
	CreatedAt string         `json:"created_at"`
	UpdatedAt string         `json:"updated_at"`
	DeletedAt sql.NullString `json:"deleted_at"`
}

type User struct {
	ID           string    `json:"id"`
	Email        string    `json:"email"`
	Name         string    `json:"name"`
	CreatedAt    time.Time `json:"created_at"`
	PasswordHash string    `json:"password_hash"`
}

```

--- 

**File:** `internal/database/querier.go`

```typescript
// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.30.0

package database

import (
	"context"
)

type Querier interface {
	CreateCycleItem(ctx context.Context, arg CreateCycleItemParams) (CycleItem, error)
	CreateExerciseLog(ctx context.Context, arg CreateExerciseLogParams) (ExerciseLog, error)
	CreateSessionPause(ctx context.Context, arg CreateSessionPauseParams) (SessionPause, error)
	CreateStudyCycle(ctx context.Context, arg CreateStudyCycleParams) (StudyCycle, error)
	CreateStudySession(ctx context.Context, arg CreateStudySessionParams) (StudySession, error)
	CreateSubject(ctx context.Context, arg CreateSubjectParams) (Subject, error)
	CreateTopic(ctx context.Context, arg CreateTopicParams) (Topic, error)
	CreateUser(ctx context.Context, arg CreateUserParams) (User, error)
	DeleteCycleItem(ctx context.Context, id string) error
	DeleteExerciseLog(ctx context.Context, id string) error
	DeleteSessionPause(ctx context.Context, id string) error
	DeleteStudyCycle(ctx context.Context, id string) error
	DeleteStudySession(ctx context.Context, id string) error
	DeleteSubject(ctx context.Context, id string) error
	DeleteTopic(ctx context.Context, id string) error
	EndSessionPause(ctx context.Context, arg EndSessionPauseParams) error
	GetAccuracyBySubject(ctx context.Context) ([]GetAccuracyBySubjectRow, error)
	GetAccuracyByTopic(ctx context.Context, subjectID string) ([]GetAccuracyByTopicRow, error)
	GetActiveCycleWithItems(ctx context.Context) ([]GetActiveCycleWithItemsRow, error)
	GetActiveStudyCycle(ctx context.Context) (StudyCycle, error)
	GetActivityHeatmap(ctx context.Context, daysCount string) ([]GetActivityHeatmapRow, error)
	GetCycleItem(ctx context.Context, id string) (CycleItem, error)
	GetExerciseLog(ctx context.Context, id string) (ExerciseLog, error)
	GetOpenSession(ctx context.Context) (GetOpenSessionRow, error)
	GetSessionPause(ctx context.Context, id string) (SessionPause, error)
	GetStudyCycle(ctx context.Context, id string) (StudyCycle, error)
	GetStudySession(ctx context.Context, id string) (StudySession, error)
	GetSubject(ctx context.Context, id string) (Subject, error)
	// Analytics Queries for Study App
	GetTimeReportBySubject(ctx context.Context, arg GetTimeReportBySubjectParams) ([]GetTimeReportBySubjectRow, error)
	GetTopic(ctx context.Context, id string) (Topic, error)
	GetUserByEmail(ctx context.Context, email string) (User, error)
	ListCycleItems(ctx context.Context, cycleID string) ([]CycleItem, error)
	ListSubjects(ctx context.Context) ([]Subject, error)
	ListTopicsBySubject(ctx context.Context, subjectID string) ([]Topic, error)
	UpdateCycleItem(ctx context.Context, arg UpdateCycleItemParams) error
	UpdateSessionDuration(ctx context.Context, arg UpdateSessionDurationParams) error
	UpdateStudyCycle(ctx context.Context, arg UpdateStudyCycleParams) error
	UpdateSubject(ctx context.Context, arg UpdateSubjectParams) error
	UpdateTopic(ctx context.Context, arg UpdateTopicParams) error
}

var _ Querier = (*Queries)(nil)

```

--- 

**File:** `internal/database/session_pauses.sql.go`

```typescript
// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.30.0
// source: session_pauses.sql

package database

import (
	"context"
	"database/sql"
)

const createSessionPause = `-- name: CreateSessionPause :one
INSERT INTO session_pauses (id, session_id, started_at)
VALUES (?, ?, ?)
RETURNING id, session_id, started_at, ended_at, duration_seconds
`

type CreateSessionPauseParams struct {
	ID        string `json:"id"`
	SessionID string `json:"session_id"`
	StartedAt string `json:"started_at"`
}

func (q *Queries) CreateSessionPause(ctx context.Context, arg CreateSessionPauseParams) (SessionPause, error) {
	row := q.db.QueryRowContext(ctx, createSessionPause, arg.ID, arg.SessionID, arg.StartedAt)
	var i SessionPause
	err := row.Scan(
		&i.ID,
		&i.SessionID,
		&i.StartedAt,
		&i.EndedAt,
		&i.DurationSeconds,
	)
	return i, err
}

const deleteSessionPause = `-- name: DeleteSessionPause :exec
DELETE FROM session_pauses
WHERE id = ?
`

func (q *Queries) DeleteSessionPause(ctx context.Context, id string) error {
	_, err := q.db.ExecContext(ctx, deleteSessionPause, id)
	return err
}

const endSessionPause = `-- name: EndSessionPause :exec
UPDATE session_pauses
SET ended_at = ?
WHERE id = ?
`

type EndSessionPauseParams struct {
	EndedAt sql.NullString `json:"ended_at"`
	ID      string         `json:"id"`
}

func (q *Queries) EndSessionPause(ctx context.Context, arg EndSessionPauseParams) error {
	_, err := q.db.ExecContext(ctx, endSessionPause, arg.EndedAt, arg.ID)
	return err
}

const getSessionPause = `-- name: GetSessionPause :one
SELECT id, session_id, started_at, ended_at, duration_seconds FROM session_pauses
WHERE id = ?
`

func (q *Queries) GetSessionPause(ctx context.Context, id string) (SessionPause, error) {
	row := q.db.QueryRowContext(ctx, getSessionPause, id)
	var i SessionPause
	err := row.Scan(
		&i.ID,
		&i.SessionID,
		&i.StartedAt,
		&i.EndedAt,
		&i.DurationSeconds,
	)
	return i, err
}

```

--- 

**File:** `internal/database/study_cycles.sql.go`

```typescript
// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.30.0
// source: study_cycles.sql

package database

import (
	"context"
	"database/sql"
)

const createStudyCycle = `-- name: CreateStudyCycle :one
INSERT INTO study_cycles (id, name, description, is_active)
VALUES (?, ?, ?, ?)
RETURNING id, name, description, is_active, created_at, updated_at, deleted_at
`

type CreateStudyCycleParams struct {
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	Description sql.NullString `json:"description"`
	IsActive    sql.NullInt64  `json:"is_active"`
}

func (q *Queries) CreateStudyCycle(ctx context.Context, arg CreateStudyCycleParams) (StudyCycle, error) {
	row := q.db.QueryRowContext(ctx, createStudyCycle,
		arg.ID,
		arg.Name,
		arg.Description,
		arg.IsActive,
	)
	var i StudyCycle
	err := row.Scan(
		&i.ID,
		&i.Name,
		&i.Description,
		&i.IsActive,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.DeletedAt,
	)
	return i, err
}

const deleteStudyCycle = `-- name: DeleteStudyCycle :exec
UPDATE study_cycles
SET deleted_at = datetime('now')
WHERE id = ?
`

func (q *Queries) DeleteStudyCycle(ctx context.Context, id string) error {
	_, err := q.db.ExecContext(ctx, deleteStudyCycle, id)
	return err
}

const getActiveCycleWithItems = `-- name: GetActiveCycleWithItems :many
SELECT 
    ci.id AS cycle_item_id,
    ci.order_index,
    ci.planned_duration_minutes,
    s.id AS subject_id,
    s.name AS subject_name,
    s.color_hex
FROM cycle_items ci
JOIN study_cycles sc ON ci.cycle_id = sc.id
JOIN subjects s ON ci.subject_id = s.id
WHERE sc.is_active = 1 
  AND sc.deleted_at IS NULL
ORDER BY ci.order_index ASC
`

type GetActiveCycleWithItemsRow struct {
	CycleItemID            string         `json:"cycle_item_id"`
	OrderIndex             int64          `json:"order_index"`
	PlannedDurationMinutes sql.NullInt64  `json:"planned_duration_minutes"`
	SubjectID              string         `json:"subject_id"`
	SubjectName            string         `json:"subject_name"`
	ColorHex               sql.NullString `json:"color_hex"`
}

func (q *Queries) GetActiveCycleWithItems(ctx context.Context) ([]GetActiveCycleWithItemsRow, error) {
	rows, err := q.db.QueryContext(ctx, getActiveCycleWithItems)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetActiveCycleWithItemsRow
	for rows.Next() {
		var i GetActiveCycleWithItemsRow
		if err := rows.Scan(
			&i.CycleItemID,
			&i.OrderIndex,
			&i.PlannedDurationMinutes,
			&i.SubjectID,
			&i.SubjectName,
			&i.ColorHex,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getActiveStudyCycle = `-- name: GetActiveStudyCycle :one
SELECT id, name, description, is_active, created_at, updated_at, deleted_at FROM study_cycles
WHERE is_active = 1 AND deleted_at IS NULL
LIMIT 1
`

func (q *Queries) GetActiveStudyCycle(ctx context.Context) (StudyCycle, error) {
	row := q.db.QueryRowContext(ctx, getActiveStudyCycle)
	var i StudyCycle
	err := row.Scan(
		&i.ID,
		&i.Name,
		&i.Description,
		&i.IsActive,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.DeletedAt,
	)
	return i, err
}

const getStudyCycle = `-- name: GetStudyCycle :one
SELECT id, name, description, is_active, created_at, updated_at, deleted_at FROM study_cycles
WHERE id = ? AND deleted_at IS NULL
`

func (q *Queries) GetStudyCycle(ctx context.Context, id string) (StudyCycle, error) {
	row := q.db.QueryRowContext(ctx, getStudyCycle, id)
	var i StudyCycle
	err := row.Scan(
		&i.ID,
		&i.Name,
		&i.Description,
		&i.IsActive,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.DeletedAt,
	)
	return i, err
}

const updateStudyCycle = `-- name: UpdateStudyCycle :exec
UPDATE study_cycles
SET name = ?, description = ?, is_active = ?, updated_at = datetime('now')
WHERE id = ? AND deleted_at IS NULL
`

type UpdateStudyCycleParams struct {
	Name        string         `json:"name"`
	Description sql.NullString `json:"description"`
	IsActive    sql.NullInt64  `json:"is_active"`
	ID          string         `json:"id"`
}

func (q *Queries) UpdateStudyCycle(ctx context.Context, arg UpdateStudyCycleParams) error {
	_, err := q.db.ExecContext(ctx, updateStudyCycle,
		arg.Name,
		arg.Description,
		arg.IsActive,
		arg.ID,
	)
	return err
}

```

--- 

**File:** `internal/database/study_sessions.sql.go`

```typescript
// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.30.0
// source: study_sessions.sql

package database

import (
	"context"
	"database/sql"
)

const createStudySession = `-- name: CreateStudySession :one
INSERT INTO study_sessions (id, subject_id, cycle_item_id, started_at)
VALUES (?, ?, ?, ?)
RETURNING id, subject_id, cycle_item_id, started_at, finished_at, gross_duration_seconds, net_duration_seconds, notes, created_at, updated_at
`

type CreateStudySessionParams struct {
	ID          string         `json:"id"`
	SubjectID   string         `json:"subject_id"`
	CycleItemID sql.NullString `json:"cycle_item_id"`
	StartedAt   string         `json:"started_at"`
}

func (q *Queries) CreateStudySession(ctx context.Context, arg CreateStudySessionParams) (StudySession, error) {
	row := q.db.QueryRowContext(ctx, createStudySession,
		arg.ID,
		arg.SubjectID,
		arg.CycleItemID,
		arg.StartedAt,
	)
	var i StudySession
	err := row.Scan(
		&i.ID,
		&i.SubjectID,
		&i.CycleItemID,
		&i.StartedAt,
		&i.FinishedAt,
		&i.GrossDurationSeconds,
		&i.NetDurationSeconds,
		&i.Notes,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const deleteStudySession = `-- name: DeleteStudySession :exec
DELETE FROM study_sessions
WHERE id = ?
`

func (q *Queries) DeleteStudySession(ctx context.Context, id string) error {
	_, err := q.db.ExecContext(ctx, deleteStudySession, id)
	return err
}

const getOpenSession = `-- name: GetOpenSession :one
SELECT 
    ss.id,
    ss.subject_id,
    ss.cycle_item_id,
    ss.started_at,
    s.name AS subject_name,
    s.color_hex
FROM study_sessions ss
JOIN subjects s ON ss.subject_id = s.id
WHERE ss.finished_at IS NULL
ORDER BY ss.started_at DESC
LIMIT 1
`

type GetOpenSessionRow struct {
	ID          string         `json:"id"`
	SubjectID   string         `json:"subject_id"`
	CycleItemID sql.NullString `json:"cycle_item_id"`
	StartedAt   string         `json:"started_at"`
	SubjectName string         `json:"subject_name"`
	ColorHex    sql.NullString `json:"color_hex"`
}

func (q *Queries) GetOpenSession(ctx context.Context) (GetOpenSessionRow, error) {
	row := q.db.QueryRowContext(ctx, getOpenSession)
	var i GetOpenSessionRow
	err := row.Scan(
		&i.ID,
		&i.SubjectID,
		&i.CycleItemID,
		&i.StartedAt,
		&i.SubjectName,
		&i.ColorHex,
	)
	return i, err
}

const getStudySession = `-- name: GetStudySession :one
SELECT id, subject_id, cycle_item_id, started_at, finished_at, gross_duration_seconds, net_duration_seconds, notes, created_at, updated_at FROM study_sessions
WHERE id = ?
`

func (q *Queries) GetStudySession(ctx context.Context, id string) (StudySession, error) {
	row := q.db.QueryRowContext(ctx, getStudySession, id)
	var i StudySession
	err := row.Scan(
		&i.ID,
		&i.SubjectID,
		&i.CycleItemID,
		&i.StartedAt,
		&i.FinishedAt,
		&i.GrossDurationSeconds,
		&i.NetDurationSeconds,
		&i.Notes,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const updateSessionDuration = `-- name: UpdateSessionDuration :exec
UPDATE study_sessions
SET finished_at = ?, gross_duration_seconds = ?, net_duration_seconds = ?, notes = ?
WHERE id = ?
`

type UpdateSessionDurationParams struct {
	FinishedAt           sql.NullString `json:"finished_at"`
	GrossDurationSeconds sql.NullInt64  `json:"gross_duration_seconds"`
	NetDurationSeconds   sql.NullInt64  `json:"net_duration_seconds"`
	Notes                sql.NullString `json:"notes"`
	ID                   string         `json:"id"`
}

func (q *Queries) UpdateSessionDuration(ctx context.Context, arg UpdateSessionDurationParams) error {
	_, err := q.db.ExecContext(ctx, updateSessionDuration,
		arg.FinishedAt,
		arg.GrossDurationSeconds,
		arg.NetDurationSeconds,
		arg.Notes,
		arg.ID,
	)
	return err
}

```

--- 

**File:** `internal/database/subjects.sql.go`

```typescript
// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.30.0
// source: subjects.sql

package database

import (
	"context"
	"database/sql"
)

const createSubject = `-- name: CreateSubject :one
INSERT INTO subjects (id, name, color_hex)
VALUES (?, ?, ?)
RETURNING id, name, color_hex, created_at, updated_at, deleted_at
`

type CreateSubjectParams struct {
	ID       string         `json:"id"`
	Name     string         `json:"name"`
	ColorHex sql.NullString `json:"color_hex"`
}

func (q *Queries) CreateSubject(ctx context.Context, arg CreateSubjectParams) (Subject, error) {
	row := q.db.QueryRowContext(ctx, createSubject, arg.ID, arg.Name, arg.ColorHex)
	var i Subject
	err := row.Scan(
		&i.ID,
		&i.Name,
		&i.ColorHex,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.DeletedAt,
	)
	return i, err
}

const deleteSubject = `-- name: DeleteSubject :exec
UPDATE subjects
SET deleted_at = datetime('now')
WHERE id = ?
`

func (q *Queries) DeleteSubject(ctx context.Context, id string) error {
	_, err := q.db.ExecContext(ctx, deleteSubject, id)
	return err
}

const getSubject = `-- name: GetSubject :one
SELECT id, name, color_hex, created_at, updated_at, deleted_at FROM subjects
WHERE id = ? AND deleted_at IS NULL
`

func (q *Queries) GetSubject(ctx context.Context, id string) (Subject, error) {
	row := q.db.QueryRowContext(ctx, getSubject, id)
	var i Subject
	err := row.Scan(
		&i.ID,
		&i.Name,
		&i.ColorHex,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.DeletedAt,
	)
	return i, err
}

const listSubjects = `-- name: ListSubjects :many
SELECT id, name, color_hex, created_at, updated_at, deleted_at FROM subjects
WHERE deleted_at IS NULL
ORDER BY name
`

func (q *Queries) ListSubjects(ctx context.Context) ([]Subject, error) {
	rows, err := q.db.QueryContext(ctx, listSubjects)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Subject
	for rows.Next() {
		var i Subject
		if err := rows.Scan(
			&i.ID,
			&i.Name,
			&i.ColorHex,
			&i.CreatedAt,
			&i.UpdatedAt,
			&i.DeletedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const updateSubject = `-- name: UpdateSubject :exec
UPDATE subjects
SET name = ?, color_hex = ?, updated_at = datetime('now')
WHERE id = ? AND deleted_at IS NULL
`

type UpdateSubjectParams struct {
	Name     string         `json:"name"`
	ColorHex sql.NullString `json:"color_hex"`
	ID       string         `json:"id"`
}

func (q *Queries) UpdateSubject(ctx context.Context, arg UpdateSubjectParams) error {
	_, err := q.db.ExecContext(ctx, updateSubject, arg.Name, arg.ColorHex, arg.ID)
	return err
}

```

--- 

**File:** `internal/database/topics.sql.go`

```typescript
// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.30.0
// source: topics.sql

package database

import (
	"context"
)

const createTopic = `-- name: CreateTopic :one
INSERT INTO topics (id, subject_id, name)
VALUES (?, ?, ?)
RETURNING id, subject_id, name, created_at, updated_at, deleted_at
`

type CreateTopicParams struct {
	ID        string `json:"id"`
	SubjectID string `json:"subject_id"`
	Name      string `json:"name"`
}

func (q *Queries) CreateTopic(ctx context.Context, arg CreateTopicParams) (Topic, error) {
	row := q.db.QueryRowContext(ctx, createTopic, arg.ID, arg.SubjectID, arg.Name)
	var i Topic
	err := row.Scan(
		&i.ID,
		&i.SubjectID,
		&i.Name,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.DeletedAt,
	)
	return i, err
}

const deleteTopic = `-- name: DeleteTopic :exec
UPDATE topics
SET deleted_at = datetime('now')
WHERE id = ?
`

func (q *Queries) DeleteTopic(ctx context.Context, id string) error {
	_, err := q.db.ExecContext(ctx, deleteTopic, id)
	return err
}

const getTopic = `-- name: GetTopic :one
SELECT id, subject_id, name, created_at, updated_at, deleted_at FROM topics
WHERE id = ? AND deleted_at IS NULL
`

func (q *Queries) GetTopic(ctx context.Context, id string) (Topic, error) {
	row := q.db.QueryRowContext(ctx, getTopic, id)
	var i Topic
	err := row.Scan(
		&i.ID,
		&i.SubjectID,
		&i.Name,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.DeletedAt,
	)
	return i, err
}

const listTopicsBySubject = `-- name: ListTopicsBySubject :many
SELECT id, subject_id, name, created_at, updated_at, deleted_at FROM topics
WHERE subject_id = ? AND deleted_at IS NULL
ORDER BY name
`

func (q *Queries) ListTopicsBySubject(ctx context.Context, subjectID string) ([]Topic, error) {
	rows, err := q.db.QueryContext(ctx, listTopicsBySubject, subjectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Topic
	for rows.Next() {
		var i Topic
		if err := rows.Scan(
			&i.ID,
			&i.SubjectID,
			&i.Name,
			&i.CreatedAt,
			&i.UpdatedAt,
			&i.DeletedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const updateTopic = `-- name: UpdateTopic :exec
UPDATE topics
SET name = ?, updated_at = datetime('now')
WHERE id = ? AND deleted_at IS NULL
`

type UpdateTopicParams struct {
	Name string `json:"name"`
	ID   string `json:"id"`
}

func (q *Queries) UpdateTopic(ctx context.Context, arg UpdateTopicParams) error {
	_, err := q.db.ExecContext(ctx, updateTopic, arg.Name, arg.ID)
	return err
}

```

--- 

**File:** `internal/database/users.sql.go`

```typescript
// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.30.0
// source: users.sql

package database

import (
	"context"
)

const createUser = `-- name: CreateUser :one
INSERT INTO users (id, email, name, password_hash)
VALUES (?, ?, ?, ?)
RETURNING id, email, name, created_at, password_hash
`

type CreateUserParams struct {
	ID           string `json:"id"`
	Email        string `json:"email"`
	Name         string `json:"name"`
	PasswordHash string `json:"password_hash"`
}

func (q *Queries) CreateUser(ctx context.Context, arg CreateUserParams) (User, error) {
	row := q.db.QueryRowContext(ctx, createUser,
		arg.ID,
		arg.Email,
		arg.Name,
		arg.PasswordHash,
	)
	var i User
	err := row.Scan(
		&i.ID,
		&i.Email,
		&i.Name,
		&i.CreatedAt,
		&i.PasswordHash,
	)
	return i, err
}

const getUserByEmail = `-- name: GetUserByEmail :one
SELECT id, email, name, created_at, password_hash FROM users
WHERE email = ? LIMIT 1
`

func (q *Queries) GetUserByEmail(ctx context.Context, email string) (User, error) {
	row := q.db.QueryRowContext(ctx, getUserByEmail, email)
	var i User
	err := row.Scan(
		&i.ID,
		&i.Email,
		&i.Name,
		&i.CreatedAt,
		&i.PasswordHash,
	)
	return i, err
}

```

--- 

**File:** `internal/handler/analytics_handler.go`

```typescript
package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/joaoapaenas/my-api/internal/service"
)

type AnalyticsHandler struct {
	svc service.AnalyticsService
}

func NewAnalyticsHandler(svc service.AnalyticsService) *AnalyticsHandler {
	return &AnalyticsHandler{svc: svc}
}

// GetTimeReport godoc
// @Summary Get net study time report by subject
// @Tags analytics
// @Produce json
// @Param start_date_from query string false "Start Date From (YYYY-MM-DD)"
// @Param start_date_to query string false "Start Date To (YYYY-MM-DD)"
// @Success 200 {array} database.GetTimeReportBySubjectRow
// @Router /analytics/time-report [get]
func (h *AnalyticsHandler) GetTimeReport(w http.ResponseWriter, r *http.Request) {
	startDateFrom := r.URL.Query().Get("start_date_from")
	startDateTo := r.URL.Query().Get("start_date_to")

	report, err := h.svc.GetTimeReport(r.Context(), startDateFrom, startDateTo)
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	h.respondWithJSON(w, http.StatusOK, report)
}

// GetGlobalAccuracy godoc
// @Summary Get global accuracy by subject
// @Tags analytics
// @Produce json
// @Success 200 {array} database.GetAccuracyBySubjectRow
// @Router /analytics/accuracy [get]
func (h *AnalyticsHandler) GetGlobalAccuracy(w http.ResponseWriter, r *http.Request) {
	report, err := h.svc.GetGlobalAccuracy(r.Context())
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	h.respondWithJSON(w, http.StatusOK, report)
}

// GetWeakPoints godoc
// @Summary Get weak points (accuracy by topic) for a subject
// @Tags analytics
// @Produce json
// @Param subject_id path string true "Subject ID"
// @Success 200 {array} database.GetAccuracyByTopicRow
// @Router /analytics/weak-points/{subject_id} [get]
func (h *AnalyticsHandler) GetWeakPoints(w http.ResponseWriter, r *http.Request) {
	subjectID := chi.URLParam(r, "subject_id")
	if subjectID == "" {
		h.respondWithError(w, http.StatusBadRequest, "Subject ID is required")
		return
	}

	report, err := h.svc.GetWeakPoints(r.Context(), subjectID)
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	h.respondWithJSON(w, http.StatusOK, report)
}

// GetHeatmap godoc
// @Summary Get study activity heatmap
// @Tags analytics
// @Produce json
// @Param days query int false "Number of days (default 30)"
// @Success 200 {array} database.GetActivityHeatmapRow
// @Router /analytics/heatmap [get]
func (h *AnalyticsHandler) GetHeatmap(w http.ResponseWriter, r *http.Request) {
	daysStr := r.URL.Query().Get("days")
	var days int64 = 30
	if daysStr != "" {
		if d, err := strconv.ParseInt(daysStr, 10, 64); err == nil {
			days = d
		}
	}

	heatmap, err := h.svc.GetHeatmap(r.Context(), days)
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	h.respondWithJSON(w, http.StatusOK, heatmap)
}

func (h *AnalyticsHandler) respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(payload)
}

func (h *AnalyticsHandler) respondWithError(w http.ResponseWriter, code int, message string) {
	h.respondWithJSON(w, code, map[string]string{"error": message})
}

```

--- 

**File:** `internal/handler/common.go`

```typescript
package handler

import "github.com/go-playground/validator/v10"

// formatValidationErrors formats validator errors into a readable map
func formatValidationErrors(err error) map[string]string {
	errors := make(map[string]string)
	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		for _, e := range validationErrors {
			errors[e.Field()] = e.Tag()
		}
	}
	return errors
}

// Response DTOs for Swagger documentation
type SubjectResponse struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	ColorHex  string `json:"color_hex,omitempty"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
	DeletedAt string `json:"deleted_at,omitempty"`
}

type TopicResponse struct {
	ID        string `json:"id"`
	SubjectID string `json:"subject_id"`
	Name      string `json:"name"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
	DeletedAt string `json:"deleted_at,omitempty"`
}

type StudyCycleResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	IsActive    int    `json:"is_active"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
	DeletedAt   string `json:"deleted_at,omitempty"`
}

type CycleItemResponse struct {
	ID                     string `json:"id"`
	CycleID                string `json:"cycle_id"`
	SubjectID              string `json:"subject_id"`
	OrderIndex             int    `json:"order_index"`
	PlannedDurationMinutes int    `json:"planned_duration_minutes,omitempty"`
	CreatedAt              string `json:"created_at"`
	UpdatedAt              string `json:"updated_at"`
}

type StudySessionResponse struct {
	ID                   string `json:"id"`
	SubjectID            string `json:"subject_id"`
	CycleItemID          string `json:"cycle_item_id,omitempty"`
	StartedAt            string `json:"started_at"`
	FinishedAt           string `json:"finished_at,omitempty"`
	GrossDurationSeconds int    `json:"gross_duration_seconds,omitempty"`
	NetDurationSeconds   int    `json:"net_duration_seconds,omitempty"`
	Notes                string `json:"notes,omitempty"`
}

type SessionPauseResponse struct {
	ID        string `json:"id"`
	SessionID string `json:"session_id"`
	StartedAt string `json:"started_at"`
	EndedAt   string `json:"ended_at,omitempty"`
}

type ExerciseLogResponse struct {
	ID             string `json:"id"`
	SessionID      string `json:"session_id,omitempty"`
	SubjectID      string `json:"subject_id"`
	TopicID        string `json:"topic_id,omitempty"`
	QuestionsCount int    `json:"questions_count"`
	CorrectCount   int    `json:"correct_count"`
	CreatedAt      string `json:"created_at"`
}

type MessageResponse struct {
	Message string `json:"message"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type ValidationErrorResponse struct {
	Error   string            `json:"error"`
	Details map[string]string `json:"details"`
}

// New response DTOs for TODO features
type CycleItemWithSubjectResponse struct {
	CycleItemID            string `json:"cycle_item_id"`
	OrderIndex             int    `json:"order_index"`
	PlannedDurationMinutes int    `json:"planned_duration_minutes,omitempty"`
	SubjectID              string `json:"subject_id"`
	SubjectName            string `json:"subject_name"`
	ColorHex               string `json:"color_hex,omitempty"`
}

type OpenSessionResponse struct {
	ID          string `json:"id"`
	SubjectID   string `json:"subject_id"`
	CycleItemID string `json:"cycle_item_id,omitempty"`
	StartedAt   string `json:"started_at"`
	SubjectName string `json:"subject_name"`
	ColorHex    string `json:"color_hex,omitempty"`
}

type TimeReportResponse struct {
	SubjectID     string  `json:"subject_id"`
	SubjectName   string  `json:"subject_name"`
	ColorHex      string  `json:"color_hex,omitempty"`
	SessionsCount int     `json:"sessions_count"`
	TotalHoursNet float64 `json:"total_hours_net"`
}

type AccuracyReportResponse struct {
	SubjectID          string  `json:"subject_id"`
	SubjectName        string  `json:"subject_name"`
	ColorHex           string  `json:"color_hex,omitempty"`
	TotalQuestions     int     `json:"total_questions"`
	TotalCorrect       int     `json:"total_correct"`
	AccuracyPercentage float64 `json:"accuracy_percentage"`
}

type TopicAccuracyResponse struct {
	TopicID            string  `json:"topic_id"`
	TopicName          string  `json:"topic_name"`
	TotalQuestions     int     `json:"total_questions"`
	TotalCorrect       int     `json:"total_correct"`
	AccuracyPercentage float64 `json:"accuracy_percentage"`
}

type HeatmapDayResponse struct {
	StudyDate     string `json:"study_date"`
	SessionsCount int    `json:"sessions_count"`
	TotalSeconds  int    `json:"total_seconds"`
}

```

--- 

**File:** `internal/handler/cycle_item_handler.go`

```typescript
package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/joaoapaenas/my-api/internal/service"
)

type CycleItemHandler struct {
	svc      service.CycleItemService
	validate *validator.Validate
}

func NewCycleItemHandler(svc service.CycleItemService) *CycleItemHandler {
	return &CycleItemHandler{svc: svc, validate: validator.New()}
}

type CreateCycleItemRequest struct {
	SubjectID              string `json:"subject_id" validate:"required"`
	OrderIndex             int    `json:"order_index" validate:"required,min=1"`
	PlannedDurationMinutes int    `json:"planned_duration_minutes" validate:"omitempty,min=1"`
}

type UpdateCycleItemRequest struct {
	SubjectID              string `json:"subject_id" validate:"required"`
	OrderIndex             int    `json:"order_index" validate:"required,min=1"`
	PlannedDurationMinutes int    `json:"planned_duration_minutes" validate:"omitempty,min=1"`
}

// CreateCycleItem godoc
// @Summary Create a new cycle item
// @Tags cycle_items
// @Accept json
// @Produce json
// @Param id path string true "Cycle ID"
// @Param input body CreateCycleItemRequest true "Cycle item info"
// @Success 201 {object} handler.CycleItemResponse
// @Router /study-cycles/{id}/items [post]
func (h *CycleItemHandler) CreateCycleItem(w http.ResponseWriter, r *http.Request) {
	cycleID := chi.URLParam(r, "id")
	if cycleID == "" {
		h.respondWithError(w, http.StatusBadRequest, "Cycle ID is required")
		return
	}

	var req CreateCycleItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	if err := h.validate.Struct(req); err != nil {
		h.respondWithJSON(w, http.StatusBadRequest, map[string]interface{}{
			"error":   "Validation failed",
			"details": formatValidationErrors(err),
		})
		return
	}

	item, err := h.svc.CreateCycleItem(r.Context(), cycleID, req.SubjectID, req.OrderIndex, req.PlannedDurationMinutes)
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	h.respondWithJSON(w, http.StatusCreated, item)
}

// ListCycleItems godoc
// @Summary List all items for a cycle
// @Tags cycle_items
// @Produce json
// @Param id path string true "Cycle ID"
// @Success 200 {array} handler.CycleItemResponse
// @Router /study-cycles/{id}/items [get]
func (h *CycleItemHandler) ListCycleItems(w http.ResponseWriter, r *http.Request) {
	cycleID := chi.URLParam(r, "id")
	if cycleID == "" {
		h.respondWithError(w, http.StatusBadRequest, "Cycle ID is required")
		return
	}

	items, err := h.svc.ListCycleItems(r.Context(), cycleID)
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	h.respondWithJSON(w, http.StatusOK, items)
}

// GetCycleItem godoc
// @Summary Get a cycle item by ID
// @Tags cycle_items
// @Produce json
// @Param id path string true "Item ID"
// @Success 200 {object} handler.CycleItemResponse
// @Router /cycle-items/{id} [get]
func (h *CycleItemHandler) GetCycleItem(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		h.respondWithError(w, http.StatusBadRequest, "Item ID is required")
		return
	}

	item, err := h.svc.GetCycleItem(r.Context(), id)
	if err != nil {
		h.respondWithError(w, http.StatusNotFound, "Cycle item not found")
		return
	}

	h.respondWithJSON(w, http.StatusOK, item)
}

// UpdateCycleItem godoc
// @Summary Update a cycle item
// @Tags cycle_items
// @Accept json
// @Produce json
// @Param id path string true "Item ID"
// @Param input body UpdateCycleItemRequest true "Cycle item info"
// @Success 200 {string} string "OK"
// @Router /cycle-items/{id} [put]
func (h *CycleItemHandler) UpdateCycleItem(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		h.respondWithError(w, http.StatusBadRequest, "Item ID is required")
		return
	}

	var req UpdateCycleItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	if err := h.validate.Struct(req); err != nil {
		h.respondWithJSON(w, http.StatusBadRequest, map[string]interface{}{
			"error":   "Validation failed",
			"details": formatValidationErrors(err),
		})
		return
	}

	err := h.svc.UpdateCycleItem(r.Context(), id, req.SubjectID, req.OrderIndex, req.PlannedDurationMinutes)
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	h.respondWithJSON(w, http.StatusOK, map[string]string{"message": "Cycle item updated successfully"})
}

// DeleteCycleItem godoc
// @Summary Delete a cycle item
// @Tags cycle_items
// @Param id path string true "Item ID"
// @Success 204
// @Router /cycle-items/{id} [delete]
func (h *CycleItemHandler) DeleteCycleItem(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		h.respondWithError(w, http.StatusBadRequest, "Item ID is required")
		return
	}

	err := h.svc.DeleteCycleItem(r.Context(), id)
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *CycleItemHandler) respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(payload)
}

func (h *CycleItemHandler) respondWithError(w http.ResponseWriter, code int, message string) {
	h.respondWithJSON(w, code, map[string]string{"error": message})
}

```

--- 

**File:** `internal/handler/exercise_log_handler.go`

```typescript
package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/joaoapaenas/my-api/internal/service"
)

type ExerciseLogHandler struct {
	svc      service.ExerciseLogService
	validate *validator.Validate
}

func NewExerciseLogHandler(svc service.ExerciseLogService) *ExerciseLogHandler {
	return &ExerciseLogHandler{svc: svc, validate: validator.New()}
}

type CreateExerciseLogRequest struct {
	SessionID      string `json:"session_id"`
	SubjectID      string `json:"subject_id" validate:"required"`
	TopicID        string `json:"topic_id"`
	QuestionsCount int    `json:"questions_count" validate:"required,min=0"`
	CorrectCount   int    `json:"correct_count" validate:"required,min=0"`
}

// CreateExerciseLog godoc
// @Summary Create a new exercise log
// @Tags exercise_logs
// @Accept json
// @Produce json
// @Param input body CreateExerciseLogRequest true "Exercise log info"
// @Success 201 {object} handler.ExerciseLogResponse
// @Router /exercise-logs [post]
func (h *ExerciseLogHandler) CreateExerciseLog(w http.ResponseWriter, r *http.Request) {
	var req CreateExerciseLogRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	if err := h.validate.Struct(req); err != nil {
		h.respondWithJSON(w, http.StatusBadRequest, map[string]interface{}{
			"error":   "Validation failed",
			"details": formatValidationErrors(err),
		})
		return
	}

	if req.CorrectCount > req.QuestionsCount {
		h.respondWithError(w, http.StatusBadRequest, "Correct count cannot exceed questions count")
		return
	}

	log, err := h.svc.CreateExerciseLog(r.Context(), req.SessionID, req.SubjectID, req.TopicID, req.QuestionsCount, req.CorrectCount)
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	h.respondWithJSON(w, http.StatusCreated, log)
}

// GetExerciseLog godoc
// @Summary Get an exercise log by ID
// @Tags exercise_logs
// @Produce json
// @Param id path string true "Log ID"
// @Success 200 {object} handler.ExerciseLogResponse
// @Router /exercise-logs/{id} [get]
func (h *ExerciseLogHandler) GetExerciseLog(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		h.respondWithError(w, http.StatusBadRequest, "Log ID is required")
		return
	}

	log, err := h.svc.GetExerciseLog(r.Context(), id)
	if err != nil {
		h.respondWithError(w, http.StatusNotFound, "Exercise log not found")
		return
	}

	h.respondWithJSON(w, http.StatusOK, log)
}

// DeleteExerciseLog godoc
// @Summary Delete an exercise log
// @Tags exercise_logs
// @Param id path string true "Log ID"
// @Success 204
// @Router /exercise-logs/{id} [delete]
func (h *ExerciseLogHandler) DeleteExerciseLog(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		h.respondWithError(w, http.StatusBadRequest, "Log ID is required")
		return
	}

	err := h.svc.DeleteExerciseLog(r.Context(), id)
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *ExerciseLogHandler) respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(payload)
}

func (h *ExerciseLogHandler) respondWithError(w http.ResponseWriter, code int, message string) {
	h.respondWithJSON(w, code, map[string]string{"error": message})
}

```

--- 

**File:** `internal/handler/session_pause_handler.go`

```typescript
package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/joaoapaenas/my-api/internal/service"
)

type SessionPauseHandler struct {
	svc      service.SessionPauseService
	validate *validator.Validate
}

func NewSessionPauseHandler(svc service.SessionPauseService) *SessionPauseHandler {
	return &SessionPauseHandler{svc: svc, validate: validator.New()}
}

type CreateSessionPauseRequest struct {
	SessionID string `json:"session_id" validate:"required"`
	StartedAt string `json:"started_at" validate:"required"`
}

type EndSessionPauseRequest struct {
	EndedAt string `json:"ended_at" validate:"required"`
}

// CreateSessionPause godoc
// @Summary Create a new session pause
// @Tags session_pauses
// @Accept json
// @Produce json
// @Param input body CreateSessionPauseRequest true "Session pause info"
// @Success 201 {object} handler.SessionPauseResponse
// @Router /session-pauses [post]
func (h *SessionPauseHandler) CreateSessionPause(w http.ResponseWriter, r *http.Request) {
	var req CreateSessionPauseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	if err := h.validate.Struct(req); err != nil {
		h.respondWithJSON(w, http.StatusBadRequest, map[string]interface{}{
			"error":   "Validation failed",
			"details": formatValidationErrors(err),
		})
		return
	}

	pause, err := h.svc.CreateSessionPause(r.Context(), req.SessionID, req.StartedAt)
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	h.respondWithJSON(w, http.StatusCreated, pause)
}

// GetSessionPause godoc
// @Summary Get a session pause by ID
// @Tags session_pauses
// @Produce json
// @Param id path string true "Pause ID"
// @Success 200 {object} handler.SessionPauseResponse
// @Router /session-pauses/{id} [get]
func (h *SessionPauseHandler) GetSessionPause(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		h.respondWithError(w, http.StatusBadRequest, "Pause ID is required")
		return
	}

	pause, err := h.svc.GetSessionPause(r.Context(), id)
	if err != nil {
		h.respondWithError(w, http.StatusNotFound, "Session pause not found")
		return
	}

	h.respondWithJSON(w, http.StatusOK, pause)
}

// EndSessionPause godoc
// @Summary End a session pause
// @Tags session_pauses
// @Accept json
// @Produce json
// @Param id path string true "Pause ID"
// @Param input body EndSessionPauseRequest true "End pause info"
// @Success 200 {string} string "OK"
// @Router /session-pauses/{id}/end [put]
func (h *SessionPauseHandler) EndSessionPause(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		h.respondWithError(w, http.StatusBadRequest, "Pause ID is required")
		return
	}

	var req EndSessionPauseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	if err := h.validate.Struct(req); err != nil {
		h.respondWithJSON(w, http.StatusBadRequest, map[string]interface{}{
			"error":   "Validation failed",
			"details": formatValidationErrors(err),
		})
		return
	}

	err := h.svc.EndSessionPause(r.Context(), id, req.EndedAt)
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	h.respondWithJSON(w, http.StatusOK, map[string]string{"message": "Pause ended successfully"})
}

// DeleteSessionPause godoc
// @Summary Delete a session pause
// @Tags session_pauses
// @Param id path string true "Pause ID"
// @Success 204
// @Router /session-pauses/{id} [delete]
func (h *SessionPauseHandler) DeleteSessionPause(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		h.respondWithError(w, http.StatusBadRequest, "Pause ID is required")
		return
	}

	err := h.svc.DeleteSessionPause(r.Context(), id)
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *SessionPauseHandler) respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(payload)
}

func (h *SessionPauseHandler) respondWithError(w http.ResponseWriter, code int, message string) {
	h.respondWithJSON(w, code, map[string]string{"error": message})
}

```

--- 

**File:** `internal/handler/study_cycle_handler.go`

```typescript
package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/joaoapaenas/my-api/internal/service"
)

type StudyCycleHandler struct {
	svc      service.StudyCycleService
	validate *validator.Validate
}

func NewStudyCycleHandler(svc service.StudyCycleService) *StudyCycleHandler {
	return &StudyCycleHandler{svc: svc, validate: validator.New()}
}

type CreateStudyCycleRequest struct {
	Name        string `json:"name" validate:"required,min=2"`
	Description string `json:"description"`
	IsActive    bool   `json:"is_active"`
}

type UpdateStudyCycleRequest struct {
	Name        string `json:"name" validate:"required,min=2"`
	Description string `json:"description"`
	IsActive    bool   `json:"is_active"`
}

// CreateStudyCycle godoc
// @Summary Create a new study cycle
// @Tags study_cycles
// @Accept json
// @Produce json
// @Param input body CreateStudyCycleRequest true "Study cycle info"
// @Success 201 {object} handler.StudyCycleResponse
// @Router /study-cycles [post]
func (h *StudyCycleHandler) CreateStudyCycle(w http.ResponseWriter, r *http.Request) {
	var req CreateStudyCycleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	if err := h.validate.Struct(req); err != nil {
		h.respondWithJSON(w, http.StatusBadRequest, map[string]interface{}{
			"error":   "Validation failed",
			"details": formatValidationErrors(err),
		})
		return
	}

	cycle, err := h.svc.CreateStudyCycle(r.Context(), req.Name, req.Description, req.IsActive)
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	h.respondWithJSON(w, http.StatusCreated, cycle)
}

// GetActiveStudyCycle godoc
// @Summary Get the active study cycle
// @Tags study_cycles
// @Produce json
// @Success 200 {object} handler.StudyCycleResponse
// @Router /study-cycles/active [get]
func (h *StudyCycleHandler) GetActiveStudyCycle(w http.ResponseWriter, r *http.Request) {
	cycle, err := h.svc.GetActiveStudyCycle(r.Context())
	if err != nil {
		h.respondWithError(w, http.StatusNotFound, "No active study cycle found")
		return
	}

	h.respondWithJSON(w, http.StatusOK, cycle)
}

// GetStudyCycle godoc
// @Summary Get a study cycle by ID
// @Tags study_cycles
// @Produce json
// @Param id path string true "Cycle ID"
// @Success 200 {object} handler.StudyCycleResponse
// @Router /study-cycles/{id} [get]
func (h *StudyCycleHandler) GetStudyCycle(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		h.respondWithError(w, http.StatusBadRequest, "Cycle ID is required")
		return
	}

	cycle, err := h.svc.GetStudyCycle(r.Context(), id)
	if err != nil {
		h.respondWithError(w, http.StatusNotFound, "Study cycle not found")
		return
	}

	h.respondWithJSON(w, http.StatusOK, cycle)
}

// UpdateStudyCycle godoc
// @Summary Update a study cycle
// @Tags study_cycles
// @Accept json
// @Produce json
// @Param id path string true "Cycle ID"
// @Param input body UpdateStudyCycleRequest true "Study cycle info"
// @Success 200 {string} string "OK"
// @Router /study-cycles/{id} [put]
func (h *StudyCycleHandler) UpdateStudyCycle(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		h.respondWithError(w, http.StatusBadRequest, "Cycle ID is required")
		return
	}

	var req UpdateStudyCycleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	if err := h.validate.Struct(req); err != nil {
		h.respondWithJSON(w, http.StatusBadRequest, map[string]interface{}{
			"error":   "Validation failed",
			"details": formatValidationErrors(err),
		})
		return
	}

	err := h.svc.UpdateStudyCycle(r.Context(), id, req.Name, req.Description, req.IsActive)
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	h.respondWithJSON(w, http.StatusOK, map[string]string{"message": "Study cycle updated successfully"})
}

// DeleteStudyCycle godoc
// @Summary Delete a study cycle
// @Tags study_cycles
// @Param id path string true "Cycle ID"
// @Success 204
// @Router /study-cycles/{id} [delete]
func (h *StudyCycleHandler) DeleteStudyCycle(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		h.respondWithError(w, http.StatusBadRequest, "Cycle ID is required")
		return
	}

	err := h.svc.DeleteStudyCycle(r.Context(), id)
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetActiveCycleWithItems godoc
// @Summary Get active cycle with all items (round-robin)
// @Tags study_cycles
// @Produce json
// @Success 200 {array} handler.CycleItemWithSubjectResponse
// @Router /study-cycles/active/items [get]
func (h *StudyCycleHandler) GetActiveCycleWithItems(w http.ResponseWriter, r *http.Request) {
	items, err := h.svc.GetActiveCycleWithItems(r.Context())
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	h.respondWithJSON(w, http.StatusOK, items)
}

func (h *StudyCycleHandler) respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(payload)
}

func (h *StudyCycleHandler) respondWithError(w http.ResponseWriter, code int, message string) {
	h.respondWithJSON(w, code, map[string]string{"error": message})
}

```

--- 

**File:** `internal/handler/study_session_handler.go`

```typescript
package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/joaoapaenas/my-api/internal/service"
)

type StudySessionHandler struct {
	svc      service.StudySessionService
	validate *validator.Validate
}

func NewStudySessionHandler(svc service.StudySessionService) *StudySessionHandler {
	return &StudySessionHandler{svc: svc, validate: validator.New()}
}

type CreateStudySessionRequest struct {
	SubjectID   string `json:"subject_id" validate:"required"`
	CycleItemID string `json:"cycle_item_id"`
	StartedAt   string `json:"started_at" validate:"required"`
}

type UpdateSessionDurationRequest struct {
	FinishedAt           string `json:"finished_at"`
	GrossDurationSeconds int    `json:"gross_duration_seconds"`
	NetDurationSeconds   int    `json:"net_duration_seconds"`
	Notes                string `json:"notes"`
}

// CreateStudySession godoc
// @Summary Create a new study session
// @Tags study_sessions
// @Accept json
// @Produce json
// @Param input body CreateStudySessionRequest true "Study session info"
// @Success 201 {object} handler.StudySessionResponse
// @Router /study-sessions [post]
func (h *StudySessionHandler) CreateStudySession(w http.ResponseWriter, r *http.Request) {
	var req CreateStudySessionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	if err := h.validate.Struct(req); err != nil {
		h.respondWithJSON(w, http.StatusBadRequest, map[string]interface{}{
			"error":   "Validation failed",
			"details": formatValidationErrors(err),
		})
		return
	}

	session, err := h.svc.CreateStudySession(r.Context(), req.SubjectID, req.CycleItemID, req.StartedAt)
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	h.respondWithJSON(w, http.StatusCreated, session)
}

// GetStudySession godoc
// @Summary Get a study session by ID
// @Tags study_sessions
// @Produce json
// @Param id path string true "Session ID"
// @Success 200 {object} handler.StudySessionResponse
// @Router /study-sessions/{id} [get]
func (h *StudySessionHandler) GetStudySession(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		h.respondWithError(w, http.StatusBadRequest, "Session ID is required")
		return
	}

	session, err := h.svc.GetStudySession(r.Context(), id)
	if err != nil {
		h.respondWithError(w, http.StatusNotFound, "Study session not found")
		return
	}

	h.respondWithJSON(w, http.StatusOK, session)
}

// UpdateSessionDuration godoc
// @Summary Update study session duration
// @Tags study_sessions
// @Accept json
// @Produce json
// @Param id path string true "Session ID"
// @Param input body UpdateSessionDurationRequest true "Session duration info"
// @Success 200 {string} string "OK"
// @Router /study-sessions/{id} [put]
func (h *StudySessionHandler) UpdateSessionDuration(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		h.respondWithError(w, http.StatusBadRequest, "Session ID is required")
		return
	}

	var req UpdateSessionDurationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	err := h.svc.UpdateSessionDuration(r.Context(), id, req.FinishedAt, req.GrossDurationSeconds, req.NetDurationSeconds, req.Notes)
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	h.respondWithJSON(w, http.StatusOK, map[string]string{"message": "Session updated successfully"})
}

// DeleteStudySession godoc
// @Summary Delete a study session
// @Tags study_sessions
// @Param id path string true "Session ID"
// @Success 204
// @Router /study-sessions/{id} [delete]
func (h *StudySessionHandler) DeleteStudySession(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		h.respondWithError(w, http.StatusBadRequest, "Session ID is required")
		return
	}

	err := h.svc.DeleteStudySession(r.Context(), id)
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetOpenSession godoc
// @Summary Get open/unfinished session (crash recovery)
// @Tags study_sessions
// @Produce json
// @Success 200 {object} handler.OpenSessionResponse
// @Router /study-sessions/open [get]
func (h *StudySessionHandler) GetOpenSession(w http.ResponseWriter, r *http.Request) {
	session, err := h.svc.GetOpenSession(r.Context())
	if err != nil {
		h.respondWithError(w, http.StatusNotFound, "No open session found")
		return
	}

	h.respondWithJSON(w, http.StatusOK, session)
}

func (h *StudySessionHandler) respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(payload)
}

func (h *StudySessionHandler) respondWithError(w http.ResponseWriter, code int, message string) {
	h.respondWithJSON(w, code, map[string]string{"error": message})
}

```

--- 

**File:** `internal/handler/subject_handler.go`

```typescript
package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/joaoapaenas/my-api/internal/service"
)

type SubjectHandler struct {
	svc      service.SubjectService
	validate *validator.Validate
}

func NewSubjectHandler(svc service.SubjectService) *SubjectHandler {
	return &SubjectHandler{svc: svc, validate: validator.New()}
}

type CreateSubjectRequest struct {
	Name     string `json:"name" validate:"required,min=2"`
	ColorHex string `json:"color_hex" validate:"omitempty,hexcolor"`
}

type UpdateSubjectRequest struct {
	Name     string `json:"name" validate:"required,min=2"`
	ColorHex string `json:"color_hex" validate:"omitempty,hexcolor"`
}

// CreateSubject godoc
// @Summary Create a new subject
// @Tags subjects
// @Accept json
// @Produce json
// @Param input body CreateSubjectRequest true "Subject info"
// @Success 201 {object} handler.SubjectResponse
// @Router /subjects [post]
func (h *SubjectHandler) CreateSubject(w http.ResponseWriter, r *http.Request) {
	var req CreateSubjectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	if err := h.validate.Struct(req); err != nil {
		h.respondWithJSON(w, http.StatusBadRequest, map[string]interface{}{
			"error":   "Validation failed",
			"details": formatValidationErrors(err),
		})
		return
	}

	subject, err := h.svc.CreateSubject(r.Context(), req.Name, req.ColorHex)
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	h.respondWithJSON(w, http.StatusCreated, subject)
}

// ListSubjects godoc
// @Summary List all subjects
// @Tags subjects
// @Produce json
// @Success 200 {array} handler.SubjectResponse
// @Router /subjects [get]
func (h *SubjectHandler) ListSubjects(w http.ResponseWriter, r *http.Request) {
	subjects, err := h.svc.ListSubjects(r.Context())
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	h.respondWithJSON(w, http.StatusOK, subjects)
}

// GetSubject godoc
// @Summary Get a subject by ID
// @Tags subjects
// @Produce json
// @Param id path string true "Subject ID"
// @Success 200 {object} handler.SubjectResponse
// @Router /subjects/{id} [get]
func (h *SubjectHandler) GetSubject(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		h.respondWithError(w, http.StatusBadRequest, "Subject ID is required")
		return
	}

	subject, err := h.svc.GetSubject(r.Context(), id)
	if err != nil {
		h.respondWithError(w, http.StatusNotFound, "Subject not found")
		return
	}

	h.respondWithJSON(w, http.StatusOK, subject)
}

// UpdateSubject godoc
// @Summary Update a subject
// @Tags subjects
// @Accept json
// @Produce json
// @Param id path string true "Subject ID"
// @Param input body UpdateSubjectRequest true "Subject info"
// @Success 200 {string} string "OK"
// @Router /subjects/{id} [put]
func (h *SubjectHandler) UpdateSubject(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		h.respondWithError(w, http.StatusBadRequest, "Subject ID is required")
		return
	}

	var req UpdateSubjectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	if err := h.validate.Struct(req); err != nil {
		h.respondWithJSON(w, http.StatusBadRequest, map[string]interface{}{
			"error":   "Validation failed",
			"details": formatValidationErrors(err),
		})
		return
	}

	err := h.svc.UpdateSubject(r.Context(), id, req.Name, req.ColorHex)
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	h.respondWithJSON(w, http.StatusOK, map[string]string{"message": "Subject updated successfully"})
}

// DeleteSubject godoc
// @Summary Delete a subject
// @Tags subjects
// @Param id path string true "Subject ID"
// @Success 204
// @Router /subjects/{id} [delete]
func (h *SubjectHandler) DeleteSubject(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		h.respondWithError(w, http.StatusBadRequest, "Subject ID is required")
		return
	}

	err := h.svc.DeleteSubject(r.Context(), id)
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *SubjectHandler) respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(payload)
}

func (h *SubjectHandler) respondWithError(w http.ResponseWriter, code int, message string) {
	h.respondWithJSON(w, code, map[string]string{"error": message})
}

```

--- 

**File:** `internal/handler/topic_handler.go`

```typescript
package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/joaoapaenas/my-api/internal/service"
)

type TopicHandler struct {
	svc      service.TopicService
	validate *validator.Validate
}

func NewTopicHandler(svc service.TopicService) *TopicHandler {
	return &TopicHandler{svc: svc, validate: validator.New()}
}

type CreateTopicRequest struct {
	Name string `json:"name" validate:"required,min=2"`
}

type UpdateTopicRequest struct {
	Name string `json:"name" validate:"required,min=2"`
}

// CreateTopic godoc
// @Summary Create a new topic for a subject
// @Tags topics
// @Accept json
// @Produce json
// @Param id path string true "Subject ID"
// @Param input body CreateTopicRequest true "Topic info"
// @Success 201 {object} handler.TopicResponse
// @Router /subjects/{id}/topics [post]
func (h *TopicHandler) CreateTopic(w http.ResponseWriter, r *http.Request) {
	subjectID := chi.URLParam(r, "id")
	if subjectID == "" {
		h.respondWithError(w, http.StatusBadRequest, "Subject ID is required")
		return
	}

	var req CreateTopicRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	if err := h.validate.Struct(req); err != nil {
		h.respondWithJSON(w, http.StatusBadRequest, map[string]interface{}{
			"error":   "Validation failed",
			"details": formatValidationErrors(err),
		})
		return
	}

	topic, err := h.svc.CreateTopic(r.Context(), subjectID, req.Name)
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	h.respondWithJSON(w, http.StatusCreated, topic)
}

// ListTopics godoc
// @Summary List all topics for a subject
// @Tags topics
// @Produce json
// @Param id path string true "Subject ID"
// @Success 200 {array} handler.TopicResponse
// @Router /subjects/{id}/topics [get]
func (h *TopicHandler) ListTopics(w http.ResponseWriter, r *http.Request) {
	subjectID := chi.URLParam(r, "id")
	if subjectID == "" {
		h.respondWithError(w, http.StatusBadRequest, "Subject ID is required")
		return
	}

	topics, err := h.svc.ListTopicsBySubject(r.Context(), subjectID)
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	h.respondWithJSON(w, http.StatusOK, topics)
}

// GetTopic godoc
// @Summary Get a topic by ID
// @Tags topics
// @Produce json
// @Param id path string true "Topic ID"
// @Success 200 {object} handler.TopicResponse
// @Router /topics/{id} [get]
func (h *TopicHandler) GetTopic(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		h.respondWithError(w, http.StatusBadRequest, "Topic ID is required")
		return
	}

	topic, err := h.svc.GetTopic(r.Context(), id)
	if err != nil {
		h.respondWithError(w, http.StatusNotFound, "Topic not found")
		return
	}

	h.respondWithJSON(w, http.StatusOK, topic)
}

// UpdateTopic godoc
// @Summary Update a topic
// @Tags topics
// @Accept json
// @Produce json
// @Param id path string true "Topic ID"
// @Param input body UpdateTopicRequest true "Topic info"
// @Success 200 {string} string "OK"
// @Router /topics/{id} [put]
func (h *TopicHandler) UpdateTopic(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		h.respondWithError(w, http.StatusBadRequest, "Topic ID is required")
		return
	}

	var req UpdateTopicRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	if err := h.validate.Struct(req); err != nil {
		h.respondWithJSON(w, http.StatusBadRequest, map[string]interface{}{
			"error":   "Validation failed",
			"details": formatValidationErrors(err),
		})
		return
	}

	err := h.svc.UpdateTopic(r.Context(), id, req.Name)
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	h.respondWithJSON(w, http.StatusOK, map[string]string{"message": "Topic updated successfully"})
}

// DeleteTopic godoc
// @Summary Delete a topic
// @Tags topics
// @Param id path string true "Topic ID"
// @Success 204
// @Router /topics/{id} [delete]
func (h *TopicHandler) DeleteTopic(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		h.respondWithError(w, http.StatusBadRequest, "Topic ID is required")
		return
	}

	err := h.svc.DeleteTopic(r.Context(), id)
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *TopicHandler) respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(payload)
}

func (h *TopicHandler) respondWithError(w http.ResponseWriter, code int, message string) {
	h.respondWithJSON(w, code, map[string]string{"error": message})
}

```

--- 

**File:** `internal/handler/user_handler.go`

```typescript
package handler

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/joaoapaenas/my-api/internal/service"
)

type UserHandler struct {
	svc      service.UserService
	validate *validator.Validate
}

func NewUserHandler(svc service.UserService) *UserHandler {
	return &UserHandler{svc: svc, validate: validator.New()}
}

type CreateUserRequest struct {
	// required: cannot be empty
	// email: must be a valid email format
	Email string `json:"email" validate:"required,email"`

	// min=2: must be at least 2 chars
	Name string `json:"name" validate:"required,min=2"`

	// min=6: must be at least 6 chars
	Password string `json:"password" validate:"required,min=6"`
}

// CreateUser godoc
// @Summary Create a new user
// @Description Create a user with email and name
// @Tags users
// @Accept json
// @Produce json
// @Param input body CreateUserRequest true "User info"
// @Success 201 {object} database.User
// @Failure 400 {string} string "Invalid request"
// @Failure 500 {string} string "Internal server error"
// @Router /users [post]
func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	// 1. Decode & Basic Validation
	var req CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	if err := h.validate.Struct(req); err != nil {
		// Return friendly validation errors
		validationErrors := formatValidationErrors(err)
		h.respondWithJSON(w, http.StatusBadRequest, map[string]interface{}{
			"error":   "Validation failed",
			"details": validationErrors,
		})
		return
	}

	// 2. Call Service (Business Logic)
	// Notice: We don't generate UUIDs here anymore.
	user, err := h.svc.CreateUser(r.Context(), req.Email, req.Name, req.Password)
	if err != nil {
		// Check for specific domain errors if you defined them
		if errors.Is(err, service.ErrEmailTaken) {
			h.respondWithError(w, http.StatusConflict, "Email already exists")
			return
		}

		slog.Error("Failed to create user", "error", err)
		h.respondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	h.respondWithJSON(w, http.StatusCreated, user)
}

// GetUser godoc
// @Summary Get user by Email
// @Tags users
// @Param email path string true "User Email"
// @Success 200 {object} database.User
// @Failure 404 {string} string "User not found"
// @Router /users/{email} [get]
func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	email := chi.URLParam(r, "email")

	user, err := h.svc.GetUserByEmail(r.Context(), email)
	if err != nil {
		// Handle "Not Found" specifically
		if err == sql.ErrNoRows || errors.Is(err, service.ErrUserNotFound) {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(user)
}

// --- Helpers ---

func (h *UserHandler) respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(payload)
}

func (h *UserHandler) respondWithError(w http.ResponseWriter, code int, message string) {
	h.respondWithJSON(w, code, map[string]string{"error": message})
}

```

--- 

**File:** `internal/handler/user_handler_test.go`

```typescript
package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/joaoapaenas/my-api/internal/database"
	"github.com/joaoapaenas/my-api/internal/handler"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockUserService is a mock implementation of service.UserService
type MockUserService struct {
	mock.Mock
}

func (m *MockUserService) CreateUser(ctx context.Context, email, name, password string) (database.User, error) {
	args := m.Called(ctx, email, name, password)
	return args.Get(0).(database.User), args.Error(1)
}

func (m *MockUserService) GetUserByEmail(ctx context.Context, email string) (database.User, error) {
	args := m.Called(ctx, email)
	return args.Get(0).(database.User), args.Error(1)
}

func TestUserHandler_CreateUser(t *testing.T) {
	mockSvc := new(MockUserService)
	h := handler.NewUserHandler(mockSvc)

	tests := []struct {
		name           string
		input          handler.CreateUserRequest
		mockReturnUser database.User
		mockReturnErr  error
		wantStatus     int
	}{
		{
			name: "Success",
			input: handler.CreateUserRequest{
				Email:    "test@example.com",
				Name:     "Test User",
				Password: "password123",
			},
			mockReturnUser: database.User{ID: "1", Email: "test@example.com"},
			mockReturnErr:  nil,
			wantStatus:     http.StatusCreated,
		},
		{
			name: "Validation Error",
			input: handler.CreateUserRequest{
				Email: "invalid-email",
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "Service Error",
			input: handler.CreateUserRequest{
				Email:    "test@example.com",
				Name:     "Test User",
				Password: "password123",
			},
			mockReturnErr: errors.New("db error"),
			wantStatus:    http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock expectation only if validation passes
			if tt.wantStatus != http.StatusBadRequest {
				mockSvc.On("CreateUser", mock.Anything, tt.input.Email, tt.input.Name, tt.input.Password).
					Return(tt.mockReturnUser, tt.mockReturnErr).
					Once()
			}

			body, _ := json.Marshal(tt.input)
			req := httptest.NewRequest("POST", "/users", bytes.NewBuffer(body))
			rr := httptest.NewRecorder()

			h.CreateUser(rr, req)

			assert.Equal(t, tt.wantStatus, rr.Code)
			if tt.wantStatus == http.StatusCreated {
				var user database.User
				json.NewDecoder(rr.Body).Decode(&user)
				assert.Equal(t, tt.mockReturnUser.Email, user.Email)
			}
		})
	}
}

```

--- 

**File:** `internal/logger/logger.go`

```typescript
package logger

import (
	"log/slog"
	"os"
)

// Init configures the global logger.
// env: "development" (text logs) or "production" (json logs)
func Init(env string) {
	var handler slog.Handler

	opts := &slog.HandlerOptions{
		Level: slog.LevelInfo, // Change to slog.LevelDebug for more verbosity
	}

	if env == "production" {
		// JSON is machine-readable (required for AWS CloudWatch)
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		// Text is human-readable (nice for local dev)
		handler = slog.NewTextHandler(os.Stdout, opts)
	}

	logger := slog.New(handler)
	slog.SetDefault(logger)
}

```

--- 

**File:** `internal/middleware/basic_auth.go`

```typescript
package middleware

import (
	"log/slog"
	"net/http"

	"github.com/joaoapaenas/my-api/internal/service"
	"golang.org/x/crypto/bcrypt"
)

type BasicAuthMiddleware struct {
	userService service.UserService
}

func NewBasicAuthMiddleware(userService service.UserService) *BasicAuthMiddleware {
	return &BasicAuthMiddleware{userService: userService}
}

func (m *BasicAuthMiddleware) BasicAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		email, password, ok := r.BasicAuth()
		if !ok {
			w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		user, err := m.userService.GetUserByEmail(r.Context(), email)
		if err != nil {
			// User not found or DB error
			slog.Warn("Basic Auth failed: user not found", "email", email, "error", err)
			w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Verify password
		if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
			slog.Warn("Basic Auth failed: invalid password", "email", email)
			w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}

```

--- 

**File:** `internal/middleware/logger.go`

```typescript
package middleware

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
)

// RequestLogger is a middleware that logs the start and end of each request.
func RequestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Get the Request ID from Chi middleware
		reqID := middleware.GetReqID(r.Context())

		// Create a wrapper to capture the status code and bytes written
		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

		defer func() {
			// Log the request completion
			slog.Info("HTTP Request",
				"method", r.Method,
				"path", r.URL.Path,
				"status", ww.Status(),
				"bytes", ww.BytesWritten(),
				"duration", time.Since(start).String(),
				"request_id", reqID,
				"ip", r.RemoteAddr,
			)
		}()

		next.ServeHTTP(ww, r)
	})
}

```

--- 

**File:** `internal/repository/analytics_repository.go`

```typescript
package repository

import (
	"context"

	"github.com/joaoapaenas/my-api/internal/database"
)

type AnalyticsRepository interface {
	GetTimeReportBySubject(ctx context.Context, arg database.GetTimeReportBySubjectParams) ([]database.GetTimeReportBySubjectRow, error)
	GetAccuracyBySubject(ctx context.Context) ([]database.GetAccuracyBySubjectRow, error)
	GetAccuracyByTopic(ctx context.Context, subjectID string) ([]database.GetAccuracyByTopicRow, error)
	GetActivityHeatmap(ctx context.Context, daysCount string) ([]database.GetActivityHeatmapRow, error)
}

type SQLAnalyticsRepository struct {
	q database.Querier
}

func NewSQLAnalyticsRepository(q database.Querier) *SQLAnalyticsRepository {
	return &SQLAnalyticsRepository{q: q}
}

func (r *SQLAnalyticsRepository) GetTimeReportBySubject(ctx context.Context, arg database.GetTimeReportBySubjectParams) ([]database.GetTimeReportBySubjectRow, error) {
	return r.q.GetTimeReportBySubject(ctx, arg)
}

func (r *SQLAnalyticsRepository) GetAccuracyBySubject(ctx context.Context) ([]database.GetAccuracyBySubjectRow, error) {
	return r.q.GetAccuracyBySubject(ctx)
}

func (r *SQLAnalyticsRepository) GetAccuracyByTopic(ctx context.Context, subjectID string) ([]database.GetAccuracyByTopicRow, error) {
	return r.q.GetAccuracyByTopic(ctx, subjectID)
}

func (r *SQLAnalyticsRepository) GetActivityHeatmap(ctx context.Context, daysCount string) ([]database.GetActivityHeatmapRow, error) {
	return r.q.GetActivityHeatmap(ctx, daysCount)
}

```

--- 

**File:** `internal/repository/cycle_item_repository.go`

```typescript
package repository

import (
	"context"

	"github.com/joaoapaenas/my-api/internal/database"
)

type CycleItemRepository interface {
	CreateCycleItem(ctx context.Context, arg database.CreateCycleItemParams) (database.CycleItem, error)
	ListCycleItems(ctx context.Context, cycleID string) ([]database.CycleItem, error)
	GetCycleItem(ctx context.Context, id string) (database.CycleItem, error)
	UpdateCycleItem(ctx context.Context, arg database.UpdateCycleItemParams) error
	DeleteCycleItem(ctx context.Context, id string) error
}

type SQLCycleItemRepository struct {
	q database.Querier
}

func NewSQLCycleItemRepository(q database.Querier) *SQLCycleItemRepository {
	return &SQLCycleItemRepository{q: q}
}

func (r *SQLCycleItemRepository) CreateCycleItem(ctx context.Context, arg database.CreateCycleItemParams) (database.CycleItem, error) {
	return r.q.CreateCycleItem(ctx, arg)
}

func (r *SQLCycleItemRepository) ListCycleItems(ctx context.Context, cycleID string) ([]database.CycleItem, error) {
	return r.q.ListCycleItems(ctx, cycleID)
}

func (r *SQLCycleItemRepository) GetCycleItem(ctx context.Context, id string) (database.CycleItem, error) {
	return r.q.GetCycleItem(ctx, id)
}

func (r *SQLCycleItemRepository) UpdateCycleItem(ctx context.Context, arg database.UpdateCycleItemParams) error {
	return r.q.UpdateCycleItem(ctx, arg)
}

func (r *SQLCycleItemRepository) DeleteCycleItem(ctx context.Context, id string) error {
	return r.q.DeleteCycleItem(ctx, id)
}

```

--- 

**File:** `internal/repository/exercise_log_repository.go`

```typescript
package repository

import (
	"context"

	"github.com/joaoapaenas/my-api/internal/database"
)

type ExerciseLogRepository interface {
	CreateExerciseLog(ctx context.Context, arg database.CreateExerciseLogParams) (database.ExerciseLog, error)
	GetExerciseLog(ctx context.Context, id string) (database.ExerciseLog, error)
	DeleteExerciseLog(ctx context.Context, id string) error
}

type SQLExerciseLogRepository struct {
	q database.Querier
}

func NewSQLExerciseLogRepository(q database.Querier) *SQLExerciseLogRepository {
	return &SQLExerciseLogRepository{q: q}
}

func (r *SQLExerciseLogRepository) CreateExerciseLog(ctx context.Context, arg database.CreateExerciseLogParams) (database.ExerciseLog, error) {
	return r.q.CreateExerciseLog(ctx, arg)
}

func (r *SQLExerciseLogRepository) GetExerciseLog(ctx context.Context, id string) (database.ExerciseLog, error) {
	return r.q.GetExerciseLog(ctx, id)
}

func (r *SQLExerciseLogRepository) DeleteExerciseLog(ctx context.Context, id string) error {
	return r.q.DeleteExerciseLog(ctx, id)
}

```

--- 

**File:** `internal/repository/session_pause_repository.go`

```typescript
package repository

import (
	"context"

	"github.com/joaoapaenas/my-api/internal/database"
)

type SessionPauseRepository interface {
	CreateSessionPause(ctx context.Context, arg database.CreateSessionPauseParams) (database.SessionPause, error)
	EndSessionPause(ctx context.Context, arg database.EndSessionPauseParams) error
	GetSessionPause(ctx context.Context, id string) (database.SessionPause, error)
	DeleteSessionPause(ctx context.Context, id string) error
}

type SQLSessionPauseRepository struct {
	q database.Querier
}

func NewSQLSessionPauseRepository(q database.Querier) *SQLSessionPauseRepository {
	return &SQLSessionPauseRepository{q: q}
}

func (r *SQLSessionPauseRepository) CreateSessionPause(ctx context.Context, arg database.CreateSessionPauseParams) (database.SessionPause, error) {
	return r.q.CreateSessionPause(ctx, arg)
}

func (r *SQLSessionPauseRepository) EndSessionPause(ctx context.Context, arg database.EndSessionPauseParams) error {
	return r.q.EndSessionPause(ctx, arg)
}

func (r *SQLSessionPauseRepository) GetSessionPause(ctx context.Context, id string) (database.SessionPause, error) {
	return r.q.GetSessionPause(ctx, id)
}

func (r *SQLSessionPauseRepository) DeleteSessionPause(ctx context.Context, id string) error {
	return r.q.DeleteSessionPause(ctx, id)
}

```

--- 

**File:** `internal/repository/study_cycle_repository.go`

```typescript
package repository

import (
	"context"

	"github.com/joaoapaenas/my-api/internal/database"
)

type StudyCycleRepository interface {
	CreateStudyCycle(ctx context.Context, arg database.CreateStudyCycleParams) (database.StudyCycle, error)
	GetActiveStudyCycle(ctx context.Context) (database.StudyCycle, error)
	GetStudyCycle(ctx context.Context, id string) (database.StudyCycle, error)
	UpdateStudyCycle(ctx context.Context, arg database.UpdateStudyCycleParams) error
	DeleteStudyCycle(ctx context.Context, id string) error
	GetActiveCycleWithItems(ctx context.Context) ([]database.GetActiveCycleWithItemsRow, error)
}

type SQLStudyCycleRepository struct {
	q database.Querier
}

func NewSQLStudyCycleRepository(q database.Querier) *SQLStudyCycleRepository {
	return &SQLStudyCycleRepository{q: q}
}

func (r *SQLStudyCycleRepository) CreateStudyCycle(ctx context.Context, arg database.CreateStudyCycleParams) (database.StudyCycle, error) {
	return r.q.CreateStudyCycle(ctx, arg)
}

func (r *SQLStudyCycleRepository) GetActiveStudyCycle(ctx context.Context) (database.StudyCycle, error) {
	return r.q.GetActiveStudyCycle(ctx)
}

func (r *SQLStudyCycleRepository) GetStudyCycle(ctx context.Context, id string) (database.StudyCycle, error) {
	return r.q.GetStudyCycle(ctx, id)
}

func (r *SQLStudyCycleRepository) UpdateStudyCycle(ctx context.Context, arg database.UpdateStudyCycleParams) error {
	return r.q.UpdateStudyCycle(ctx, arg)
}

func (r *SQLStudyCycleRepository) DeleteStudyCycle(ctx context.Context, id string) error {
	return r.q.DeleteStudyCycle(ctx, id)
}

func (r *SQLStudyCycleRepository) GetActiveCycleWithItems(ctx context.Context) ([]database.GetActiveCycleWithItemsRow, error) {
	return r.q.GetActiveCycleWithItems(ctx)
}

```

--- 

**File:** `internal/repository/study_session_repository.go`

```typescript
package repository

import (
	"context"

	"github.com/joaoapaenas/my-api/internal/database"
)

type StudySessionRepository interface {
	CreateStudySession(ctx context.Context, arg database.CreateStudySessionParams) (database.StudySession, error)
	UpdateSessionDuration(ctx context.Context, arg database.UpdateSessionDurationParams) error
	GetStudySession(ctx context.Context, id string) (database.StudySession, error)
	DeleteStudySession(ctx context.Context, id string) error
	GetOpenSession(ctx context.Context) (database.GetOpenSessionRow, error)
}

type SQLStudySessionRepository struct {
	q database.Querier
}

func NewSQLStudySessionRepository(q database.Querier) *SQLStudySessionRepository {
	return &SQLStudySessionRepository{q: q}
}

func (r *SQLStudySessionRepository) CreateStudySession(ctx context.Context, arg database.CreateStudySessionParams) (database.StudySession, error) {
	return r.q.CreateStudySession(ctx, arg)
}

func (r *SQLStudySessionRepository) UpdateSessionDuration(ctx context.Context, arg database.UpdateSessionDurationParams) error {
	return r.q.UpdateSessionDuration(ctx, arg)
}

func (r *SQLStudySessionRepository) GetStudySession(ctx context.Context, id string) (database.StudySession, error) {
	return r.q.GetStudySession(ctx, id)
}

func (r *SQLStudySessionRepository) DeleteStudySession(ctx context.Context, id string) error {
	return r.q.DeleteStudySession(ctx, id)
}

func (r *SQLStudySessionRepository) GetOpenSession(ctx context.Context) (database.GetOpenSessionRow, error) {
	return r.q.GetOpenSession(ctx)
}

```

--- 

**File:** `internal/repository/subject_repository.go`

```typescript
package repository

import (
	"context"

	"github.com/joaoapaenas/my-api/internal/database"
)

type SubjectRepository interface {
	CreateSubject(ctx context.Context, arg database.CreateSubjectParams) (database.Subject, error)
	ListSubjects(ctx context.Context) ([]database.Subject, error)
	GetSubject(ctx context.Context, id string) (database.Subject, error)
	UpdateSubject(ctx context.Context, arg database.UpdateSubjectParams) error
	DeleteSubject(ctx context.Context, id string) error
}

type SQLSubjectRepository struct {
	q database.Querier
}

func NewSQLSubjectRepository(q database.Querier) *SQLSubjectRepository {
	return &SQLSubjectRepository{q: q}
}

func (r *SQLSubjectRepository) CreateSubject(ctx context.Context, arg database.CreateSubjectParams) (database.Subject, error) {
	return r.q.CreateSubject(ctx, arg)
}

func (r *SQLSubjectRepository) ListSubjects(ctx context.Context) ([]database.Subject, error) {
	return r.q.ListSubjects(ctx)
}

func (r *SQLSubjectRepository) GetSubject(ctx context.Context, id string) (database.Subject, error) {
	return r.q.GetSubject(ctx, id)
}

func (r *SQLSubjectRepository) UpdateSubject(ctx context.Context, arg database.UpdateSubjectParams) error {
	return r.q.UpdateSubject(ctx, arg)
}

func (r *SQLSubjectRepository) DeleteSubject(ctx context.Context, id string) error {
	return r.q.DeleteSubject(ctx, id)
}

```

--- 

**File:** `internal/repository/topic_repository.go`

```typescript
package repository

import (
	"context"

	"github.com/joaoapaenas/my-api/internal/database"
)

type TopicRepository interface {
	CreateTopic(ctx context.Context, arg database.CreateTopicParams) (database.Topic, error)
	ListTopicsBySubject(ctx context.Context, subjectID string) ([]database.Topic, error)
	GetTopic(ctx context.Context, id string) (database.Topic, error)
	UpdateTopic(ctx context.Context, arg database.UpdateTopicParams) error
	DeleteTopic(ctx context.Context, id string) error
}

type SQLTopicRepository struct {
	q database.Querier
}

func NewSQLTopicRepository(q database.Querier) *SQLTopicRepository {
	return &SQLTopicRepository{q: q}
}

func (r *SQLTopicRepository) CreateTopic(ctx context.Context, arg database.CreateTopicParams) (database.Topic, error) {
	return r.q.CreateTopic(ctx, arg)
}

func (r *SQLTopicRepository) ListTopicsBySubject(ctx context.Context, subjectID string) ([]database.Topic, error) {
	return r.q.ListTopicsBySubject(ctx, subjectID)
}

func (r *SQLTopicRepository) GetTopic(ctx context.Context, id string) (database.Topic, error) {
	return r.q.GetTopic(ctx, id)
}

func (r *SQLTopicRepository) UpdateTopic(ctx context.Context, arg database.UpdateTopicParams) error {
	return r.q.UpdateTopic(ctx, arg)
}

func (r *SQLTopicRepository) DeleteTopic(ctx context.Context, id string) error {
	return r.q.DeleteTopic(ctx, id)
}

```

--- 

**File:** `internal/repository/user_repository.go`

```typescript
package repository

import (
	"context"

	"github.com/joaoapaenas/my-api/internal/database"
)

type UserRepository interface {
	CreateUser(ctx context.Context, arg database.CreateUserParams) (database.User, error)
	GetUserByEmail(ctx context.Context, email string) (database.User, error)
}

type SQLUserRepository struct {
	q database.Querier
}

func NewSQLUserRepository(q database.Querier) *SQLUserRepository {
	return &SQLUserRepository{q: q}
}

func (r *SQLUserRepository) CreateUser(ctx context.Context, arg database.CreateUserParams) (database.User, error) {
	return r.q.CreateUser(ctx, arg)
}

func (r *SQLUserRepository) GetUserByEmail(ctx context.Context, email string) (database.User, error) {
	return r.q.GetUserByEmail(ctx, email)
}

```

--- 

**File:** `internal/service/analytics_service.go`

```typescript
package service

import (
	"context"
	"fmt"

	"github.com/joaoapaenas/my-api/internal/database"
	"github.com/joaoapaenas/my-api/internal/repository"
)

type AnalyticsService interface {
	GetTimeReport(ctx context.Context, startDateFrom, startDateTo string) ([]database.GetTimeReportBySubjectRow, error)
	GetGlobalAccuracy(ctx context.Context) ([]database.GetAccuracyBySubjectRow, error)
	GetWeakPoints(ctx context.Context, subjectID string) ([]database.GetAccuracyByTopicRow, error)
	GetHeatmap(ctx context.Context, daysCount int64) ([]database.GetActivityHeatmapRow, error)
}

type AnalyticsManager struct {
	repo repository.AnalyticsRepository
}

func NewAnalyticsManager(repo repository.AnalyticsRepository) *AnalyticsManager {
	return &AnalyticsManager{repo: repo}
}

func (s *AnalyticsManager) GetTimeReport(ctx context.Context, startDateFrom, startDateTo string) ([]database.GetTimeReportBySubjectRow, error) {
	return s.repo.GetTimeReportBySubject(ctx, database.GetTimeReportBySubjectParams{
		StartDateFrom: startDateFrom,
		StartDateTo:   startDateTo,
	})
}

func (s *AnalyticsManager) GetGlobalAccuracy(ctx context.Context) ([]database.GetAccuracyBySubjectRow, error) {
	return s.repo.GetAccuracyBySubject(ctx)
}

func (s *AnalyticsManager) GetWeakPoints(ctx context.Context, subjectID string) ([]database.GetAccuracyByTopicRow, error) {
	return s.repo.GetAccuracyByTopic(ctx, subjectID)
}

func (s *AnalyticsManager) GetHeatmap(ctx context.Context, daysCount int64) ([]database.GetActivityHeatmapRow, error) {
	// Default to 30 days if 0 or negative
	if daysCount <= 0 {
		daysCount = 30
	}
	return s.repo.GetActivityHeatmap(ctx, fmt.Sprintf("%d", daysCount))
}

```

--- 

**File:** `internal/service/cycle_item_service.go`

```typescript
package service

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/joaoapaenas/my-api/internal/database"
	"github.com/joaoapaenas/my-api/internal/repository"
)

type CycleItemService interface {
	CreateCycleItem(ctx context.Context, cycleID, subjectID string, orderIndex int, plannedDurationMinutes int) (database.CycleItem, error)
	ListCycleItems(ctx context.Context, cycleID string) ([]database.CycleItem, error)
	GetCycleItem(ctx context.Context, id string) (database.CycleItem, error)
	UpdateCycleItem(ctx context.Context, id, subjectID string, orderIndex int, plannedDurationMinutes int) error
	DeleteCycleItem(ctx context.Context, id string) error
}

type CycleItemManager struct {
	repo repository.CycleItemRepository
}

func NewCycleItemManager(repo repository.CycleItemRepository) *CycleItemManager {
	return &CycleItemManager{repo: repo}
}

func (s *CycleItemManager) CreateCycleItem(ctx context.Context, cycleID, subjectID string, orderIndex int, plannedDurationMinutes int) (database.CycleItem, error) {
	id := uuid.New().String()

	var duration sql.NullInt64
	if plannedDurationMinutes > 0 {
		duration = sql.NullInt64{Int64: int64(plannedDurationMinutes), Valid: true}
	}

	return s.repo.CreateCycleItem(ctx, database.CreateCycleItemParams{
		ID:                     id,
		CycleID:                cycleID,
		SubjectID:              subjectID,
		OrderIndex:             int64(orderIndex),
		PlannedDurationMinutes: duration,
	})
}

func (s *CycleItemManager) ListCycleItems(ctx context.Context, cycleID string) ([]database.CycleItem, error) {
	return s.repo.ListCycleItems(ctx, cycleID)
}

func (s *CycleItemManager) GetCycleItem(ctx context.Context, id string) (database.CycleItem, error) {
	return s.repo.GetCycleItem(ctx, id)
}

func (s *CycleItemManager) UpdateCycleItem(ctx context.Context, id, subjectID string, orderIndex int, plannedDurationMinutes int) error {
	var duration sql.NullInt64
	if plannedDurationMinutes > 0 {
		duration = sql.NullInt64{Int64: int64(plannedDurationMinutes), Valid: true}
	}

	return s.repo.UpdateCycleItem(ctx, database.UpdateCycleItemParams{
		SubjectID:              subjectID,
		OrderIndex:             int64(orderIndex),
		PlannedDurationMinutes: duration,
		ID:                     id,
	})
}

func (s *CycleItemManager) DeleteCycleItem(ctx context.Context, id string) error {
	return s.repo.DeleteCycleItem(ctx, id)
}

```

--- 

**File:** `internal/service/cycle_item_service_test.go`

```typescript
package service_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/joaoapaenas/my-api/internal/database"
	"github.com/joaoapaenas/my-api/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockCycleItemRepository is a mock implementation of repository.CycleItemRepository
type MockCycleItemRepository struct {
	mock.Mock
}

func (m *MockCycleItemRepository) CreateCycleItem(ctx context.Context, arg database.CreateCycleItemParams) (database.CycleItem, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(database.CycleItem), args.Error(1)
}

func (m *MockCycleItemRepository) ListCycleItems(ctx context.Context, cycleID string) ([]database.CycleItem, error) {
	args := m.Called(ctx, cycleID)
	return args.Get(0).([]database.CycleItem), args.Error(1)
}

func (m *MockCycleItemRepository) GetCycleItem(ctx context.Context, id string) (database.CycleItem, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(database.CycleItem), args.Error(1)
}

func (m *MockCycleItemRepository) UpdateCycleItem(ctx context.Context, arg database.UpdateCycleItemParams) error {
	args := m.Called(ctx, arg)
	return args.Error(0)
}

func (m *MockCycleItemRepository) DeleteCycleItem(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func TestCycleItemManager_CreateCycleItem(t *testing.T) {
	mockRepo := new(MockCycleItemRepository)
	svc := service.NewCycleItemManager(mockRepo)

	ctx := context.Background()
	cycleID := "cycle-uuid"
	subjectID := "subject-uuid"
	orderIndex := 1
	plannedDuration := 60

	mockRepo.On("CreateCycleItem", ctx, mock.MatchedBy(func(arg database.CreateCycleItemParams) bool {
		return arg.CycleID == cycleID && arg.SubjectID == subjectID && arg.OrderIndex == int64(orderIndex) && arg.PlannedDurationMinutes.Int64 == int64(plannedDuration)
	})).Return(database.CycleItem{
		ID:                     "item-uuid",
		CycleID:                cycleID,
		SubjectID:              subjectID,
		OrderIndex:             int64(orderIndex),
		PlannedDurationMinutes: sql.NullInt64{Int64: int64(plannedDuration), Valid: true},
	}, nil)

	item, err := svc.CreateCycleItem(ctx, cycleID, subjectID, orderIndex, plannedDuration)

	assert.NoError(t, err)
	assert.Equal(t, int64(orderIndex), item.OrderIndex)
	assert.Equal(t, int64(plannedDuration), item.PlannedDurationMinutes.Int64)
	mockRepo.AssertExpectations(t)
}

func TestCycleItemManager_ListCycleItems(t *testing.T) {
	mockRepo := new(MockCycleItemRepository)
	svc := service.NewCycleItemManager(mockRepo)

	ctx := context.Background()
	cycleID := "cycle-uuid"
	expectedItems := []database.CycleItem{
		{ID: "1", CycleID: cycleID, OrderIndex: 1},
		{ID: "2", CycleID: cycleID, OrderIndex: 2},
	}

	mockRepo.On("ListCycleItems", ctx, cycleID).Return(expectedItems, nil)

	items, err := svc.ListCycleItems(ctx, cycleID)

	assert.NoError(t, err)
	assert.Len(t, items, 2)
	assert.Equal(t, int64(1), items[0].OrderIndex)
	mockRepo.AssertExpectations(t)
}

```

--- 

**File:** `internal/service/exercise_log_service.go`

```typescript
package service

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/joaoapaenas/my-api/internal/database"
	"github.com/joaoapaenas/my-api/internal/repository"
)

type ExerciseLogService interface {
	CreateExerciseLog(ctx context.Context, sessionID, subjectID, topicID string, questionsCount, correctCount int) (database.ExerciseLog, error)
	GetExerciseLog(ctx context.Context, id string) (database.ExerciseLog, error)
	DeleteExerciseLog(ctx context.Context, id string) error
}

type ExerciseLogManager struct {
	repo repository.ExerciseLogRepository
}

func NewExerciseLogManager(repo repository.ExerciseLogRepository) *ExerciseLogManager {
	return &ExerciseLogManager{repo: repo}
}

func (s *ExerciseLogManager) CreateExerciseLog(ctx context.Context, sessionID, subjectID, topicID string, questionsCount, correctCount int) (database.ExerciseLog, error) {
	id := uuid.New().String()

	var session sql.NullString
	if sessionID != "" {
		session = sql.NullString{String: sessionID, Valid: true}
	}

	var topic sql.NullString
	if topicID != "" {
		topic = sql.NullString{String: topicID, Valid: true}
	}

	return s.repo.CreateExerciseLog(ctx, database.CreateExerciseLogParams{
		ID:             id,
		SessionID:      session,
		SubjectID:      subjectID,
		TopicID:        topic,
		QuestionsCount: int64(questionsCount),
		CorrectCount:   int64(correctCount),
	})
}

func (s *ExerciseLogManager) GetExerciseLog(ctx context.Context, id string) (database.ExerciseLog, error) {
	return s.repo.GetExerciseLog(ctx, id)
}

func (s *ExerciseLogManager) DeleteExerciseLog(ctx context.Context, id string) error {
	return s.repo.DeleteExerciseLog(ctx, id)
}

```

--- 

**File:** `internal/service/session_pause_service.go`

```typescript
package service

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/joaoapaenas/my-api/internal/database"
	"github.com/joaoapaenas/my-api/internal/repository"
)

type SessionPauseService interface {
	CreateSessionPause(ctx context.Context, sessionID, startedAt string) (database.SessionPause, error)
	EndSessionPause(ctx context.Context, id, endedAt string) error
	GetSessionPause(ctx context.Context, id string) (database.SessionPause, error)
	DeleteSessionPause(ctx context.Context, id string) error
}

type SessionPauseManager struct {
	repo repository.SessionPauseRepository
}

func NewSessionPauseManager(repo repository.SessionPauseRepository) *SessionPauseManager {
	return &SessionPauseManager{repo: repo}
}

func (s *SessionPauseManager) CreateSessionPause(ctx context.Context, sessionID, startedAt string) (database.SessionPause, error) {
	id := uuid.New().String()
	return s.repo.CreateSessionPause(ctx, database.CreateSessionPauseParams{
		ID:        id,
		SessionID: sessionID,
		StartedAt: startedAt,
	})
}

func (s *SessionPauseManager) EndSessionPause(ctx context.Context, id, endedAt string) error {
	var ended sql.NullString
	if endedAt != "" {
		ended = sql.NullString{String: endedAt, Valid: true}
	}

	return s.repo.EndSessionPause(ctx, database.EndSessionPauseParams{
		EndedAt: ended,
		ID:      id,
	})
}

func (s *SessionPauseManager) GetSessionPause(ctx context.Context, id string) (database.SessionPause, error) {
	return s.repo.GetSessionPause(ctx, id)
}

func (s *SessionPauseManager) DeleteSessionPause(ctx context.Context, id string) error {
	return s.repo.DeleteSessionPause(ctx, id)
}

```

--- 

**File:** `internal/service/study_cycle_service.go`

```typescript
package service

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/joaoapaenas/my-api/internal/database"
	"github.com/joaoapaenas/my-api/internal/repository"
)

type StudyCycleService interface {
	CreateStudyCycle(ctx context.Context, name, description string, isActive bool) (database.StudyCycle, error)
	GetActiveStudyCycle(ctx context.Context) (database.StudyCycle, error)
	GetStudyCycle(ctx context.Context, id string) (database.StudyCycle, error)
	UpdateStudyCycle(ctx context.Context, id, name, description string, isActive bool) error
	DeleteStudyCycle(ctx context.Context, id string) error
	GetActiveCycleWithItems(ctx context.Context) ([]database.GetActiveCycleWithItemsRow, error)
}

type StudyCycleManager struct {
	repo repository.StudyCycleRepository
}

func NewStudyCycleManager(repo repository.StudyCycleRepository) *StudyCycleManager {
	return &StudyCycleManager{repo: repo}
}

func (s *StudyCycleManager) CreateStudyCycle(ctx context.Context, name, description string, isActive bool) (database.StudyCycle, error) {
	id := uuid.New().String()

	var desc sql.NullString
	if description != "" {
		desc = sql.NullString{String: description, Valid: true}
	}

	var active sql.NullInt64
	if isActive {
		active = sql.NullInt64{Int64: 1, Valid: true}
	} else {
		active = sql.NullInt64{Int64: 0, Valid: true}
	}

	return s.repo.CreateStudyCycle(ctx, database.CreateStudyCycleParams{
		ID:          id,
		Name:        name,
		Description: desc,
		IsActive:    active,
	})
}

func (s *StudyCycleManager) GetActiveStudyCycle(ctx context.Context) (database.StudyCycle, error) {
	return s.repo.GetActiveStudyCycle(ctx)
}

func (s *StudyCycleManager) GetStudyCycle(ctx context.Context, id string) (database.StudyCycle, error) {
	return s.repo.GetStudyCycle(ctx, id)
}

func (s *StudyCycleManager) UpdateStudyCycle(ctx context.Context, id, name, description string, isActive bool) error {
	var desc sql.NullString
	if description != "" {
		desc = sql.NullString{String: description, Valid: true}
	}

	var active sql.NullInt64
	if isActive {
		active = sql.NullInt64{Int64: 1, Valid: true}
	} else {
		active = sql.NullInt64{Int64: 0, Valid: true}
	}

	return s.repo.UpdateStudyCycle(ctx, database.UpdateStudyCycleParams{
		Name:        name,
		Description: desc,
		IsActive:    active,
		ID:          id,
	})
}

func (s *StudyCycleManager) DeleteStudyCycle(ctx context.Context, id string) error {
	return s.repo.DeleteStudyCycle(ctx, id)
}

func (s *StudyCycleManager) GetActiveCycleWithItems(ctx context.Context) ([]database.GetActiveCycleWithItemsRow, error) {
	return s.repo.GetActiveCycleWithItems(ctx)
}

```

--- 

**File:** `internal/service/study_cycle_service_test.go`

```typescript
package service_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/joaoapaenas/my-api/internal/database"
	"github.com/joaoapaenas/my-api/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockStudyCycleRepository is a mock implementation of repository.StudyCycleRepository
type MockStudyCycleRepository struct {
	mock.Mock
}

func (m *MockStudyCycleRepository) CreateStudyCycle(ctx context.Context, arg database.CreateStudyCycleParams) (database.StudyCycle, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(database.StudyCycle), args.Error(1)
}

func (m *MockStudyCycleRepository) GetActiveStudyCycle(ctx context.Context) (database.StudyCycle, error) {
	args := m.Called(ctx)
	return args.Get(0).(database.StudyCycle), args.Error(1)
}

func (m *MockStudyCycleRepository) GetStudyCycle(ctx context.Context, id string) (database.StudyCycle, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(database.StudyCycle), args.Error(1)
}

func (m *MockStudyCycleRepository) UpdateStudyCycle(ctx context.Context, arg database.UpdateStudyCycleParams) error {
	args := m.Called(ctx, arg)
	return args.Error(0)
}

func (m *MockStudyCycleRepository) DeleteStudyCycle(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockStudyCycleRepository) GetActiveCycleWithItems(ctx context.Context) ([]database.GetActiveCycleWithItemsRow, error) {
	args := m.Called(ctx)
	return args.Get(0).([]database.GetActiveCycleWithItemsRow), args.Error(1)
}

func TestStudyCycleManager_CreateStudyCycle(t *testing.T) {
	mockRepo := new(MockStudyCycleRepository)
	svc := service.NewStudyCycleManager(mockRepo)

	ctx := context.Background()
	name := "Cycle 1"
	description := "First Cycle"
	isActive := true

	mockRepo.On("CreateStudyCycle", ctx, mock.MatchedBy(func(arg database.CreateStudyCycleParams) bool {
		return arg.Name == name && arg.Description.String == description && arg.IsActive.Int64 == 1
	})).Return(database.StudyCycle{
		ID:          "cycle-uuid",
		Name:        name,
		Description: sql.NullString{String: description, Valid: true},
		IsActive:    sql.NullInt64{Int64: 1, Valid: true},
	}, nil)

	cycle, err := svc.CreateStudyCycle(ctx, name, description, isActive)

	assert.NoError(t, err)
	assert.Equal(t, name, cycle.Name)
	assert.Equal(t, int64(1), cycle.IsActive.Int64)
	mockRepo.AssertExpectations(t)
}

func TestStudyCycleManager_GetActiveStudyCycle(t *testing.T) {
	mockRepo := new(MockStudyCycleRepository)
	svc := service.NewStudyCycleManager(mockRepo)

	ctx := context.Background()
	expectedCycle := database.StudyCycle{ID: "active-uuid", Name: "Active Cycle", IsActive: sql.NullInt64{Int64: 1, Valid: true}}

	mockRepo.On("GetActiveStudyCycle", ctx).Return(expectedCycle, nil)

	cycle, err := svc.GetActiveStudyCycle(ctx)

	assert.NoError(t, err)
	assert.Equal(t, "Active Cycle", cycle.Name)
	mockRepo.AssertExpectations(t)
}

```

--- 

**File:** `internal/service/study_session_service.go`

```typescript
package service

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/joaoapaenas/my-api/internal/database"
	"github.com/joaoapaenas/my-api/internal/repository"
)

type StudySessionService interface {
	CreateStudySession(ctx context.Context, subjectID, cycleItemID, startedAt string) (database.StudySession, error)
	UpdateSessionDuration(ctx context.Context, id, finishedAt string, grossSeconds, netSeconds int, notes string) error
	GetStudySession(ctx context.Context, id string) (database.StudySession, error)
	DeleteStudySession(ctx context.Context, id string) error
	GetOpenSession(ctx context.Context) (database.GetOpenSessionRow, error)
}

type StudySessionManager struct {
	repo repository.StudySessionRepository
}

func NewStudySessionManager(repo repository.StudySessionRepository) *StudySessionManager {
	return &StudySessionManager{repo: repo}
}

func (s *StudySessionManager) CreateStudySession(ctx context.Context, subjectID, cycleItemID, startedAt string) (database.StudySession, error) {
	id := uuid.New().String()

	var cycleItem sql.NullString
	if cycleItemID != "" {
		cycleItem = sql.NullString{String: cycleItemID, Valid: true}
	}

	return s.repo.CreateStudySession(ctx, database.CreateStudySessionParams{
		ID:          id,
		SubjectID:   subjectID,
		CycleItemID: cycleItem,
		StartedAt:   startedAt,
	})
}

func (s *StudySessionManager) UpdateSessionDuration(ctx context.Context, id, finishedAt string, grossSeconds, netSeconds int, notes string) error {
	var finished sql.NullString
	if finishedAt != "" {
		finished = sql.NullString{String: finishedAt, Valid: true}
	}

	var gross sql.NullInt64
	if grossSeconds > 0 {
		gross = sql.NullInt64{Int64: int64(grossSeconds), Valid: true}
	}

	var net sql.NullInt64
	if netSeconds > 0 {
		net = sql.NullInt64{Int64: int64(netSeconds), Valid: true}
	}

	var sessionNotes sql.NullString
	if notes != "" {
		sessionNotes = sql.NullString{String: notes, Valid: true}
	}

	return s.repo.UpdateSessionDuration(ctx, database.UpdateSessionDurationParams{
		FinishedAt:           finished,
		GrossDurationSeconds: gross,
		NetDurationSeconds:   net,
		Notes:                sessionNotes,
		ID:                   id,
	})
}

func (s *StudySessionManager) GetStudySession(ctx context.Context, id string) (database.StudySession, error) {
	return s.repo.GetStudySession(ctx, id)
}

func (s *StudySessionManager) DeleteStudySession(ctx context.Context, id string) error {
	return s.repo.DeleteStudySession(ctx, id)
}

func (s *StudySessionManager) GetOpenSession(ctx context.Context) (database.GetOpenSessionRow, error) {
	return s.repo.GetOpenSession(ctx)
}

```

--- 

**File:** `internal/service/study_session_service_test.go`

```typescript
package service_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/joaoapaenas/my-api/internal/database"
	"github.com/joaoapaenas/my-api/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockStudySessionRepository is a mock implementation of repository.StudySessionRepository
type MockStudySessionRepository struct {
	mock.Mock
}

func (m *MockStudySessionRepository) CreateStudySession(ctx context.Context, arg database.CreateStudySessionParams) (database.StudySession, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(database.StudySession), args.Error(1)
}

func (m *MockStudySessionRepository) UpdateSessionDuration(ctx context.Context, arg database.UpdateSessionDurationParams) error {
	args := m.Called(ctx, arg)
	return args.Error(0)
}

func (m *MockStudySessionRepository) GetStudySession(ctx context.Context, id string) (database.StudySession, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(database.StudySession), args.Error(1)
}

func (m *MockStudySessionRepository) DeleteStudySession(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockStudySessionRepository) GetOpenSession(ctx context.Context) (database.GetOpenSessionRow, error) {
	args := m.Called(ctx)
	return args.Get(0).(database.GetOpenSessionRow), args.Error(1)
}

func TestStudySessionManager_CreateStudySession(t *testing.T) {
	mockRepo := new(MockStudySessionRepository)
	svc := service.NewStudySessionManager(mockRepo)

	ctx := context.Background()
	subjectID := "subject-uuid"
	cycleItemID := "item-uuid"
	startedAt := "2023-10-27T10:00:00Z"

	mockRepo.On("CreateStudySession", ctx, mock.MatchedBy(func(arg database.CreateStudySessionParams) bool {
		return arg.SubjectID == subjectID && arg.CycleItemID.String == cycleItemID && arg.StartedAt == startedAt
	})).Return(database.StudySession{
		ID:          "session-uuid",
		SubjectID:   subjectID,
		CycleItemID: sql.NullString{String: cycleItemID, Valid: true},
		StartedAt:   startedAt,
	}, nil)

	session, err := svc.CreateStudySession(ctx, subjectID, cycleItemID, startedAt)

	assert.NoError(t, err)
	assert.Equal(t, subjectID, session.SubjectID)
	assert.Equal(t, cycleItemID, session.CycleItemID.String)
	mockRepo.AssertExpectations(t)
}

func TestStudySessionManager_UpdateSessionDuration(t *testing.T) {
	mockRepo := new(MockStudySessionRepository)
	svc := service.NewStudySessionManager(mockRepo)

	ctx := context.Background()
	sessionID := "session-uuid"
	finishedAt := "2023-10-27T11:00:00Z"
	gross := 3600
	net := 3000
	notes := "Good session"

	mockRepo.On("UpdateSessionDuration", ctx, mock.MatchedBy(func(arg database.UpdateSessionDurationParams) bool {
		return arg.ID == sessionID && arg.FinishedAt.String == finishedAt && arg.GrossDurationSeconds.Int64 == int64(gross) && arg.NetDurationSeconds.Int64 == int64(net)
	})).Return(nil)

	err := svc.UpdateSessionDuration(ctx, sessionID, finishedAt, gross, net, notes)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

```

--- 

**File:** `internal/service/subject_service.go`

```typescript
package service

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/joaoapaenas/my-api/internal/database"
	"github.com/joaoapaenas/my-api/internal/repository"
)

type SubjectService interface {
	CreateSubject(ctx context.Context, name, colorHex string) (database.Subject, error)
	ListSubjects(ctx context.Context) ([]database.Subject, error)
	GetSubject(ctx context.Context, id string) (database.Subject, error)
	UpdateSubject(ctx context.Context, id, name, colorHex string) error
	DeleteSubject(ctx context.Context, id string) error
}

type SubjectManager struct {
	repo repository.SubjectRepository
}

func NewSubjectManager(repo repository.SubjectRepository) *SubjectManager {
	return &SubjectManager{repo: repo}
}

func (s *SubjectManager) CreateSubject(ctx context.Context, name, colorHex string) (database.Subject, error) {
	id := uuid.New().String()

	var color sql.NullString
	if colorHex != "" {
		color = sql.NullString{String: colorHex, Valid: true}
	}

	return s.repo.CreateSubject(ctx, database.CreateSubjectParams{
		ID:       id,
		Name:     name,
		ColorHex: color,
	})
}

func (s *SubjectManager) ListSubjects(ctx context.Context) ([]database.Subject, error) {
	return s.repo.ListSubjects(ctx)
}

func (s *SubjectManager) GetSubject(ctx context.Context, id string) (database.Subject, error) {
	return s.repo.GetSubject(ctx, id)
}

func (s *SubjectManager) UpdateSubject(ctx context.Context, id, name, colorHex string) error {
	var color sql.NullString
	if colorHex != "" {
		color = sql.NullString{String: colorHex, Valid: true}
	}

	return s.repo.UpdateSubject(ctx, database.UpdateSubjectParams{
		Name:     name,
		ColorHex: color,
		ID:       id,
	})
}

func (s *SubjectManager) DeleteSubject(ctx context.Context, id string) error {
	return s.repo.DeleteSubject(ctx, id)
}

```

--- 

**File:** `internal/service/subject_service_test.go`

```typescript
package service_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/joaoapaenas/my-api/internal/database"
	"github.com/joaoapaenas/my-api/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockSubjectRepository is a mock implementation of repository.SubjectRepository
type MockSubjectRepository struct {
	mock.Mock
}

func (m *MockSubjectRepository) CreateSubject(ctx context.Context, arg database.CreateSubjectParams) (database.Subject, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(database.Subject), args.Error(1)
}

func (m *MockSubjectRepository) ListSubjects(ctx context.Context) ([]database.Subject, error) {
	args := m.Called(ctx)
	return args.Get(0).([]database.Subject), args.Error(1)
}

func (m *MockSubjectRepository) GetSubject(ctx context.Context, id string) (database.Subject, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(database.Subject), args.Error(1)
}

func (m *MockSubjectRepository) UpdateSubject(ctx context.Context, arg database.UpdateSubjectParams) error {
	args := m.Called(ctx, arg)
	return args.Error(0)
}

func (m *MockSubjectRepository) DeleteSubject(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func TestSubjectManager_CreateSubject(t *testing.T) {
	mockRepo := new(MockSubjectRepository)
	svc := service.NewSubjectManager(mockRepo)

	ctx := context.Background()
	name := "Mathematics"
	colorHex := "#FF0000"

	mockRepo.On("CreateSubject", ctx, mock.MatchedBy(func(arg database.CreateSubjectParams) bool {
		return arg.Name == name && arg.ColorHex.String == colorHex && arg.ColorHex.Valid
	})).Return(database.Subject{
		ID:       "uuid",
		Name:     name,
		ColorHex: sql.NullString{String: colorHex, Valid: true},
	}, nil)

	subject, err := svc.CreateSubject(ctx, name, colorHex)

	assert.NoError(t, err)
	assert.Equal(t, name, subject.Name)
	assert.Equal(t, colorHex, subject.ColorHex.String)
	mockRepo.AssertExpectations(t)
}

func TestSubjectManager_ListSubjects(t *testing.T) {
	mockRepo := new(MockSubjectRepository)
	svc := service.NewSubjectManager(mockRepo)

	ctx := context.Background()
	expectedSubjects := []database.Subject{
		{ID: "1", Name: "Math"},
		{ID: "2", Name: "Physics"},
	}

	mockRepo.On("ListSubjects", ctx).Return(expectedSubjects, nil)

	subjects, err := svc.ListSubjects(ctx)

	assert.NoError(t, err)
	assert.Len(t, subjects, 2)
	assert.Equal(t, "Math", subjects[0].Name)
	mockRepo.AssertExpectations(t)
}

```

--- 

**File:** `internal/service/topic_service.go`

```typescript
package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/joaoapaenas/my-api/internal/database"
	"github.com/joaoapaenas/my-api/internal/repository"
)

type TopicService interface {
	CreateTopic(ctx context.Context, subjectID, name string) (database.Topic, error)
	ListTopicsBySubject(ctx context.Context, subjectID string) ([]database.Topic, error)
	GetTopic(ctx context.Context, id string) (database.Topic, error)
	UpdateTopic(ctx context.Context, id, name string) error
	DeleteTopic(ctx context.Context, id string) error
}

type TopicManager struct {
	repo repository.TopicRepository
}

func NewTopicManager(repo repository.TopicRepository) *TopicManager {
	return &TopicManager{repo: repo}
}

func (s *TopicManager) CreateTopic(ctx context.Context, subjectID, name string) (database.Topic, error) {
	id := uuid.New().String()
	return s.repo.CreateTopic(ctx, database.CreateTopicParams{
		ID:        id,
		SubjectID: subjectID,
		Name:      name,
	})
}

func (s *TopicManager) ListTopicsBySubject(ctx context.Context, subjectID string) ([]database.Topic, error) {
	return s.repo.ListTopicsBySubject(ctx, subjectID)
}

func (s *TopicManager) GetTopic(ctx context.Context, id string) (database.Topic, error) {
	return s.repo.GetTopic(ctx, id)
}

func (s *TopicManager) UpdateTopic(ctx context.Context, id, name string) error {
	return s.repo.UpdateTopic(ctx, database.UpdateTopicParams{
		Name: name,
		ID:   id,
	})
}

func (s *TopicManager) DeleteTopic(ctx context.Context, id string) error {
	return s.repo.DeleteTopic(ctx, id)
}

```

--- 

**File:** `internal/service/topic_service_test.go`

```typescript
package service_test

import (
	"context"
	"testing"

	"github.com/joaoapaenas/my-api/internal/database"
	"github.com/joaoapaenas/my-api/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockTopicRepository is a mock implementation of repository.TopicRepository
type MockTopicRepository struct {
	mock.Mock
}

func (m *MockTopicRepository) CreateTopic(ctx context.Context, arg database.CreateTopicParams) (database.Topic, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(database.Topic), args.Error(1)
}

func (m *MockTopicRepository) ListTopicsBySubject(ctx context.Context, subjectID string) ([]database.Topic, error) {
	args := m.Called(ctx, subjectID)
	return args.Get(0).([]database.Topic), args.Error(1)
}

func (m *MockTopicRepository) GetTopic(ctx context.Context, id string) (database.Topic, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(database.Topic), args.Error(1)
}

func (m *MockTopicRepository) UpdateTopic(ctx context.Context, arg database.UpdateTopicParams) error {
	args := m.Called(ctx, arg)
	return args.Error(0)
}

func (m *MockTopicRepository) DeleteTopic(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func TestTopicManager_CreateTopic(t *testing.T) {
	mockRepo := new(MockTopicRepository)
	svc := service.NewTopicManager(mockRepo)

	ctx := context.Background()
	subjectID := "subject-uuid"
	name := "Algebra"

	mockRepo.On("CreateTopic", ctx, mock.MatchedBy(func(arg database.CreateTopicParams) bool {
		return arg.SubjectID == subjectID && arg.Name == name
	})).Return(database.Topic{
		ID:        "topic-uuid",
		SubjectID: subjectID,
		Name:      name,
	}, nil)

	topic, err := svc.CreateTopic(ctx, subjectID, name)

	assert.NoError(t, err)
	assert.Equal(t, name, topic.Name)
	assert.Equal(t, subjectID, topic.SubjectID)
	mockRepo.AssertExpectations(t)
}

func TestTopicManager_GetTopic(t *testing.T) {
	mockRepo := new(MockTopicRepository)
	svc := service.NewTopicManager(mockRepo)

	ctx := context.Background()
	topicID := "topic-uuid"
	expectedTopic := database.Topic{ID: topicID, Name: "Algebra"}

	mockRepo.On("GetTopic", ctx, topicID).Return(expectedTopic, nil)

	topic, err := svc.GetTopic(ctx, topicID)

	assert.NoError(t, err)
	assert.Equal(t, expectedTopic.Name, topic.Name)
	mockRepo.AssertExpectations(t)
}

```

--- 

**File:** `internal/service/user_service.go`

```typescript
package service

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/joaoapaenas/my-api/internal/database"
	"github.com/joaoapaenas/my-api/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserNotFound = errors.New("user not found")
	ErrEmailTaken   = errors.New("email already taken")
)

type UserService interface {
	CreateUser(ctx context.Context, email, name, password string) (database.User, error)
	GetUserByEmail(ctx context.Context, email string) (database.User, error)
}

// UserManager implements UserService
type UserManager struct {
	repo repository.UserRepository
}

func NewUserManager(repo repository.UserRepository) *UserManager {
	return &UserManager{repo: repo}
}

func (s *UserManager) CreateUser(ctx context.Context, email, name, password string) (database.User, error) {
	// Logic: Generate UUID here, not in the handler
	id := uuid.New().String()

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return database.User{}, err
	}

	user, err := s.repo.CreateUser(ctx, database.CreateUserParams{
		ID:           id,
		Email:        email,
		Name:         name,
		PasswordHash: string(hashedPassword),
	})
	if err != nil {
		// In a real app, check for specific DB errors (like unique constraint violation)
		// and return ErrEmailTaken. For now, we return the raw error.
		return database.User{}, err
	}
	return user, nil
}

func (s *UserManager) GetUserByEmail(ctx context.Context, email string) (database.User, error) {
	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		// Assuming standard sql.ErrNoRows check happens here or in repo
		// Ideally, you map sql.ErrNoRows -> ErrUserNotFound here
		return database.User{}, err
	}
	return user, nil
}

```

--- 

**File:** `internal/service/user_service_test.go`

```typescript
package service_test

import (
	"context"
	"testing"

	"github.com/joaoapaenas/my-api/internal/database"
	"github.com/joaoapaenas/my-api/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
)

// MockUserRepository is a mock implementation of repository.UserRepository
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) CreateUser(ctx context.Context, arg database.CreateUserParams) (database.User, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(database.User), args.Error(1)
}

func (m *MockUserRepository) GetUserByEmail(ctx context.Context, email string) (database.User, error) {
	args := m.Called(ctx, email)
	return args.Get(0).(database.User), args.Error(1)
}

func TestUserManager_CreateUser(t *testing.T) {
	mockRepo := new(MockUserRepository)
	svc := service.NewUserManager(mockRepo)

	ctx := context.Background()
	email := "test@example.com"
	name := "Test User"
	password := "password123"

	mockRepo.On("CreateUser", ctx, mock.MatchedBy(func(arg database.CreateUserParams) bool {
		// Verify arguments
		if arg.Email != email || arg.Name != name {
			return false
		}
		// Verify password is hashed
		err := bcrypt.CompareHashAndPassword([]byte(arg.PasswordHash), []byte(password))
		return err == nil
	})).Return(database.User{
		ID:    "uuid",
		Email: email,
		Name:  name,
	}, nil)

	user, err := svc.CreateUser(ctx, email, name, password)

	assert.NoError(t, err)
	assert.Equal(t, email, user.Email)
	assert.Equal(t, name, user.Name)
	mockRepo.AssertExpectations(t)
}

```

