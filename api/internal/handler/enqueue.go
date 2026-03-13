package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/kwstnr-jc/task-orchestrator/api/internal/config"
	"github.com/kwstnr-jc/task-orchestrator/api/internal/db"
)

type EnqueueHandler struct {
	store    db.TaskStore
	sqsURL   string
	sqsClient *sqs.Client
}

func NewEnqueueHandler(store db.TaskStore, cfg *config.Config) *EnqueueHandler {
	h := &EnqueueHandler{
		store:  store,
		sqsURL: cfg.SQSQueueURL,
	}

	if cfg.SQSQueueURL != "" {
		awsCfg, err := awsconfig.LoadDefaultConfig(context.Background(),
			awsconfig.WithRegion(cfg.AWSRegion),
		)
		if err == nil {
			h.sqsClient = sqs.NewFromConfig(awsCfg)
		}
	}
	return h
}

type enqueueMessage struct {
	TaskID   string `json:"task_id"`
	Title    string `json:"title"`
	TaskType string `json:"task_type"`
}

func (h *EnqueueHandler) Enqueue(w http.ResponseWriter, r *http.Request) {
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

	if h.sqsClient == nil || h.sqsURL == "" {
		writeError(w, http.StatusServiceUnavailable, "SQS not configured")
		return
	}

	msg := enqueueMessage{
		TaskID:   task.ID.String(),
		Title:    task.Title,
		TaskType: string(task.TaskType),
	}
	body, _ := json.Marshal(msg)

	_, err = h.sqsClient.SendMessage(r.Context(), &sqs.SendMessageInput{
		QueueUrl:    aws.String(h.sqsURL),
		MessageBody: aws.String(string(body)),
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to enqueue task")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"status":  "enqueued",
		"task_id": task.ID.String(),
	})
}
