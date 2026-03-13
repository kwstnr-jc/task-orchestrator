package db

import (
	"context"

	"github.com/google/uuid"
)

// TaskStore defines the interface for task database operations.
type TaskStore interface {
	ListTasks(ctx context.Context, params ListTasksParams) ([]Task, error)
	GetTask(ctx context.Context, id uuid.UUID) (Task, error)
	CreateTask(ctx context.Context, p CreateTaskParams) (Task, error)
	UpdateTask(ctx context.Context, id uuid.UUID, p UpdateTaskParams) (Task, error)
	DeleteTask(ctx context.Context, id uuid.UUID) error
}

// ProjectStore defines the interface for project database operations.
type ProjectStore interface {
	ListProjects(ctx context.Context) ([]Project, error)
	GetProject(ctx context.Context, id uuid.UUID) (Project, error)
	CreateProject(ctx context.Context, p CreateProjectParams) (Project, error)
	UpdateProject(ctx context.Context, id uuid.UUID, p UpdateProjectParams) (Project, error)
	DeleteProject(ctx context.Context, id uuid.UUID) error
}

// UserStore defines the interface for user database operations.
type UserStore interface {
	UpsertUser(ctx context.Context, username string, displayName *string) (User, error)
}

// Verify Store implements all interfaces at compile time.
var (
	_ TaskStore    = (*Store)(nil)
	_ ProjectStore = (*Store)(nil)
	_ UserStore    = (*Store)(nil)
)
