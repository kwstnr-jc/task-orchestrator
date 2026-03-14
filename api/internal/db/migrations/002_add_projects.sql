-- Create projects table
CREATE TABLE projects (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL UNIQUE,
    description TEXT,
    color TEXT DEFAULT '#3b82f6',
    created_by TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    FOREIGN KEY (created_by) REFERENCES users(github_username)
);

-- Add project_id to tasks
ALTER TABLE tasks ADD COLUMN project_id UUID REFERENCES projects(id) ON DELETE CASCADE;

-- Add indexes
CREATE INDEX idx_tasks_project_id ON tasks(project_id);
CREATE INDEX idx_projects_created_by ON projects(created_by);
