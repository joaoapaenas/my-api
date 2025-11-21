package tests

import (
	"context"
	"database/sql"
	"testing"

	"github.com/joaoapaenas/my-api/internal/database"
	_ "github.com/mattn/go-sqlite3"
)

func TestCreateUserFlow(t *testing.T) {
	// 1. Setup In-Memory DB
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	// 2. Apply Schema manually for test (or load .sql file)
	_, err = db.Exec(`CREATE TABLE users (id TEXT PRIMARY KEY, email TEXT, name TEXT, created_at DATETIME);`)
	if err != nil {
		t.Fatal(err)
	}

	q := database.New(db)

	// 3. Run Logic
	user, err := q.CreateUser(context.Background(), database.CreateUserParams{
		ID:    "123",
		Email: "test@example.com",
		Name:  "Tester",
	})

	// 4. Assertions
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if user.Email != "test@example.com" {
		t.Errorf("expected email test@example.com, got %s", user.Email)
	}
}
