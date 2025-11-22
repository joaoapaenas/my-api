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
	"log"
	"log/slog" // Use standard library structured logger
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
	"github.com/joaoapaenas/my-api/internal/logger" // Import your logger
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
	// 1. Load Config
	cfg, err := config.Load()
	if err != nil {
		// We use standard log here because logger isn't init'd yet
		log.Fatalf("Failed to load config: %v", err)
	}

	// 2. Initialize Logger
	logger.Init(cfg.Env)
	slog.Info("Starting application", "env", cfg.Env, "port", cfg.Port)

	// 2. Database Connection
	db, err := sql.Open("sqlite", cfg.DBUrl)
	if err != nil {
		slog.Error("Failed to open db", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		slog.Error("Failed to ping db", "error", err)
		os.Exit(1)
	}

	// 3. Wiring Layers
	queries := database.New(db)
	userService := service.NewUserManager(queries)
	userHandler := handler.NewUserHandler(userService)

	// 4. Router Setup
	r := chi.NewRouter()
	// middleware.Logger is okay for dev, but in prod we often rely on our own logging
	r.Use(middleware.RequestID) // Adds a request ID to context (crucial for tracing)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	// Documentation
	r.Get("/swagger/doc.json", func(w http.ResponseWriter, r *http.Request) {
		doc := docs.SwaggerInfo.ReadDoc()
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(doc))
	})
	r.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("http://localhost:8080/swagger/doc.json"),
	))

	// Routes
	r.Route("/users", func(r chi.Router) {
		r.Post("/", userHandler.CreateUser)
		r.Get("/{email}", userHandler.GetUser)
	})

	// 5. Graceful Shutdown Configuration
	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: r,
	}

	// Start Server in a Goroutine (so it doesn't block)
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("Server failed to start", "error", err)
			os.Exit(1)
		}
	}()
	slog.Info("Server is ready to handle requests")

	// 6. Wait for Interrupt Signal (Ctrl+C)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	// Block until signal is received
	sig := <-quit
	slog.Info("Shutting down server...", "signal", sig.String())

	// Create a timeout context (wait 10s for running requests to finish)
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
	"flag"
	"log"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	cmd := flag.String("cmd", "", "Command: up or down")
	flag.Parse()

	// 1. Open DB using the "sqlite" driver provided by the migration library (modernc)
	// We do not need to import glebarez here because 'database/sqlite' above
	// already registers a driver named "sqlite".
	db, err := sql.Open("sqlite", "dev.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// 2. Create Migration Driver instance
	driver, err := sqlite.WithInstance(db, &sqlite.Config{})
	if err != nil {
		log.Fatal(err)
	}

	// 3. Initialize Migrate
	m, err := migrate.NewWithDatabaseInstance(
		"file://sql/schema",
		"sqlite",
		driver,
	)
	if err != nil {
		log.Fatal(err)
	}

	// 4. Run Command
	switch *cmd {
	case "up":
		err := m.Up()
		if err != nil && err != migrate.ErrNoChange {
			log.Fatal(err)
		}
		log.Println("Migrated UP successfully!")
	case "down":
		err := m.Down()
		if err != nil && err != migrate.ErrNoChange {
			log.Fatal(err)
		}
		log.Println("Migrated DOWN successfully!")
	default:
		log.Fatal("Unknown command. Use -cmd=up or -cmd=down")
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
        "handler.CreateUserRequest": {
            "type": "object",
            "properties": {
                "email": {
                    "type": "string"
                },
                "name": {
                    "type": "string"
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
        "handler.CreateUserRequest": {
            "type": "object",
            "properties": {
                "email": {
                    "type": "string"
                },
                "name": {
                    "type": "string"
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
  handler.CreateUserRequest:
    properties:
      email:
        type: string
      name:
        type: string
    type: object
host: localhost:8080
info:
  contact: {}
  description: Production ready starter guide.
  title: My Go API
  version: "1.0"
paths:
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
	"path/filepath" // Add this
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
			// FIX: Resolve absolute path to avoid "unable to open" errors on Windows
			wd, _ := os.Getwd()
			dbPath := filepath.Join(wd, "dev.db")

			// FIX: Add Pragmas for Windows robustness
			// _pragma=busy_timeout(5000): Wait 5s if db is locked (fixes "database is locked")
			// _pragma=journal_mode(WAL): Better concurrency
			cfg.DBUrl = fmt.Sprintf("file:%s?_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)", dbPath)
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

**File:** `internal/database/models.go`

```typescript
// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.30.0

package database

import (
	"time"
)

type User struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
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
	CreateUser(ctx context.Context, arg CreateUserParams) (User, error)
	GetUserByEmail(ctx context.Context, email string) (User, error)
}

var _ Querier = (*Queries)(nil)

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
INSERT INTO users (id, email, name)
VALUES (?, ?, ?)
RETURNING id, email, name, created_at
`

type CreateUserParams struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
}

func (q *Queries) CreateUser(ctx context.Context, arg CreateUserParams) (User, error) {
	row := q.db.QueryRowContext(ctx, createUser, arg.ID, arg.Email, arg.Name)
	var i User
	err := row.Scan(
		&i.ID,
		&i.Email,
		&i.Name,
		&i.CreatedAt,
	)
	return i, err
}

const getUserByEmail = `-- name: GetUserByEmail :one
SELECT id, email, name, created_at FROM users
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
	)
	return i, err
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
	user, err := h.svc.CreateUser(r.Context(), req.Email, req.Name)
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

func formatValidationErrors(err error) map[string]string {
	errors := make(map[string]string)
	// Assert that it is a validator.ValidationErrors
	if validationErrs, ok := err.(validator.ValidationErrors); ok {
		for _, e := range validationErrs {
			// Simple message mapping
			switch e.Tag() {
			case "required":
				errors[e.Field()] = "This field is required"
			case "email":
				errors[e.Field()] = "Invalid email format"
			case "min":
				errors[e.Field()] = "Must be at least " + e.Param() + " characters"
			default:
				errors[e.Field()] = "Invalid value"
			}
		}
	}
	return errors
}

```

--- 

**File:** `internal/handler/user_handler_test.go`

```typescript
package handler

import (
	"net/http/httptest"
	"testing"
)

func TestGetUser_Validation(t *testing.T) {
	// This tests ONLY routing/basic logic, avoiding DB for simplicity in this snippet.
	// For DB mocking, you would mock the 'database.Querier' interface.

	tests := []struct {
		name       string
		url        string
		wantStatus int
	}{
		{
			name:       "Missing Email",
			url:        "/users/", // Chi might handle 404 here
			wantStatus: 404,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.url, nil)
			rr := httptest.NewRecorder()

			// Note: In a real test, inject a MockQuerier here
			h := NewUserHandler(nil)
			// We can't call the actual method without the mock,
			// so this serves as a structural example.
			_ = h
			_ = req
			_ = rr
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

**File:** `internal/service/user_service.go`

```typescript
package service

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/joaoapaenas/my-api/internal/database"
)

var (
	ErrUserNotFound = errors.New("user not found")
	ErrEmailTaken   = errors.New("email already taken")
)

// UserService defines the business logic behavior
type UserService interface {
	CreateUser(ctx context.Context, email, name string) (database.User, error)
	GetUserByEmail(ctx context.Context, email string) (database.User, error)
}

// UserManager implements UserService
type UserManager struct {
	repo database.Querier
}

func NewUserManager(repo database.Querier) *UserManager {
	return &UserManager{repo: repo}
}

func (s *UserManager) CreateUser(ctx context.Context, email, name string) (database.User, error) {
	// Logic: Generate UUID here, not in the handler
	id := uuid.New().String()

	user, err := s.repo.CreateUser(ctx, database.CreateUserParams{
		ID:    id,
		Email: email,
		Name:  name,
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

