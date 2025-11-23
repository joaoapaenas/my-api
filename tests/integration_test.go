package tests

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/joaoapaenas/my-api/internal/database"
	"github.com/joaoapaenas/my-api/internal/handler"
	"github.com/joaoapaenas/my-api/internal/repository"
	"github.com/joaoapaenas/my-api/internal/service"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
)

func TestIntegration_CreateUserFlow(t *testing.T) {
	// 1. Setup In-Memory DB
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	// 2. Apply Schema
	_, err = db.Exec(`
		CREATE TABLE users (
			id TEXT PRIMARY KEY, 
			email TEXT, 
			name TEXT, 
			password_hash TEXT NOT NULL DEFAULT '',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);
	`)
	if err != nil {
		t.Fatal(err)
	}

	// 3. Wiring
	queries := database.New(db)
	repo := repository.NewSQLUserRepository(queries)
	svc := service.NewUserManager(repo)
	h := handler.NewUserHandler(svc)

	// 4. Test Request
	reqBody := handler.CreateUserRequest{
		Email:    "integration@example.com",
		Name:     "Integration User",
		Password: "password123",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/users", bytes.NewBuffer(body))
	rr := httptest.NewRecorder()

	// 5. Execute
	r := chi.NewRouter()
	r.Post("/users", h.CreateUser)
	r.ServeHTTP(rr, req)

	// 6. Assertions
	assert.Equal(t, http.StatusCreated, rr.Code, "Expected status 201")

	var user database.User
	err = json.NewDecoder(rr.Body).Decode(&user)
	assert.NoError(t, err, "Failed to decode response body")
	assert.Equal(t, "integration@example.com", user.Email, "Email mismatch in response")
	assert.NotEmpty(t, user.ID, "User ID should not be empty")

	// 7. Verify DB
	savedUser, err := repo.GetUserByEmail(context.Background(), "integration@example.com")
	assert.NoError(t, err, "Failed to get user from DB")
	assert.Equal(t, user.ID, savedUser.ID, "ID mismatch between response and DB")
}

func TestIntegration_SubjectFlow(t *testing.T) {
	// 1. Setup In-Memory DB
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	// 2. Apply Schema
	_, err = db.Exec(`
		PRAGMA foreign_keys = ON;
		CREATE TABLE users (
			id TEXT PRIMARY KEY, 
			email TEXT, 
			name TEXT, 
			password_hash TEXT NOT NULL DEFAULT '',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);
		CREATE TABLE subjects (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			color_hex TEXT,
			created_at TEXT NOT NULL DEFAULT (datetime('now')),
			updated_at TEXT NOT NULL DEFAULT (datetime('now')),
			deleted_at TEXT
		);
	`)
	if err != nil {
		t.Fatal(err)
	}

	// 3. Wiringgo
	queries := database.New(db)

	// User Setup
	userRepo := repository.NewSQLUserRepository(queries)
	userSvc := service.NewUserManager(userRepo)
	// Subject Setup
	subjectRepo := repository.NewSQLSubjectRepository(queries)
	subjectSvc := service.NewSubjectManager(subjectRepo)
	subjectHandler := handler.NewSubjectHandler(subjectSvc)

	ctx := context.Background()
	_, err = userSvc.CreateUser(ctx, "test@example.com", "Tester", "password123")
	assert.NoError(t, err)

	// 4. Test Request (Create Subject)
	reqBody := handler.CreateSubjectRequest{
		Name:     "Integration Math",
		ColorHex: "#FF0000",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/subjects", bytes.NewBuffer(body))
	// Add Basic Auth Header
	req.SetBasicAuth("test@example.com", "password123")
	rr := httptest.NewRecorder()

	// 5. Execute
	r := chi.NewRouter()
	// We need to import middleware package to use the real one,
	// or we can just test the handler logic if we trust the middleware unit tests.
	// Let's just test the handler + DB flow here.
	r.Post("/subjects", subjectHandler.CreateSubject)
	r.ServeHTTP(rr, req)

	// 6. Assertions
	assert.Equal(t, http.StatusCreated, rr.Code, "Expected status 201")

	var subject database.Subject
	err = json.NewDecoder(rr.Body).Decode(&subject)
	assert.NoError(t, err)
	assert.Equal(t, "Integration Math", subject.Name)

	// 7. Verify DB
	savedSubject, err := subjectRepo.GetSubject(ctx, subject.ID)
	assert.NoError(t, err)
	assert.Equal(t, subject.Name, savedSubject.Name)
}

func TestIntegration_TopicFlow(t *testing.T) {
	// 1. Setup DB
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	// Apply Schema (Users, Subjects, Topics)
	_, err = db.Exec(`
		PRAGMA foreign_keys = ON;
		CREATE TABLE users (id TEXT PRIMARY KEY, email TEXT, name TEXT, password_hash TEXT NOT NULL DEFAULT '', created_at DATETIME DEFAULT CURRENT_TIMESTAMP);
		CREATE TABLE subjects (id TEXT PRIMARY KEY, name TEXT NOT NULL, color_hex TEXT, created_at TEXT NOT NULL DEFAULT (datetime('now')), updated_at TEXT NOT NULL DEFAULT (datetime('now')), deleted_at TEXT);
		CREATE TABLE topics (id TEXT PRIMARY KEY, subject_id TEXT NOT NULL, name TEXT NOT NULL, created_at TEXT NOT NULL DEFAULT (datetime('now')), updated_at TEXT NOT NULL DEFAULT (datetime('now')), deleted_at TEXT, FOREIGN KEY (subject_id) REFERENCES subjects(id) ON DELETE CASCADE);
	`)
	if err != nil {
		t.Fatal(err)
	}

	queries := database.New(db)

	// Services
	userSvc := service.NewUserManager(repository.NewSQLUserRepository(queries))
	subjectSvc := service.NewSubjectManager(repository.NewSQLSubjectRepository(queries))
	topicSvc := service.NewTopicManager(repository.NewSQLTopicRepository(queries))
	topicHandler := handler.NewTopicHandler(topicSvc)

	ctx := context.Background()
	_, _ = userSvc.CreateUser(ctx, "test@example.com", "Tester", "pass")
	subject, _ := subjectSvc.CreateSubject(ctx, "Math", "#000")

	// Test Create Topic
	reqBody := handler.CreateTopicRequest{
		Name: "Algebra",
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/subjects/"+subject.ID+"/topics", bytes.NewBuffer(body))
	req.SetBasicAuth("test@example.com", "pass")
	rr := httptest.NewRecorder()

	r := chi.NewRouter()
	r.Post("/subjects/{id}/topics", topicHandler.CreateTopic)

	// Chi URL param mocking
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", subject.ID)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code)

	var topic database.Topic
	json.NewDecoder(rr.Body).Decode(&topic)
	assert.Equal(t, "Algebra", topic.Name)
	assert.Equal(t, subject.ID, topic.SubjectID)
}

func TestIntegration_StudyCycleFlow(t *testing.T) {
	// 1. Setup DB
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	// Apply Schema (Users, StudyCycles)
	_, err = db.Exec(`
		PRAGMA foreign_keys = ON;
		CREATE TABLE users (id TEXT PRIMARY KEY, email TEXT, name TEXT, password_hash TEXT NOT NULL DEFAULT '', created_at DATETIME DEFAULT CURRENT_TIMESTAMP);
		CREATE TABLE study_cycles (id TEXT PRIMARY KEY, name TEXT NOT NULL, description TEXT, is_active INTEGER DEFAULT 0, created_at TEXT NOT NULL DEFAULT (datetime('now')), updated_at TEXT NOT NULL DEFAULT (datetime('now')), deleted_at TEXT);
	`)
	if err != nil {
		t.Fatal(err)
	}

	queries := database.New(db)

	// Services
	userSvc := service.NewUserManager(repository.NewSQLUserRepository(queries))
	cycleSvc := service.NewStudyCycleManager(repository.NewSQLStudyCycleRepository(queries))
	cycleHandler := handler.NewStudyCycleHandler(cycleSvc)

	ctx := context.Background()
	_, _ = userSvc.CreateUser(ctx, "test@example.com", "Tester", "pass")

	// Test Create Cycle
	reqBody := handler.CreateStudyCycleRequest{
		Name:     "Exam Prep",
		IsActive: true,
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/study-cycles", bytes.NewBuffer(body))
	req.SetBasicAuth("test@example.com", "pass")
	rr := httptest.NewRecorder()

	r := chi.NewRouter()
	r.Post("/study-cycles", cycleHandler.CreateStudyCycle)
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code)

	var cycle database.StudyCycle
	json.NewDecoder(rr.Body).Decode(&cycle)
	assert.Equal(t, "Exam Prep", cycle.Name)
	assert.Equal(t, int64(1), cycle.IsActive.Int64)
}

func TestIntegration_CycleItemFlow(t *testing.T) {
	// 1. Setup DB
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	// Apply Schema
	_, err = db.Exec(`
		PRAGMA foreign_keys = ON;
		CREATE TABLE users (id TEXT PRIMARY KEY, email TEXT, name TEXT, password_hash TEXT NOT NULL DEFAULT '', created_at DATETIME DEFAULT CURRENT_TIMESTAMP);
		CREATE TABLE subjects (id TEXT PRIMARY KEY, name TEXT NOT NULL, color_hex TEXT, created_at TEXT NOT NULL DEFAULT (datetime('now')), updated_at TEXT NOT NULL DEFAULT (datetime('now')), deleted_at TEXT);
		CREATE TABLE study_cycles (id TEXT PRIMARY KEY, name TEXT NOT NULL, description TEXT, is_active INTEGER DEFAULT 0, created_at TEXT NOT NULL DEFAULT (datetime('now')), updated_at TEXT NOT NULL DEFAULT (datetime('now')), deleted_at TEXT);
		CREATE TABLE cycle_items (id TEXT PRIMARY KEY, cycle_id TEXT NOT NULL, subject_id TEXT NOT NULL, order_index INTEGER NOT NULL, planned_duration_minutes INTEGER DEFAULT 60, created_at TEXT NOT NULL DEFAULT (datetime('now')), updated_at TEXT NOT NULL DEFAULT (datetime('now')), FOREIGN KEY (cycle_id) REFERENCES study_cycles(id) ON DELETE CASCADE, FOREIGN KEY (subject_id) REFERENCES subjects(id) ON DELETE CASCADE);
	`)
	if err != nil {
		t.Fatal(err)
	}

	queries := database.New(db)

	// Services
	userSvc := service.NewUserManager(repository.NewSQLUserRepository(queries))
	subjectSvc := service.NewSubjectManager(repository.NewSQLSubjectRepository(queries))
	cycleSvc := service.NewStudyCycleManager(repository.NewSQLStudyCycleRepository(queries))
	itemSvc := service.NewCycleItemManager(repository.NewSQLCycleItemRepository(queries))
	itemHandler := handler.NewCycleItemHandler(itemSvc)

	ctx := context.Background()
	_, err = userSvc.CreateUser(ctx, "test@example.com", "Tester", "pass")
	assert.NoError(t, err)
	subject, err := subjectSvc.CreateSubject(ctx, "Math", "#000")
	assert.NoError(t, err)
	cycle, err := cycleSvc.CreateStudyCycle(ctx, "Cycle 1", "", true)
	assert.NoError(t, err)

	// Test Create Cycle Item
	reqBody := handler.CreateCycleItemRequest{
		SubjectID:              subject.ID,
		OrderIndex:             1,
		PlannedDurationMinutes: 60,
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/study-cycles/"+cycle.ID+"/items", bytes.NewBuffer(body))
	req.SetBasicAuth("test@example.com", "pass")
	rr := httptest.NewRecorder()

	r := chi.NewRouter()
	r.Post("/study-cycles/{id}/items", itemHandler.CreateCycleItem)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", cycle.ID)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code)

	var item database.CycleItem
	err = json.NewDecoder(rr.Body).Decode(&item)
	assert.NoError(t, err, "Failed to decode response body")
	assert.Equal(t, int64(1), item.OrderIndex)
	assert.Equal(t, subject.ID, item.SubjectID)
}

func TestIntegration_StudySessionFlow(t *testing.T) {
	// 1. Setup DB
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	// Apply Schema
	_, err = db.Exec(`
		PRAGMA foreign_keys = ON;
		CREATE TABLE users (id TEXT PRIMARY KEY, email TEXT, name TEXT, password_hash TEXT NOT NULL DEFAULT '', created_at DATETIME DEFAULT CURRENT_TIMESTAMP);
		CREATE TABLE subjects (id TEXT PRIMARY KEY, name TEXT NOT NULL, color_hex TEXT, created_at TEXT NOT NULL DEFAULT (datetime('now')), updated_at TEXT NOT NULL DEFAULT (datetime('now')), deleted_at TEXT);
		CREATE TABLE study_cycles (id TEXT PRIMARY KEY, name TEXT NOT NULL, description TEXT, is_active INTEGER DEFAULT 0, created_at TEXT NOT NULL DEFAULT (datetime('now')), updated_at TEXT NOT NULL DEFAULT (datetime('now')), deleted_at TEXT);
		CREATE TABLE cycle_items (id TEXT PRIMARY KEY, cycle_id TEXT NOT NULL, subject_id TEXT NOT NULL, order_index INTEGER NOT NULL, planned_duration_minutes INTEGER DEFAULT 60, created_at TEXT NOT NULL DEFAULT (datetime('now')), updated_at TEXT NOT NULL DEFAULT (datetime('now')), FOREIGN KEY (cycle_id) REFERENCES study_cycles(id) ON DELETE CASCADE, FOREIGN KEY (subject_id) REFERENCES subjects(id) ON DELETE CASCADE);
		CREATE TABLE study_sessions (id TEXT PRIMARY KEY, subject_id TEXT NOT NULL, cycle_item_id TEXT, started_at TEXT NOT NULL, finished_at TEXT, gross_duration_seconds INTEGER DEFAULT 0, net_duration_seconds INTEGER DEFAULT 0, notes TEXT, created_at TEXT NOT NULL DEFAULT (datetime('now')), updated_at TEXT NOT NULL DEFAULT (datetime('now')), FOREIGN KEY (subject_id) REFERENCES subjects(id), FOREIGN KEY (cycle_item_id) REFERENCES cycle_items(id));
	`)
	if err != nil {
		t.Fatal(err)
	}

	queries := database.New(db)

	// Services
	userSvc := service.NewUserManager(repository.NewSQLUserRepository(queries))
	subjectSvc := service.NewSubjectManager(repository.NewSQLSubjectRepository(queries))
	sessionSvc := service.NewStudySessionManager(repository.NewSQLStudySessionRepository(queries))
	sessionHandler := handler.NewStudySessionHandler(sessionSvc)

	ctx := context.Background()
	_, _ = userSvc.CreateUser(ctx, "test@example.com", "Tester", "pass")
	subject, _ := subjectSvc.CreateSubject(ctx, "Math", "#000")

	// Test Start Session
	reqBody := handler.CreateStudySessionRequest{
		SubjectID: subject.ID,
		StartedAt: "2023-10-27T10:00:00Z",
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/study-sessions", bytes.NewBuffer(body))
	req.SetBasicAuth("test@example.com", "pass")
	rr := httptest.NewRecorder()

	r := chi.NewRouter()
	r.Post("/study-sessions", sessionHandler.CreateStudySession)
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code)

	var session database.StudySession
	json.NewDecoder(rr.Body).Decode(&session)
	assert.Equal(t, subject.ID, session.SubjectID)
	assert.NotEmpty(t, session.StartedAt)
}
