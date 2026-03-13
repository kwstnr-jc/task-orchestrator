package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/kwstnr-jc/task-orchestrator/api/internal/auth"
	"github.com/kwstnr-jc/task-orchestrator/api/internal/db"
)

type TaskHandler struct {
	store db.TaskStore
}

func NewTaskHandler(store db.TaskStore) *TaskHandler {
	return &TaskHandler{store: store}
}

func (h *TaskHandler) List(w http.ResponseWriter, r *http.Request) {
	params := db.ListTasksParams{}
	if t := r.URL.Query().Get("type"); t != "" {
		tt := db.TaskType(t)
		params.TaskType = &tt
	}
	if s := r.URL.Query().Get("state"); s != "" {
		st := db.TaskState(s)
		params.State = &st
	}
	if p := r.URL.Query().Get("project_id"); p != "" {
		pid, err := uuid.Parse(p)
		if err == nil {
			params.ProjectID = &pid
		}
	}

	tasks, err := h.store.ListTasks(r.Context(), params)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, tasks)
}

func (h *TaskHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid task ID")
		return
	}

	task, err := h.store.GetTask(r.Context(), id)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			writeError(w, http.StatusNotFound, "task not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, task)
}

func (h *TaskHandler) Create(w http.ResponseWriter, r *http.Request) {
	user, _ := auth.UserFromContext(r.Context())

	var params db.CreateTaskParams
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if params.Title == "" {
		writeError(w, http.StatusBadRequest, "title is required")
		return
	}
	if params.TaskType == "" {
		params.TaskType = db.TaskTypeDev
	}
	params.CreatedBy = user.Username

	task, err := h.store.CreateTask(r.Context(), params)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, task)
}

func (h *TaskHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid task ID")
		return
	}

	var params db.UpdateTaskParams
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if params.State != nil {
		existing, err := h.store.GetTask(r.Context(), id)
		if err != nil {
			if errors.Is(err, db.ErrNotFound) {
				writeError(w, http.StatusNotFound, "task not found")
				return
			}
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		if !db.IsValidTransition(existing.TaskType, existing.State, *params.State) {
			writeError(w, http.StatusBadRequest,
				"invalid state transition from "+string(existing.State)+" to "+string(*params.State))
			return
		}
	}

	task, err := h.store.UpdateTask(r.Context(), id, params)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			writeError(w, http.StatusNotFound, "task not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, task)
}

func (h *TaskHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid task ID")
		return
	}

	if err := h.store.DeleteTask(r.Context(), id); err != nil {
		if errors.Is(err, db.ErrNotFound) {
			writeError(w, http.StatusNotFound, "task not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

