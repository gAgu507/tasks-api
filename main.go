package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

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

func main() {
	ctx := context.Background()

	pool, err := pgxpool.New(ctx, os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal("connect to db: ", err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		log.Fatal("ping db: ", err)
	}

	server := &Server{store: &PostgresStore{pool: pool}}

	// tasks := []Task{
	// 	{ID: 1, Title: "Task 1", Description: "Description for Task 1", Completed: false},
	// 	{ID: 2, Title: "Task 2", Description: "Description for Task 2", Completed: true},
	// }
	// server := &Server{store: &SliceStore{tasks: tasks}}

	mux := http.NewServeMux()
	mux.HandleFunc("/", server.handlerRoot)
	mux.HandleFunc("GET  /tasks", server.handleListTasks)
	mux.HandleFunc("GET /tasks/{id}", server.handleGetTask)
	mux.HandleFunc("PUT /tasks/{id}", server.handleChangeTask)
	mux.HandleFunc("DELETE /tasks/{id}", server.handleDeleteTask)
	mux.HandleFunc("POST /tasks", server.handleCreateTask)

	log.Println("Server is running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
