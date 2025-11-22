package repository

import (
	"context"

	"github.com/joaoapaenas/my-api/internal/database"
)

type StudyRepository interface {
	// Subjects
	CreateSubject(ctx context.Context, arg database.CreateSubjectParams) (database.Subject, error)
	ListSubjects(ctx context.Context) ([]database.Subject, error)

	// Topics
	CreateTopic(ctx context.Context, arg database.CreateTopicParams) (database.Topic, error)
	ListTopicsBySubject(ctx context.Context, subjectID string) ([]database.Topic, error)

	// Study Cycles
	CreateStudyCycle(ctx context.Context, arg database.CreateStudyCycleParams) (database.StudyCycle, error)
	GetActiveStudyCycle(ctx context.Context) (database.StudyCycle, error)

	// Cycle Items
	CreateCycleItem(ctx context.Context, arg database.CreateCycleItemParams) (database.CycleItem, error)
	ListCycleItems(ctx context.Context, cycleID string) ([]database.CycleItem, error)

	// Sessions
	CreateStudySession(ctx context.Context, arg database.CreateStudySessionParams) (database.StudySession, error)
	UpdateSessionDuration(ctx context.Context, arg database.UpdateSessionDurationParams) error

	// Pauses
	CreateSessionPause(ctx context.Context, arg database.CreateSessionPauseParams) (database.SessionPause, error)
	EndSessionPause(ctx context.Context, arg database.EndSessionPauseParams) error

	// Exercises
	CreateExerciseLog(ctx context.Context, arg database.CreateExerciseLogParams) (database.ExerciseLog, error)
}

type SQLStudyRepository struct {
	q database.Querier
}

func NewSQLStudyRepository(q database.Querier) *SQLStudyRepository {
	return &SQLStudyRepository{q: q}
}

func (r *SQLStudyRepository) CreateSubject(ctx context.Context, arg database.CreateSubjectParams) (database.Subject, error) {
	return r.q.CreateSubject(ctx, arg)
}

func (r *SQLStudyRepository) ListSubjects(ctx context.Context) ([]database.Subject, error) {
	return r.q.ListSubjects(ctx)
}

func (r *SQLStudyRepository) CreateTopic(ctx context.Context, arg database.CreateTopicParams) (database.Topic, error) {
	return r.q.CreateTopic(ctx, arg)
}

func (r *SQLStudyRepository) ListTopicsBySubject(ctx context.Context, subjectID string) ([]database.Topic, error) {
	return r.q.ListTopicsBySubject(ctx, subjectID)
}

func (r *SQLStudyRepository) CreateStudyCycle(ctx context.Context, arg database.CreateStudyCycleParams) (database.StudyCycle, error) {
	return r.q.CreateStudyCycle(ctx, arg)
}

func (r *SQLStudyRepository) GetActiveStudyCycle(ctx context.Context) (database.StudyCycle, error) {
	return r.q.GetActiveStudyCycle(ctx)
}

func (r *SQLStudyRepository) CreateCycleItem(ctx context.Context, arg database.CreateCycleItemParams) (database.CycleItem, error) {
	return r.q.CreateCycleItem(ctx, arg)
}

func (r *SQLStudyRepository) ListCycleItems(ctx context.Context, cycleID string) ([]database.CycleItem, error) {
	return r.q.ListCycleItems(ctx, cycleID)
}

func (r *SQLStudyRepository) CreateStudySession(ctx context.Context, arg database.CreateStudySessionParams) (database.StudySession, error) {
	return r.q.CreateStudySession(ctx, arg)
}

func (r *SQLStudyRepository) UpdateSessionDuration(ctx context.Context, arg database.UpdateSessionDurationParams) error {
	return r.q.UpdateSessionDuration(ctx, arg)
}

func (r *SQLStudyRepository) CreateSessionPause(ctx context.Context, arg database.CreateSessionPauseParams) (database.SessionPause, error) {
	return r.q.CreateSessionPause(ctx, arg)
}

func (r *SQLStudyRepository) EndSessionPause(ctx context.Context, arg database.EndSessionPauseParams) error {
	return r.q.EndSessionPause(ctx, arg)
}

func (r *SQLStudyRepository) CreateExerciseLog(ctx context.Context, arg database.CreateExerciseLogParams) (database.ExerciseLog, error) {
	return r.q.CreateExerciseLog(ctx, arg)
}
