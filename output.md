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

	r.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("http://localhost:8080/swagger/doc.json"),
	))

	r.Route("/users", func(r chi.Router) {
		r.Post("/", userHandler.CreateUser)
		r.Get("/{email}", userHandler.GetUser)
	})

	log.Printf("Server starting on :%s", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, r); err != nil {
		log.Fatal(err)
	}
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
	"os"
	"time"
)

type Config struct {
	Port    string
	DBUrl   string
	Env     string
	Timeout time.Duration
}

func Load() *Config {
	// In production, use a library like kelseyhightower/envconfig
	return &Config{
		Port:    getEnv("PORT", "8080"),
		DBUrl:   getEnv("DB_URL", "./dev.db"),
		Env:     getEnv("ENV", "development"),
		Timeout: 5 * time.Second,
	}
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
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/joaoapaenas/my-api/internal/database"
)

type UserHandler struct {
	repo database.Querier // Use the interface generated by sqlc
}

func NewUserHandler(repo database.Querier) *UserHandler {
	return &UserHandler{repo: repo}
}

type CreateUserRequest struct {
	Email string `json:"email"`
	Name  string `json:"name"`
}

// CreateUser godoc
// @Summary Create a new user
// @Description Create a user with email and name
// @Tags users
// @Accept json
// @Produce json
// @Param input body CreateUserRequest true "User info"
// @Success 201 {object} database.User
// @Router /users [post]
func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	id := uuid.New().String()

	// Always pass r.Context() for cancellation and timeouts
	user, err := h.repo.CreateUser(r.Context(), database.CreateUserParams{
		ID:    id,
		Email: req.Email,
		Name:  req.Name,
	})

	if err != nil {
		// Log the actual error internally here
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}

// GetUser godoc
// @Summary Get user by Email
// @Tags users
// @Param email path string true "User Email"
// @Success 200 {object} database.User
// @Router /users/{email} [get]
func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	email := chi.URLParam(r, "email")

	user, err := h.repo.GetUserByEmail(r.Context(), email)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(user)
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

