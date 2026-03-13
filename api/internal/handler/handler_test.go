package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/kwstnr-jc/task-orchestrator/api/internal/auth"
	"github.com/kwstnr-jc/task-orchestrator/api/internal/db"
)

// ── Mock stores ─────────────────────────────────────

type mockTaskStore struct {
	tasks []db.Task
	err   error
}

func (m *mockTaskStore) ListTasks(_ context.Context, _ db.ListTasksParams) ([]db.Task, error) {
	return m.tasks, m.err
}
func (m *mockTaskStore) GetTask(_ context.Context, id uuid.UUID) (db.Task, error) {
	for _, t := range m.tasks {
		if t.ID == id {
			return t, nil
		}
	}
	if m.err != nil {
		return db.Task{}, m.err
	}
	return db.Task{}, db.ErrNotFound
}
func (m *mockTaskStore) CreateTask(_ context.Context, p db.CreateTaskParams) (db.Task, error) {
	if m.err != nil {
		return db.Task{}, m.err
	}
	return db.Task{
		ID:        uuid.New(),
		Title:     p.Title,
		TaskType:  p.TaskType,
		State:     db.DefaultState(p.TaskType),
		Priority:  p.Priority,
		Metadata:  json.RawMessage(`{}`),
		CreatedBy: p.CreatedBy,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}, nil
}
func (m *mockTaskStore) UpdateTask(_ context.Context, id uuid.UUID, p db.UpdateTaskParams) (db.Task, error) {
	for _, t := range m.tasks {
		if t.ID == id {
			if p.Title != nil {
				t.Title = *p.Title
			}
			if p.State != nil {
				t.State = *p.State
			}
			return t, nil
		}
	}
	if m.err != nil {
		return db.Task{}, m.err
	}
	return db.Task{}, db.ErrNotFound
}
func (m *mockTaskStore) DeleteTask(_ context.Context, id uuid.UUID) error {
	for _, t := range m.tasks {
		if t.ID == id {
			return nil
		}
	}
	if m.err != nil {
		return m.err
	}
	return db.ErrNotFound
}

type mockProjectStore struct {
	projects []db.Project
	err      error
}

func (m *mockProjectStore) ListProjects(_ context.Context) ([]db.Project, error) {
	return m.projects, m.err
}
func (m *mockProjectStore) GetProject(_ context.Context, id uuid.UUID) (db.Project, error) {
	for _, p := range m.projects {
		if p.ID == id {
			return p, nil
		}
	}
	return db.Project{}, db.ErrNotFound
}
func (m *mockProjectStore) CreateProject(_ context.Context, p db.CreateProjectParams) (db.Project, error) {
	if m.err != nil {
		return db.Project{}, m.err
	}
	return db.Project{
		ID:        uuid.New(),
		Name:      p.Name,
		Color:     "#3b82f6",
		CreatedBy: p.CreatedBy,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}, nil
}
func (m *mockProjectStore) UpdateProject(_ context.Context, id uuid.UUID, p db.UpdateProjectParams) (db.Project, error) {
	for _, proj := range m.projects {
		if proj.ID == id {
			if p.Name != nil {
				proj.Name = *p.Name
			}
			return proj, nil
		}
	}
	return db.Project{}, db.ErrNotFound
}
func (m *mockProjectStore) DeleteProject(_ context.Context, id uuid.UUID) error {
	for _, p := range m.projects {
		if p.ID == id {
			return nil
		}
	}
	return db.ErrNotFound
}

// ── Helpers ─────────────────────────────────────────

func withUser(r *http.Request) *http.Request {
	ctx := context.WithValue(r.Context(), auth.UserKey, auth.UserInfo{
		Username:    "testuser",
		DisplayName: "Test User",
		Source:      "dev",
	})
	return r.WithContext(ctx)
}

func withChiParam(r *http.Request, key, value string) *http.Request {
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add(key, value)
	return r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
}

func parseJSON(t *testing.T, body *bytes.Buffer) map[string]any {
	t.Helper()
	var result map[string]any
	if err := json.Unmarshal(body.Bytes(), &result); err != nil {
		t.Fatalf("failed to parse JSON response: %v", err)
	}
	return result
}

// ── Task Handler Tests ──────────────────────────────

func TestTaskList(t *testing.T) {
	store := &mockTaskStore{
		tasks: []db.Task{
			{ID: uuid.New(), Title: "Task 1", TaskType: db.TaskTypeDev, State: db.TaskStateDraft},
			{ID: uuid.New(), Title: "Task 2", TaskType: db.TaskTypeDev, State: db.TaskStateDraft},
		},
	}
	h := NewTaskHandler(store)

	req := httptest.NewRequest("GET", "/api/tasks", nil)
	w := httptest.NewRecorder()
	h.List(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var tasks []db.Task
	json.Unmarshal(w.Body.Bytes(), &tasks)
	if len(tasks) != 2 {
		t.Fatalf("expected 2 tasks, got %d", len(tasks))
	}
}

func TestTaskGet(t *testing.T) {
	id := uuid.New()
	store := &mockTaskStore{
		tasks: []db.Task{{ID: id, Title: "Test Task", TaskType: db.TaskTypeDev}},
	}
	h := NewTaskHandler(store)

	req := httptest.NewRequest("GET", "/api/tasks/"+id.String(), nil)
	req = withChiParam(req, "id", id.String())
	w := httptest.NewRecorder()
	h.Get(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	result := parseJSON(t, w.Body)
	if result["title"] != "Test Task" {
		t.Fatalf("expected 'Test Task', got %v", result["title"])
	}
}

func TestTaskGetNotFound(t *testing.T) {
	store := &mockTaskStore{}
	h := NewTaskHandler(store)

	id := uuid.New()
	req := httptest.NewRequest("GET", "/api/tasks/"+id.String(), nil)
	req = withChiParam(req, "id", id.String())
	w := httptest.NewRecorder()
	h.Get(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestTaskGetInvalidID(t *testing.T) {
	h := NewTaskHandler(&mockTaskStore{})

	req := httptest.NewRequest("GET", "/api/tasks/not-a-uuid", nil)
	req = withChiParam(req, "id", "not-a-uuid")
	w := httptest.NewRecorder()
	h.Get(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestTaskCreate(t *testing.T) {
	store := &mockTaskStore{}
	h := NewTaskHandler(store)

	body := `{"title":"New Task","task_type":"dev"}`
	req := httptest.NewRequest("POST", "/api/tasks", bytes.NewBufferString(body))
	req = withUser(req)
	w := httptest.NewRecorder()
	h.Create(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}
	result := parseJSON(t, w.Body)
	if result["title"] != "New Task" {
		t.Fatalf("expected 'New Task', got %v", result["title"])
	}
}

func TestTaskCreateMissingTitle(t *testing.T) {
	h := NewTaskHandler(&mockTaskStore{})

	body := `{"task_type":"dev"}`
	req := httptest.NewRequest("POST", "/api/tasks", bytes.NewBufferString(body))
	req = withUser(req)
	w := httptest.NewRecorder()
	h.Create(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestTaskCreateInvalidJSON(t *testing.T) {
	h := NewTaskHandler(&mockTaskStore{})

	req := httptest.NewRequest("POST", "/api/tasks", bytes.NewBufferString("{invalid"))
	req = withUser(req)
	w := httptest.NewRecorder()
	h.Create(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestTaskCreateDefaultsType(t *testing.T) {
	store := &mockTaskStore{}
	h := NewTaskHandler(store)

	body := `{"title":"No Type"}`
	req := httptest.NewRequest("POST", "/api/tasks", bytes.NewBufferString(body))
	req = withUser(req)
	w := httptest.NewRecorder()
	h.Create(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", w.Code)
	}
	result := parseJSON(t, w.Body)
	if result["task_type"] != "dev" {
		t.Fatalf("expected task_type 'dev', got %v", result["task_type"])
	}
}

func TestTaskUpdate(t *testing.T) {
	id := uuid.New()
	store := &mockTaskStore{
		tasks: []db.Task{{ID: id, Title: "Old", TaskType: db.TaskTypeDev, State: db.TaskStateDraft}},
	}
	h := NewTaskHandler(store)

	body := `{"title":"Updated"}`
	req := httptest.NewRequest("PUT", "/api/tasks/"+id.String(), bytes.NewBufferString(body))
	req = withChiParam(req, "id", id.String())
	w := httptest.NewRecorder()
	h.Update(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	result := parseJSON(t, w.Body)
	if result["title"] != "Updated" {
		t.Fatalf("expected 'Updated', got %v", result["title"])
	}
}

func TestTaskUpdateInvalidTransition(t *testing.T) {
	id := uuid.New()
	store := &mockTaskStore{
		tasks: []db.Task{{ID: id, Title: "Task", TaskType: db.TaskTypeDev, State: db.TaskStateDraft}},
	}
	h := NewTaskHandler(store)

	body := `{"state":"done"}`
	req := httptest.NewRequest("PUT", "/api/tasks/"+id.String(), bytes.NewBufferString(body))
	req = withChiParam(req, "id", id.String())
	w := httptest.NewRecorder()
	h.Update(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestTaskUpdateValidTransition(t *testing.T) {
	id := uuid.New()
	store := &mockTaskStore{
		tasks: []db.Task{{ID: id, Title: "Task", TaskType: db.TaskTypeDev, State: db.TaskStateDraft}},
	}
	h := NewTaskHandler(store)

	body := `{"state":"refine"}`
	req := httptest.NewRequest("PUT", "/api/tasks/"+id.String(), bytes.NewBufferString(body))
	req = withChiParam(req, "id", id.String())
	w := httptest.NewRecorder()
	h.Update(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestTaskDelete(t *testing.T) {
	id := uuid.New()
	store := &mockTaskStore{
		tasks: []db.Task{{ID: id, Title: "Task"}},
	}
	h := NewTaskHandler(store)

	req := httptest.NewRequest("DELETE", "/api/tasks/"+id.String(), nil)
	req = withChiParam(req, "id", id.String())
	w := httptest.NewRecorder()
	h.Delete(w, req)

	if w.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", w.Code)
	}
}

func TestTaskDeleteNotFound(t *testing.T) {
	h := NewTaskHandler(&mockTaskStore{})

	id := uuid.New()
	req := httptest.NewRequest("DELETE", "/api/tasks/"+id.String(), nil)
	req = withChiParam(req, "id", id.String())
	w := httptest.NewRecorder()
	h.Delete(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

// ── Project Handler Tests ───────────────────────────

func TestProjectList(t *testing.T) {
	store := &mockProjectStore{
		projects: []db.Project{
			{ID: uuid.New(), Name: "Project A", Color: "#3b82f6"},
		},
	}
	h := NewProjectHandler(store)

	req := httptest.NewRequest("GET", "/api/projects", nil)
	w := httptest.NewRecorder()
	h.List(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var projects []db.Project
	json.Unmarshal(w.Body.Bytes(), &projects)
	if len(projects) != 1 {
		t.Fatalf("expected 1 project, got %d", len(projects))
	}
}

func TestProjectCreate(t *testing.T) {
	store := &mockProjectStore{}
	h := NewProjectHandler(store)

	body := `{"name":"Test Project"}`
	req := httptest.NewRequest("POST", "/api/projects", bytes.NewBufferString(body))
	req = withUser(req)
	w := httptest.NewRecorder()
	h.Create(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}
	result := parseJSON(t, w.Body)
	if result["name"] != "Test Project" {
		t.Fatalf("expected 'Test Project', got %v", result["name"])
	}
}

func TestProjectCreateMissingName(t *testing.T) {
	h := NewProjectHandler(&mockProjectStore{})

	body := `{"description":"no name"}`
	req := httptest.NewRequest("POST", "/api/projects", bytes.NewBufferString(body))
	req = withUser(req)
	w := httptest.NewRecorder()
	h.Create(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestProjectGetNotFound(t *testing.T) {
	h := NewProjectHandler(&mockProjectStore{})

	id := uuid.New()
	req := httptest.NewRequest("GET", "/api/projects/"+id.String(), nil)
	req = withChiParam(req, "id", id.String())
	w := httptest.NewRecorder()
	h.Get(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestProjectDelete(t *testing.T) {
	id := uuid.New()
	store := &mockProjectStore{
		projects: []db.Project{{ID: id, Name: "Test"}},
	}
	h := NewProjectHandler(store)

	req := httptest.NewRequest("DELETE", "/api/projects/"+id.String(), nil)
	req = withChiParam(req, "id", id.String())
	w := httptest.NewRecorder()
	h.Delete(w, req)

	if w.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", w.Code)
	}
}

// ── Health Handler Test ─────────────────────────────

func TestHealthCheck(t *testing.T) {
	req := httptest.NewRequest("GET", "/api/health", nil)
	w := httptest.NewRecorder()
	HealthCheck(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	result := parseJSON(t, w.Body)
	if result["status"] != "ok" {
		t.Fatalf("expected status 'ok', got %v", result["status"])
	}
}
