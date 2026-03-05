-- name: UpsertUser :one
INSERT INTO users (github_username, display_name)
VALUES ($1, $2)
ON CONFLICT (github_username) DO UPDATE SET display_name = EXCLUDED.display_name
RETURNING *;

-- name: GetUser :one
SELECT * FROM users WHERE github_username = $1;
