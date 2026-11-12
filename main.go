package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("task not found")

type Task struct {
	ID          int    `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Completed   bool   `json:"completed"`
}

type TaskStore interface {
	List(context.Context) []Task
	Create(context.Context, Task)
	Change(context.Context, Task) error
	Delete(context.Context, int) error
	Get(context.Context, int) (Task, error)
}

// type SliceStore struct {
// 	mu    sync.Mutex
// 	tasks []Task
// }

// func (store *SliceStore) List(ctx context.Context) []Task {
// 	store.mu.Lock()
// 	defer store.mu.Unlock()
// 	return append([]Task(nil), store.tasks...)
// }

// func (store *SliceStore) Create(ctx context.Context, task Task) {
// 	store.mu.Lock()
// 	defer store.mu.Unlock()
// 	store.tasks = append(store.tasks, task)
// }

// func (store *SliceStore) Change(ctx context.Context, task Task) error {
// 	store.mu.Lock()
// 	defer store.mu.Unlock()
// 	for i, t := range store.tasks {
// 		if t.ID == task.ID {
// 			store.tasks[i] = task
// 			return nil
// 		}
// 	}
// 	return ErrNotFound
// }

// func (store *SliceStore) Delete(id int) error {
// 	store.mu.Lock()
// 	defer store.mu.Unlock()
// 	for i, t := range store.tasks {
// 		if t.ID == id {
// 			store.tasks = append(store.tasks[:i], store.tasks[i+1:]...)
// 			return nil
// 		}
// 	}
// 	return ErrNotFound
// }

type PostgresStore struct {
	pool *pgxpool.Pool
}

func (store *PostgresStore) List(ctx context.Context) []Task {
	rows, err := store.pool.Query(ctx, "SELECT id, title, description, completed FROM tasks")
	if err != nil {
		fmt.Println("Error querying tasks:", err)
		return nil
	}

	defer rows.Close()

	var tasks []Task
	for rows.Next() {
		var task Task
		if err := rows.Scan(&task.ID, &task.Title, &task.Description, &task.Completed); err != nil {
			fmt.Println("Error scanning task:", err)
			return nil
		}
		tasks = append(tasks, task)
	}

	if err := rows.Err(); err != nil {
		fmt.Println("Error iterating over rows:", err)
		return nil
	}

	return tasks
}

func (store *PostgresStore) Create(ctx context.Context, task Task) {
	_, err := store.pool.Exec(ctx, "INSERT INTO tasks (title, description, completed) VALUES ($1, $2, $3)", task.Title, task.Description, task.Completed)
	if err != nil {
		fmt.Println("Error inserting task:", err)
	}
}

func (store *PostgresStore) Change(ctx context.Context, task Task) error {
	tag, err := store.pool.Exec(ctx, "UPDATE tasks SET title=$1, description=$2, completed=$3 WHERE id=$4", task.Title, task.Description, task.Completed, task.ID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (store *PostgresStore) Delete(ctx context.Context, id int) error {
	tag, err := store.pool.Exec(ctx, "DELETE FROM tasks WHERE id=$1", id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (store *PostgresStore) Get(ctx context.Context, id int) (Task, error) {
	var task Task
	row := store.pool.QueryRow(ctx, "SELECT id, title, description, completed FROM tasks WHERE id=$1", id)
	err := row.Scan(&task.ID, &task.Title, &task.Description, &task.Completed)
	if errors.Is(err, pgx.ErrNoRows) {
		return Task{}, ErrNotFound
	}
	if err != nil {
		return Task{}, err
	}
	return task, nil
}

type Server struct {
	store TaskStore
}

func (s *Server) handlerRoot(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "You've reached the root endpoint!")
}

func (s *Server) handleListTasks(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(s.store.List(r.Context()))
}

func (s *Server) handleGetTask(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "bad id", http.StatusBadRequest)
		return
	}

	task, err := s.store.Get(r.Context(), id)
	if errors.Is(err, ErrNotFound) {
		http.Error(w, "task not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(task)
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
	s.store.Create(r.Context(), Task{
		ID:          len(s.store.List(r.Context())) + 1,
		Title:       in.Title,
		Description: in.Description,
		Completed:   false,
	})
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

	err = s.store.Change(r.Context(), Task{
		ID:          id,
		Title:       in.Title,
		Description: in.Description,
		Completed:   in.Completed,
	})
	if errors.Is(err, ErrNotFound) {
		http.Error(w, "task not found", http.StatusNotFound)
		return
	}
	if err != nil {
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
	if errors.Is(err, ErrNotFound) {
		http.Error(w, "task not found", http.StatusNotFound)
		return
	}
	if err != nil {
		log.Println("Error deleting task:", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func main() {
	ctx := context.Background()

	pool, err := pgxpool.New(ctx, os.Getenv("DATABASE_URL"))
	if err != nil {
		fmt.Println("Error connecting to the database:", err)
		return
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		fmt.Println("Error pinging the database:", err)
		return
	}

	server := &Server{store: &PostgresStore{pool: pool}}
	// tasks := []Task{
	// 	{ID: 1, Title: "Task 1", Description: "Description for Task 1", Completed: false},
	// 	{ID: 2, Title: "Task 2", Description: "Description for Task 2", Completed: true},
	// }
	// server := &Server{store: &SliceStore{tasks: tasks}}

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
