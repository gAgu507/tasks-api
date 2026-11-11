package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"sync"
)

type Task struct {
	ID          int    `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Completed   bool   `json:"completed"`
}

type TaskStore interface {
	List() []Task
	Create(task Task)
	Change(task Task)
	Delete(id int)
}

type SliceStore struct {
	mu    sync.Mutex
	tasks []Task
}

func (store *SliceStore) List() []Task {
	store.mu.Lock()
	defer store.mu.Unlock()
	return append([]Task(nil), store.tasks...)
}

func (store *SliceStore) Create(task Task) {
	store.mu.Lock()
	defer store.mu.Unlock()
	store.tasks = append(store.tasks, task)
}

func (store *SliceStore) Change(task Task) {
	store.mu.Lock()
	defer store.mu.Unlock()
	for i, t := range store.tasks {
		if t.ID == task.ID {
			store.tasks[i] = task
			return
		}
	}
}

func (store *SliceStore) Delete(id int) {
	store.mu.Lock()
	defer store.mu.Unlock()
	for i, t := range store.tasks {
		if t.ID == id {
			store.tasks = append(store.tasks[:i], store.tasks[i+1:]...)
			return
		}
	}
}

type Server struct {
	store TaskStore
}

func (s *Server) handlerRoot(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "You've reached the root endpoint!")
}

func (s *Server) handleListTasks(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(s.store.List())
}

func (s *Server) handleGetTask(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr[1 : len(idStr)-1])
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(s.store.List()[id-1])

}

func (s *Server) handlePing(w http.ResponseWriter, r *http.Request) {
	status := map[string]string{"status": "ok"}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)

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
	s.store.Create(Task{
		ID:          len(s.store.List()) + 1,
		Title:       in.Title,
		Description: in.Description,
		Completed:   false,
	})
}

func (s *Server) handleDeleteTask(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr[1 : len(idStr)-1])
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	s.store.Delete(id)
}

func (s *Server) handleChangeTask(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr[1 : len(idStr)-1])
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var in struct {
		Title       string `json:"title"`
		Description string `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	tasks := s.store.List()
	if id < 1 || id > len(tasks) {
		http.Error(w, "task not found", http.StatusNotFound)
		return
	}

	tasks[id-1] = Task{
		ID:          id,
		Title:       in.Title,
		Description: in.Description,
	}
	s.store.Change(tasks[id-1])
}

func main() {

	tasks := []Task{
		{ID: 1, Title: "Task 1", Description: "Description for Task 1", Completed: false},
		{ID: 2, Title: "Task 2", Description: "Description for Task 2", Completed: true},
	}
	server := &Server{store: &SliceStore{tasks: tasks}}

	http.HandleFunc("/", server.handlerRoot)
	http.HandleFunc("/ping", server.handlePing)
	http.HandleFunc("GET  /tasks", server.handleListTasks)
	http.HandleFunc("GET /tasks/{id}", server.handleGetTask)
	http.HandleFunc("PUT /tasks/{id}", server.handleChangeTask)
	http.HandleFunc("DELETE /tasks/{id}", server.handleDeleteTask)
	http.HandleFunc("POST /tasks", server.handleCreateTask)

	fmt.Println("Server is running on http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}
