package handlers

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gAgu507/tasks-api/internal/auth"
	"github.com/gAgu507/tasks-api/internal/store"
	"golang.org/x/crypto/bcrypt"
)

type Server struct {
	store     store.TaskStore
	users     store.UserStore
	jwtSecret []byte
}

func NewServer(tasks store.TaskStore, users store.UserStore, jwtSecret []byte) *Server {
	return &Server{store: tasks, users: users, jwtSecret: jwtSecret}
}

func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /{$}", s.handlerRoot)
	mux.HandleFunc("GET /tasks", s.handleListTasks)
	mux.HandleFunc("POST /tasks", s.handleCreateTask)
	mux.HandleFunc("GET /tasks/{id}", s.handleGetTask)
	mux.HandleFunc("PUT /tasks/{id}", s.handleChangeTask)
	mux.HandleFunc("DELETE /tasks/{id}", s.handleDeleteTask)
	mux.HandleFunc("POST /register", s.handleRegister)
	mux.HandleFunc("POST /login", s.handleLogin)
	return mux
}

// Part with handlers for Tasks.
func (s *Server) handlerRoot(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) handleListTasks(w http.ResponseWriter, r *http.Request) {
	tasks, err := s.store.List(r.Context())
	if err != nil {
		log.Println("list tasks: ", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, tasks)
}

func (s *Server) handleGetTask(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "bad id", http.StatusBadRequest)
		return
	}

	task, err := s.store.Get(r.Context(), id)
	if errors.Is(err, store.ErrNotFound) {
		http.Error(w, "task not found", http.StatusNotFound)
		return
	}
	if err != nil {
		log.Println("get task: ", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, task)
}

func (s *Server) handleCreateTask(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Title       string `json:"title"`
		Description string `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if in.Title == "" {
		http.Error(w, "title is required", http.StatusBadRequest)
		return
	}
	task, err := s.store.Create(r.Context(), store.Task{
		Title:       in.Title,
		Description: in.Description,
	})
	if err != nil {
		log.Println("create task ", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusCreated, task)
}

func (s *Server) handleChangeTask(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "bad id", http.StatusBadRequest)
		return
	}

	var in struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		Completed   bool   `json:"completed"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	err = s.store.Change(r.Context(), store.Task{
		ID:          id,
		Title:       in.Title,
		Description: in.Description,
		Completed:   in.Completed,
	})
	if errors.Is(err, store.ErrNotFound) {
		http.Error(w, "task not found", http.StatusNotFound)
		return
	}
	if err != nil {
		log.Println("change task: ", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleDeleteTask(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "bad id", http.StatusBadRequest)
		return
	}

	err = s.store.Delete(r.Context(), id)
	if errors.Is(err, store.ErrNotFound) {
		http.Error(w, "task not found", http.StatusNotFound)
		return
	}
	if err != nil {
		log.Println("delete task: ", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// Part with handlers for Users.
func (s *Server) handleRegister(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	email := strings.ToLower(strings.TrimSpace(in.Email))
	if !strings.Contains(email, "@") {
		http.Error(w, "invalid email", http.StatusBadRequest)
		return
	}
	if len(in.Password) < 8 || len(in.Password) > 72 {
		http.Error(w, "password must be 8-72 characters", http.StatusBadRequest)
		return
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Println("hash password:", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	user, err := s.users.CreateUser(r.Context(), email, string(hash))
	if errors.Is(err, store.ErrEmailTaken) {
		http.Error(w, "email already registered", http.StatusConflict)
		return
	}
	if err != nil {
		log.Println("create user:", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{"id": user.ID, "email": user.Email})
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	email := strings.ToLower(strings.TrimSpace(in.Email))

	user, err := s.users.GetUserByEmail(r.Context(), email)
	if errors.Is(err, store.ErrNotFound) {
		http.Error(w, "invalid email or password", http.StatusUnauthorized)
		return
	}
	if err != nil {
		log.Println("get user:", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(in.Password))
	if err != nil {
		http.Error(w, "invalid email or password", http.StatusUnauthorized)
		return
	}

	token, err := auth.NewToken(s.jwtSecret, user.ID, 24*time.Hour)
	if err != nil {
		log.Println("create token:", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"token": token})
}
