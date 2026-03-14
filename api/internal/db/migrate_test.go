package db

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

func setupTestDB(t *testing.T) *pgxpool.Pool {
	t.Helper()
	ctx := context.Background()

	pgContainer, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("test"),
		postgres.WithPassword("test"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second)),
	)
	if err != nil {
		t.Fatalf("failed to start postgres container: %v", err)
	}
	t.Cleanup(func() { _ = pgContainer.Terminate(ctx) })

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("failed to get connection string: %v", err)
	}

	pool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		t.Fatalf("failed to create pool: %v", err)
	}
	t.Cleanup(func() { pool.Close() })

	return pool
}

func TestMigrate_AppliesAll(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	pool := setupTestDB(t)
	ctx := context.Background()

	// Run migrations
	if err := Migrate(ctx, pool); err != nil {
		t.Fatalf("Migrate failed: %v", err)
	}

	// Verify schema_migrations has entries
	var count int
	err := pool.QueryRow(ctx, `SELECT COUNT(*) FROM schema_migrations`).Scan(&count)
	if err != nil {
		t.Fatalf("failed to query schema_migrations: %v", err)
	}
	if count != 2 {
		t.Fatalf("expected 2 migrations applied, got %d", count)
	}

	// Verify tasks table exists with expected columns
	var exists bool
	err = pool.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM information_schema.columns
			WHERE table_name = 'tasks' AND column_name = 'project_id'
		)
	`).Scan(&exists)
	if err != nil {
		t.Fatalf("failed to check tasks.project_id: %v", err)
	}
	if !exists {
		t.Fatal("expected tasks.project_id column to exist after migrations")
	}

	// Verify projects table exists
	err = pool.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM information_schema.tables
			WHERE table_name = 'projects'
		)
	`).Scan(&exists)
	if err != nil {
		t.Fatalf("failed to check projects table: %v", err)
	}
	if !exists {
		t.Fatal("expected projects table to exist after migrations")
	}

	// Verify users table exists
	err = pool.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM information_schema.tables
			WHERE table_name = 'users'
		)
	`).Scan(&exists)
	if err != nil {
		t.Fatalf("failed to check users table: %v", err)
	}
	if !exists {
		t.Fatal("expected users table to exist after migrations")
	}
}

func TestMigrate_Idempotent(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	pool := setupTestDB(t)
	ctx := context.Background()

	// Run twice
	if err := Migrate(ctx, pool); err != nil {
		t.Fatalf("first Migrate failed: %v", err)
	}
	if err := Migrate(ctx, pool); err != nil {
		t.Fatalf("second Migrate failed: %v", err)
	}

	// Still only 2 entries
	var count int
	err := pool.QueryRow(ctx, `SELECT COUNT(*) FROM schema_migrations`).Scan(&count)
	if err != nil {
		t.Fatalf("failed to query schema_migrations: %v", err)
	}
	if count != 2 {
		t.Fatalf("expected 2 migrations, got %d", count)
	}
}

func TestMigrate_StoreOperations(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	pool := setupTestDB(t)
	ctx := context.Background()

	if err := Migrate(ctx, pool); err != nil {
		t.Fatalf("Migrate failed: %v", err)
	}

	store := NewStore(pool)

	// Create a user (needed for FK on projects)
	displayName := "Test User"
	user, err := store.UpsertUser(ctx, "testuser", &displayName)
	if err != nil {
		t.Fatalf("UpsertUser failed: %v", err)
	}
	if user.GithubUsername != "testuser" {
		t.Fatalf("expected username 'testuser', got %s", user.GithubUsername)
	}

	// Create a project
	project, err := store.CreateProject(ctx, CreateProjectParams{
		Name:      "Test Project",
		CreatedBy: "testuser",
	})
	if err != nil {
		t.Fatalf("CreateProject failed: %v", err)
	}
	if project.Name != "Test Project" {
		t.Fatalf("expected project name 'Test Project', got %s", project.Name)
	}

	// Create a task in the project
	task, err := store.CreateTask(ctx, CreateTaskParams{
		Title:     "Test Task",
		TaskType:  TaskTypeDev,
		ProjectID: &project.ID,
		CreatedBy: "testuser",
	})
	if err != nil {
		t.Fatalf("CreateTask failed: %v", err)
	}
	if task.Title != "Test Task" {
		t.Fatalf("expected task title 'Test Task', got %s", task.Title)
	}
	if task.State != TaskStateDraft {
		t.Fatalf("expected state 'draft', got %s", task.State)
	}
	if task.ProjectID == nil || *task.ProjectID != project.ID {
		t.Fatal("expected task to be linked to project")
	}

	// List tasks
	tasks, err := store.ListTasks(ctx, ListTasksParams{})
	if err != nil {
		t.Fatalf("ListTasks failed: %v", err)
	}
	if len(tasks) != 1 {
		t.Fatalf("expected 1 task, got %d", len(tasks))
	}

	// Update task state
	refine := TaskStateRefine
	updated, err := store.UpdateTask(ctx, task.ID, UpdateTaskParams{State: &refine})
	if err != nil {
		t.Fatalf("UpdateTask failed: %v", err)
	}
	if updated.State != TaskStateRefine {
		t.Fatalf("expected state 'refine', got %s", updated.State)
	}

	// Delete task
	if err := store.DeleteTask(ctx, task.ID); err != nil {
		t.Fatalf("DeleteTask failed: %v", err)
	}

	// Delete project
	if err := store.DeleteProject(ctx, project.ID); err != nil {
		t.Fatalf("DeleteProject failed: %v", err)
	}
}
