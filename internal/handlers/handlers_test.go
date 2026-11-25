package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gAgu507/tasks-api/internal/auth"
	"github.com/gAgu507/tasks-api/internal/store"
	"golang.org/x/crypto/bcrypt"
)

var testSecret = []byte("test-secret-test-secret-test-secret")

type fakeStore struct {
	tasks []store.Task
	err   error
}

func (f *fakeStore) List(ctx context.Context) ([]store.Task, error) {
	if f.err != nil {
		return nil, f.err
	}
	return append([]store.Task{}, f.tasks...), nil
}

func (f *fakeStore) Get(ctx context.Context, id int) (store.Task, error) {
	if f.err != nil {
		return store.Task{}, f.err
	}
	for _, t := range f.tasks {
		if t.ID == id {
			return t, nil
		}
	}
	return store.Task{}, store.ErrNotFound
}

func (f *fakeStore) Create(ctx context.Context, task store.Task) (store.Task, error) {
	if f.err != nil {
		return store.Task{}, f.err
	}
	task.ID = len(f.tasks) + 1
	f.tasks = append(f.tasks, task)
	return task, nil
}

func (f *fakeStore) Change(ctx context.Context, task store.Task) error {
	if f.err != nil {
		return f.err
	}
	for i, t := range f.tasks {
		if t.ID == task.ID {
			f.tasks[i] = task
			return nil
		}
	}
	return store.ErrNotFound
}

func (f *fakeStore) Delete(ctx context.Context, id int) error {
	if f.err != nil {
		return f.err
	}
	for i, t := range f.tasks {
		if t.ID == id {
			f.tasks = append(f.tasks[:i], f.tasks[i+1:]...)
			return nil
		}
	}
	return store.ErrNotFound
}

type fakeUsers struct {
	users []store.User
	err   error
}

func (f *fakeUsers) CreateUser(ctx context.Context, email, hash string) (store.User, error) {
	if f.err != nil {
		return store.User{}, f.err
	}
	for _, u := range f.users {
		if u.Email == email {
			return store.User{}, store.ErrEmailTaken
		}
	}
	u := store.User{ID: len(f.users) + 1, Email: email, PasswordHash: hash}
	f.users = append(f.users, u)
	return u, nil
}

func (f *fakeUsers) GetUserByEmail(ctx context.Context, email string) (store.User, error) {
	if f.err != nil {
		return store.User{}, f.err
	}
	for _, u := range f.users {
		if u.Email == email {
			return u, nil
		}
	}
	return store.User{}, store.ErrNotFound
}

// Test for Tasks
func do(fs *fakeStore, method, path string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, nil)
	rec := httptest.NewRecorder()
	NewServer(fs, &fakeUsers{}, testSecret).Routes().ServeHTTP(rec, req)
	return rec
}

func TestRoot(t *testing.T) {
	rec := do(&fakeStore{}, http.MethodGet, "/")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestListTaskEmpty(t *testing.T) {
	rec := do(&fakeStore{}, http.MethodGet, "/tasks")

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if body := strings.TrimSpace(rec.Body.String()); body != "[]" {
		t.Errorf("body = %q, want %q", body, "[]")
	}
}

func TestListTasks_NotEmpty(t *testing.T) {
	fs := &fakeStore{tasks: []store.Task{
		{ID: 1, Title: "first"},
		{ID: 2, Title: "second", Completed: true},
	}}
	rec := do(fs, http.MethodGet, "/tasks")

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	var got []store.Task
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("bad json: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("got %d tasks, want 2", len(got))
	}
	if got[1].Title != "second" || !got[1].Completed {
		t.Errorf("unexpected second task: %+v", got[1])
	}
}

func TestListTasks_StoreError(t *testing.T) {
	fs := &fakeStore{err: context.DeadlineExceeded}
	rec := do(fs, http.MethodGet, "/tasks")

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
	}
}

func doBody(fs *fakeStore, method, path, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	NewServer(fs, &fakeUsers{}, testSecret).Routes().ServeHTTP(rec, req)
	return rec
}

func TestGetTask(t *testing.T) {
	tests := []struct {
		name      string
		path      string
		wantCode  int
		wantTitle string
	}{
		{"found", "/tasks/1", http.StatusOK, "first"},
		{"found after gap", "/tasks/3", http.StatusOK, "third"},
		{"deleted id", "/tasks/2", http.StatusNotFound, ""},
		{"unknown id", "/tasks/999", http.StatusNotFound, ""},
		{"not a number", "/tasks/abc", http.StatusBadRequest, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := &fakeStore{tasks: []store.Task{
				{ID: 1, Title: "first"},
				{ID: 3, Title: "third"},
			}}
			rec := do(fs, http.MethodGet, tt.path)

			if rec.Code != tt.wantCode {
				t.Fatalf("status = %d, want %d", rec.Code, tt.wantCode)
			}
			if tt.wantTitle != "" {
				var got store.Task
				if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
					t.Fatalf("bad json: %v", err)
				}
				if got.Title != tt.wantTitle {
					t.Errorf("title = %q, want %q", got.Title, tt.wantTitle)
				}
			}
		})
	}
}

func TestCreateTask(t *testing.T) {
	tests := []struct {
		name     string
		body     string
		wantCode int
	}{
		{"valid", `{"title":"buy bread","description":"rye"}`, http.StatusCreated},
		{"empty title", `{"title":""}`, http.StatusBadRequest},
		{"no title field", `{"description":"x"}`, http.StatusBadRequest},
		{"broken json", `{"title":`, http.StatusBadRequest},
		{"empty body", ``, http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := doBody(&fakeStore{}, http.MethodPost, "/tasks", tt.body)
			if rec.Code != tt.wantCode {
				t.Errorf("status = %d, want %d", rec.Code, tt.wantCode)
			}
		})
	}
}

func TestCreateTask_Response(t *testing.T) {
	fs := &fakeStore{}
	rec := doBody(fs, http.MethodPost, "/tasks", `{"title":"buy bread"}`)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusCreated)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", ct)
	}

	var got store.Task
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("bad json: %v", err)
	}
	if got.ID != 1 || got.Title != "buy bread" || got.Completed {
		t.Errorf("unexpected task: %+v", got)
	}
	if len(fs.tasks) != 1 {
		t.Errorf("store has %d tasks, want 1", len(fs.tasks))
	}
}

func TestCreateTask_StoreError(t *testing.T) {
	fs := &fakeStore{err: context.DeadlineExceeded}
	rec := doBody(fs, http.MethodPost, "/tasks", `{"title":"buy bread"}`)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
	}
}

func TestChangeAndDelete(t *testing.T) {
	tests := []struct {
		name     string
		method   string
		path     string
		body     string
		wantCode int
	}{
		{"put ok", http.MethodPut, "/tasks/1", `{"title":"new","completed":true}`, http.StatusNoContent},
		{"put not found", http.MethodPut, "/tasks/9", `{"title":"new"}`, http.StatusNotFound},
		{"put bad id", http.MethodPut, "/tasks/abc", `{"title":"new"}`, http.StatusBadRequest},
		{"put bad json", http.MethodPut, "/tasks/1", `{`, http.StatusBadRequest},
		{"delete ok", http.MethodDelete, "/tasks/1", "", http.StatusNoContent},
		{"delete not found", http.MethodDelete, "/tasks/9", "", http.StatusNotFound},
		{"delete bad id", http.MethodDelete, "/tasks/abc", "", http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := &fakeStore{tasks: []store.Task{{ID: 1, Title: "first"}}}
			rec := doBody(fs, tt.method, tt.path, tt.body)
			if rec.Code != tt.wantCode {
				t.Errorf("status = %d, want %d", rec.Code, tt.wantCode)
			}
		})
	}
}

// Tests for Users
func TestRegister(t *testing.T) {
	tests := []struct {
		name     string
		body     string
		wantCode int
	}{
		{"valid", `{"email":"Anna@Example.com","password":"secret123"}`, http.StatusCreated},
		{"duplicate", `{"email":"taken@example.com","password":"secret123"}`, http.StatusConflict},
		{"bad email", `{"email":"nope","password":"secret123"}`, http.StatusBadRequest},
		{"short password", `{"email":"a@example.com","password":"123"}`, http.StatusBadRequest},
		{"broken json", `{"email":`, http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			users := &fakeUsers{users: []store.User{{ID: 1, Email: "taken@example.com"}}}
			req := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(tt.body))
			rec := httptest.NewRecorder()
			NewServer(&fakeStore{}, users, testSecret).Routes().ServeHTTP(rec, req)

			if rec.Code != tt.wantCode {
				t.Fatalf("status = %d, want %d", rec.Code, tt.wantCode)
			}
			if rec.Code == http.StatusCreated {
				body := rec.Body.String()
				if strings.Contains(body, "password") || strings.Contains(body, "$2") {
					t.Errorf("response leaks password data: %s", body)
				}
			}
		})
	}
}

func TestLogin(t *testing.T) {
	hash, err := bcrypt.GenerateFromPassword([]byte("secret123"), bcrypt.MinCost)
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name     string
		body     string
		wantCode int
	}{
		{"ok", `{"email":"anna@example.com","password":"secret123"}`, http.StatusOK},
		{"email case ignored", `{"email":"ANNA@example.com","password":"secret123"}`, http.StatusOK},
		{"wrong password", `{"email":"anna@example.com","password":"wrongpass"}`, http.StatusUnauthorized},
		{"unknown email", `{"email":"bob@example.com","password":"secret123"}`, http.StatusUnauthorized},
		{"broken json", `{`, http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			users := &fakeUsers{users: []store.User{
				{ID: 7, Email: "anna@example.com", PasswordHash: string(hash)},
			}}
			req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(tt.body))
			rec := httptest.NewRecorder()
			NewServer(&fakeStore{}, users, testSecret).Routes().ServeHTTP(rec, req)

			if rec.Code != tt.wantCode {
				t.Fatalf("status = %d, want %d", rec.Code, tt.wantCode)
			}
			if tt.wantCode != http.StatusOK {
				return
			}
			var got struct{ Token string }
			if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
				t.Fatalf("bad json: %v", err)
			}
			id, err := auth.ParseToken(testSecret, got.Token)
			if err != nil || id != 7 {
				t.Errorf("token user id = %d, err = %v, want 7", id, err)
			}
		})
	}
}
