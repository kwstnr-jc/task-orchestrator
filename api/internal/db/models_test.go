package db

import "testing"

func TestIsValidTransition_Dev(t *testing.T) {
	tests := []struct {
		from, to TaskState
		want     bool
	}{
		{TaskStateDraft, TaskStateRefine, true},
		{TaskStateRefine, TaskStateApproved, true},
		{TaskStateRefine, TaskStateDraft, true},
		{TaskStateApproved, TaskStateInProgress, true},
		{TaskStateInProgress, TaskStateDone, true},
		{TaskStateInProgress, TaskStateApproved, true},
		// Invalid
		{TaskStateDraft, TaskStateApproved, false},
		{TaskStateDraft, TaskStateDone, false},
		{TaskStateDone, TaskStateDraft, false},
		{TaskStateApproved, TaskStateDraft, false},
	}
	for _, tt := range tests {
		got := IsValidTransition(TaskTypeDev, tt.from, tt.to)
		if got != tt.want {
			t.Errorf("Dev %s→%s: got %v, want %v", tt.from, tt.to, got, tt.want)
		}
	}
}

func TestIsValidTransition_Research(t *testing.T) {
	tests := []struct {
		from, to TaskState
		want     bool
	}{
		{TaskStateDraft, TaskStateInProgress, true},
		{TaskStateInProgress, TaskStateDone, true},
		{TaskStateInProgress, TaskStateDraft, true},
		// Invalid
		{TaskStateDraft, TaskStateDone, false},
		{TaskStateDraft, TaskStateRefine, false},
		{TaskStateDone, TaskStateDraft, false},
	}
	for _, tt := range tests {
		got := IsValidTransition(TaskTypeResearch, tt.from, tt.to)
		if got != tt.want {
			t.Errorf("Research %s→%s: got %v, want %v", tt.from, tt.to, got, tt.want)
		}
	}
}

func TestDefaultState(t *testing.T) {
	if got := DefaultState(TaskTypeDev); got != TaskStateDraft {
		t.Errorf("Dev default: got %s, want draft", got)
	}
	if got := DefaultState(TaskTypeResearch); got != TaskStateDraft {
		t.Errorf("Research default: got %s, want draft", got)
	}
}
