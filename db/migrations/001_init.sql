CREATE TYPE task_type AS ENUM ('dev', 'research');
CREATE TYPE task_state AS ENUM ('draft', 'refine', 'approved', 'in_progress', 'done');

CREATE TABLE tasks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title TEXT NOT NULL,
    description TEXT,
    task_type task_type NOT NULL DEFAULT 'dev',
    state task_state NOT NULL DEFAULT 'draft',
    priority INTEGER DEFAULT 0,
    metadata JSONB DEFAULT '{}',
    created_by TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE users (
    github_username TEXT PRIMARY KEY,
    display_name TEXT,
    role TEXT DEFAULT 'admin',
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_tasks_state ON tasks(state);
CREATE INDEX idx_tasks_type ON tasks(task_type);
CREATE INDEX idx_tasks_created_by ON tasks(created_by);
