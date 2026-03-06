package db

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type TaskType string

const (
	TaskTypeDev      TaskType = "dev"
	TaskTypeResearch TaskType = "research"
)

type TaskState string

const (
	TaskStateDraft      TaskState = "draft"
	TaskStateRefine     TaskState = "refine"
	TaskStateApproved   TaskState = "approved"
	TaskStateInProgress TaskState = "in_progress"
	TaskStateDone       TaskState = "done"
)

// ValidTransitions defines allowed state transitions per task type.
var ValidTransitions = map[TaskType]map[TaskState][]TaskState{
	TaskTypeDev: {
		TaskStateDraft:      {TaskStateRefine},
		TaskStateRefine:     {TaskStateApproved, TaskStateDraft},
		TaskStateApproved:   {TaskStateInProgress},
		TaskStateInProgress: {TaskStateDone, TaskStateApproved},
		TaskStateDone:       {},
	},
	TaskTypeResearch: {
		TaskStateDraft:      {TaskStateInProgress},
		TaskStateInProgress: {TaskStateDone, TaskStateDraft},
		TaskStateDone:       {},
	},
}

func IsValidTransition(taskType TaskType, from, to TaskState) bool {
	targets, ok := ValidTransitions[taskType][from]
	if !ok {
		return false
	}
	for _, t := range targets {
		if t == to {
			return true
		}
	}
	return false
}

func DefaultState(tt TaskType) TaskState {
	if tt == TaskTypeResearch {
		return TaskStateDraft
	}
	return TaskStateDraft
}

type Task struct {
	ID          uuid.UUID       `json:"id"`
	Title       string          `json:"title"`
	Description *string         `json:"description"`
	TaskType    TaskType        `json:"task_type"`
	State       TaskState       `json:"state"`
	Priority    int32           `json:"priority"`
	Metadata    json.RawMessage `json:"metadata"`
	ProjectID   *uuid.UUID      `json:"project_id"`
	CreatedBy   string          `json:"created_by"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

type Project struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description *string   `json:"description"`
	Color       string    `json:"color"`
	CreatedBy   string    `json:"created_by"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type User struct {
	GithubUsername string    `json:"github_username"`
	DisplayName   *string   `json:"display_name"`
	Role          string    `json:"role"`
	CreatedAt     time.Time `json:"created_at"`
}
