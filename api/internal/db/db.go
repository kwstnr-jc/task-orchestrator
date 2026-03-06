package db

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Store struct {
	pool *pgxpool.Pool
}

func NewStore(pool *pgxpool.Pool) *Store {
	return &Store{pool: pool}
}

func Connect(ctx context.Context, databaseURL string) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("parse db config: %w", err)
	}
	cfg.MaxConns = 10
	cfg.MinConns = 2
	cfg.MaxConnLifetime = 30 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("create pool: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping db: %w", err)
	}
	return pool, nil
}

type ListTasksParams struct {
	TaskType *TaskType
	State    *TaskState
}

func (s *Store) ListTasks(ctx context.Context, params ListTasksParams) ([]Task, error) {
	query := `SELECT id, title, description, task_type, state, priority, metadata, created_by, created_at, updated_at
		FROM tasks WHERE 1=1`
	args := []any{}
	argIdx := 1

	if params.TaskType != nil {
		query += fmt.Sprintf(" AND task_type = $%d", argIdx)
		args = append(args, *params.TaskType)
		argIdx++
	}
	if params.State != nil {
		query += fmt.Sprintf(" AND state = $%d", argIdx)
		args = append(args, *params.State)
	}
	query += " ORDER BY priority DESC, created_at DESC"

	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list tasks: %w", err)
	}
	defer rows.Close()

	tasks := []Task{}
	for rows.Next() {
		var t Task
		if err := rows.Scan(&t.ID, &t.Title, &t.Description, &t.TaskType, &t.State,
			&t.Priority, &t.Metadata, &t.CreatedBy, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan task: %w", err)
		}
		tasks = append(tasks, t)
	}
	return tasks, rows.Err()
}

func (s *Store) GetTask(ctx context.Context, id uuid.UUID) (Task, error) {
	var t Task
	err := s.pool.QueryRow(ctx,
		`SELECT id, title, description, task_type, state, priority, metadata, created_by, created_at, updated_at
		 FROM tasks WHERE id = $1`, id).
		Scan(&t.ID, &t.Title, &t.Description, &t.TaskType, &t.State,
			&t.Priority, &t.Metadata, &t.CreatedBy, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return t, ErrNotFound
		}
		return t, fmt.Errorf("get task: %w", err)
	}
	return t, nil
}

var ErrNotFound = fmt.Errorf("not found")

type CreateTaskParams struct {
	Title       string          `json:"title"`
	Description *string         `json:"description"`
	TaskType    TaskType        `json:"task_type"`
	Priority    int32           `json:"priority"`
	Metadata    json.RawMessage `json:"metadata"`
	CreatedBy   string          `json:"-"`
}

func (s *Store) CreateTask(ctx context.Context, p CreateTaskParams) (Task, error) {
	state := DefaultState(p.TaskType)
	if p.Metadata == nil {
		p.Metadata = json.RawMessage(`{}`)
	}

	var t Task
	err := s.pool.QueryRow(ctx,
		`INSERT INTO tasks (title, description, task_type, state, priority, metadata, created_by)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)
		 RETURNING id, title, description, task_type, state, priority, metadata, created_by, created_at, updated_at`,
		p.Title, p.Description, p.TaskType, state, p.Priority, p.Metadata, p.CreatedBy).
		Scan(&t.ID, &t.Title, &t.Description, &t.TaskType, &t.State,
			&t.Priority, &t.Metadata, &t.CreatedBy, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		return t, fmt.Errorf("create task: %w", err)
	}
	return t, nil
}

type UpdateTaskParams struct {
	Title       *string         `json:"title"`
	Description *string         `json:"description"`
	State       *TaskState      `json:"state"`
	Priority    *int32          `json:"priority"`
	Metadata    json.RawMessage `json:"metadata"`
}

func (s *Store) UpdateTask(ctx context.Context, id uuid.UUID, p UpdateTaskParams) (Task, error) {
	query := `UPDATE tasks SET updated_at = NOW()`
	args := []any{}
	argIdx := 1

	if p.Title != nil {
		query += fmt.Sprintf(", title = $%d", argIdx)
		args = append(args, *p.Title)
		argIdx++
	}
	if p.Description != nil {
		query += fmt.Sprintf(", description = $%d", argIdx)
		args = append(args, *p.Description)
		argIdx++
	}
	if p.State != nil {
		query += fmt.Sprintf(", state = $%d", argIdx)
		args = append(args, *p.State)
		argIdx++
	}
	if p.Priority != nil {
		query += fmt.Sprintf(", priority = $%d", argIdx)
		args = append(args, *p.Priority)
		argIdx++
	}
	if p.Metadata != nil {
		query += fmt.Sprintf(", metadata = $%d", argIdx)
		args = append(args, p.Metadata)
		argIdx++
	}

	query += fmt.Sprintf(" WHERE id = $%d", argIdx)
	args = append(args, id)

	query += ` RETURNING id, title, description, task_type, state, priority, metadata, created_by, created_at, updated_at`

	var t Task
	err := s.pool.QueryRow(ctx, query, args...).
		Scan(&t.ID, &t.Title, &t.Description, &t.TaskType, &t.State,
			&t.Priority, &t.Metadata, &t.CreatedBy, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return t, ErrNotFound
		}
		return t, fmt.Errorf("update task: %w", err)
	}
	return t, nil
}

func (s *Store) DeleteTask(ctx context.Context, id uuid.UUID) error {
	tag, err := s.pool.Exec(ctx, `DELETE FROM tasks WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete task: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *Store) UpsertUser(ctx context.Context, username string, displayName *string) (User, error) {
	var u User
	err := s.pool.QueryRow(ctx,
		`INSERT INTO users (github_username, display_name)
		 VALUES ($1, $2)
		 ON CONFLICT (github_username) DO UPDATE SET display_name = EXCLUDED.display_name
		 RETURNING github_username, display_name, role, created_at`,
		username, displayName).
		Scan(&u.GithubUsername, &u.DisplayName, &u.Role, &u.CreatedAt)
	if err != nil {
		return u, fmt.Errorf("upsert user: %w", err)
	}
	return u, nil
}
