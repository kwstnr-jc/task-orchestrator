-- name: ListTasks :many
SELECT * FROM tasks
WHERE
    (sqlc.narg('task_type')::task_type IS NULL OR task_type = sqlc.narg('task_type')::task_type)
    AND (sqlc.narg('state')::task_state IS NULL OR state = sqlc.narg('state')::task_state)
ORDER BY priority DESC, created_at DESC;

-- name: GetTask :one
SELECT * FROM tasks WHERE id = $1;

-- name: CreateTask :one
INSERT INTO tasks (title, description, task_type, state, priority, metadata, created_by)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: UpdateTask :one
UPDATE tasks
SET
    title = COALESCE(sqlc.narg('title'), title),
    description = COALESCE(sqlc.narg('description'), description),
    state = COALESCE(sqlc.narg('state'), state),
    priority = COALESCE(sqlc.narg('priority'), priority),
    metadata = COALESCE(sqlc.narg('metadata'), metadata),
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteTask :exec
DELETE FROM tasks WHERE id = $1;
