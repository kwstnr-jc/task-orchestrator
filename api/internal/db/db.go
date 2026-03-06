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

var ErrNotFound = fmt.Errorf("not found")

// ── Tasks ────────────────────────────────────────────

const taskColumns = `id, title, description, task_type, state, priority, metadata, project_id, created_by, created_at, updated_at`

func scanTask(row pgx.Row) (Task, error) {
	var t Task
	err := row.Scan(&t.ID, &t.Title, &t.Description, &t.TaskType, &t.State,
		&t.Priority, &t.Metadata, &t.ProjectID, &t.CreatedBy, &t.CreatedAt, &t.UpdatedAt)
	return t, err
}

type ListTasksParams struct {
	TaskType  *TaskType
	State     *TaskState
	ProjectID *uuid.UUID
}

func (s *Store) ListTasks(ctx context.Context, params ListTasksParams) ([]Task, error) {
	query := `SELECT ` + taskColumns + ` FROM tasks WHERE 1=1`
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
		argIdx++
	}
	if params.ProjectID != nil {
		query += fmt.Sprintf(" AND project_id = $%d", argIdx)
		args = append(args, *params.ProjectID)
		argIdx++
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
			&t.Priority, &t.Metadata, &t.ProjectID, &t.CreatedBy, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan task: %w", err)
		}
		tasks = append(tasks, t)
	}
	return tasks, rows.Err()
}

func (s *Store) GetTask(ctx context.Context, id uuid.UUID) (Task, error) {
	t, err := scanTask(s.pool.QueryRow(ctx,
		`SELECT `+taskColumns+` FROM tasks WHERE id = $1`, id))
	if err != nil {
		if err == pgx.ErrNoRows {
			return t, ErrNotFound
		}
		return t, fmt.Errorf("get task: %w", err)
	}
	return t, nil
}

type CreateTaskParams struct {
	Title       string          `json:"title"`
	Description *string         `json:"description"`
	TaskType    TaskType        `json:"task_type"`
	Priority    int32           `json:"priority"`
	Metadata    json.RawMessage `json:"metadata"`
	ProjectID   *uuid.UUID      `json:"project_id"`
	CreatedBy   string          `json:"-"`
}

func (s *Store) CreateTask(ctx context.Context, p CreateTaskParams) (Task, error) {
	state := DefaultState(p.TaskType)
	if p.Metadata == nil {
		p.Metadata = json.RawMessage(`{}`)
	}

	t, err := scanTask(s.pool.QueryRow(ctx,
		`INSERT INTO tasks (title, description, task_type, state, priority, metadata, project_id, created_by)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		 RETURNING `+taskColumns,
		p.Title, p.Description, p.TaskType, state, p.Priority, p.Metadata, p.ProjectID, p.CreatedBy))
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
	ProjectID   *uuid.UUID      `json:"project_id"`
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
	if p.ProjectID != nil {
		query += fmt.Sprintf(", project_id = $%d", argIdx)
		args = append(args, *p.ProjectID)
		argIdx++
	}

	query += fmt.Sprintf(" WHERE id = $%d", argIdx)
	args = append(args, id)
	query += ` RETURNING ` + taskColumns

	t, err := scanTask(s.pool.QueryRow(ctx, query, args...))
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

// ── Users ────────────────────────────────────────────

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

// ── Projects ─────────────────────────────────────────

func (s *Store) ListProjects(ctx context.Context) ([]Project, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT id, name, description, color, created_by, created_at, updated_at
		 FROM projects ORDER BY name ASC`)
	if err != nil {
		return nil, fmt.Errorf("list projects: %w", err)
	}
	defer rows.Close()

	projects := []Project{}
	for rows.Next() {
		var p Project
		if err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.Color, &p.CreatedBy, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan project: %w", err)
		}
		projects = append(projects, p)
	}
	return projects, rows.Err()
}

func (s *Store) GetProject(ctx context.Context, id uuid.UUID) (Project, error) {
	var p Project
	err := s.pool.QueryRow(ctx,
		`SELECT id, name, description, color, created_by, created_at, updated_at
		 FROM projects WHERE id = $1`, id).
		Scan(&p.ID, &p.Name, &p.Description, &p.Color, &p.CreatedBy, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return p, ErrNotFound
		}
		return p, fmt.Errorf("get project: %w", err)
	}
	return p, nil
}

type CreateProjectParams struct {
	Name        string  `json:"name"`
	Description *string `json:"description"`
	Color       string  `json:"color"`
	CreatedBy   string  `json:"-"`
}

func (s *Store) CreateProject(ctx context.Context, p CreateProjectParams) (Project, error) {
	color := p.Color
	if color == "" {
		color = "#3b82f6"
	}

	var proj Project
	err := s.pool.QueryRow(ctx,
		`INSERT INTO projects (name, description, color, created_by)
		 VALUES ($1, $2, $3, $4)
		 RETURNING id, name, description, color, created_by, created_at, updated_at`,
		p.Name, p.Description, color, p.CreatedBy).
		Scan(&proj.ID, &proj.Name, &proj.Description, &proj.Color, &proj.CreatedBy, &proj.CreatedAt, &proj.UpdatedAt)
	if err != nil {
		return proj, fmt.Errorf("create project: %w", err)
	}
	return proj, nil
}

type UpdateProjectParams struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
	Color       *string `json:"color"`
}

func (s *Store) UpdateProject(ctx context.Context, id uuid.UUID, p UpdateProjectParams) (Project, error) {
	query := `UPDATE projects SET updated_at = NOW()`
	args := []any{}
	argIdx := 1

	if p.Name != nil {
		query += fmt.Sprintf(", name = $%d", argIdx)
		args = append(args, *p.Name)
		argIdx++
	}
	if p.Description != nil {
		query += fmt.Sprintf(", description = $%d", argIdx)
		args = append(args, *p.Description)
		argIdx++
	}
	if p.Color != nil {
		query += fmt.Sprintf(", color = $%d", argIdx)
		args = append(args, *p.Color)
		argIdx++
	}

	query += fmt.Sprintf(" WHERE id = $%d", argIdx)
	args = append(args, id)
	query += ` RETURNING id, name, description, color, created_by, created_at, updated_at`

	var proj Project
	err := s.pool.QueryRow(ctx, query, args...).
		Scan(&proj.ID, &proj.Name, &proj.Description, &proj.Color, &proj.CreatedBy, &proj.CreatedAt, &proj.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return proj, ErrNotFound
		}
		return proj, fmt.Errorf("update project: %w", err)
	}
	return proj, nil
}

func (s *Store) DeleteProject(ctx context.Context, id uuid.UUID) error {
	tag, err := s.pool.Exec(ctx, `DELETE FROM projects WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete project: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
