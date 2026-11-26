package handlers

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/gAgu507/tasks-api/internal/store"
)

func newTwoUsersStore() *fakeStore {
	return &fakeStore{
		tasks:  []store.Task{{ID: 1, Title: "mine"}, {ID: 2, Title: "theirs"}},
		owners: map[int]int{1: testUserID, 2: 2},
	}
}

func TestList_OnlyOwnTasks(t *testing.T) {
	rec := do(newTwoUsersStore(), http.MethodGet, "/tasks")

	var got []store.Task
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("bad json: %v", err)
	}
	if len(got) != 1 || got[0].Title != "mine" {
		t.Errorf("got %+v, want only the task \"mine\"", got)
	}
}

func TestForeignTask_Hidden(t *testing.T) {
	tests := []struct {
		name   string
		method string
		body   string
	}{
		{"get", http.MethodGet, ""},
		{"put", http.MethodPut, `{"title":"hacked","completed":true}`},
		{"delete", http.MethodDelete, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := newTwoUsersStore()
			rec := doBody(fs, tt.method, "/tasks/2", tt.body)

			if rec.Code != http.StatusNotFound {
				t.Fatalf("status = %d, want 404", rec.Code)
			}
			// чужая задача осталась целой
			if len(fs.tasks) != 2 || fs.tasks[1].Title != "theirs" {
				t.Errorf("foreign task was modified: %+v", fs.tasks)
			}
		})
	}
}

func TestCreate_AssignsCurrentUser(t *testing.T) {
	fs := &fakeStore{}
	rec := doBody(fs, http.MethodPost, "/tasks", `{"title":"new"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201", rec.Code)
	}
	if fs.owners[1] != testUserID {
		t.Errorf("owner = %d, want %d", fs.owners[1], testUserID)
	}
}
