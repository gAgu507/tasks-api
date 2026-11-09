package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type Task struct {
	ID          int    `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Completed   bool   `json:"completed"`
}

func handlerRoot(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "You've reached the root endpoint!")
}

func handlePing(w http.ResponseWriter, r *http.Request) {
	status := map[string]string{"status": "ok"}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)

}

func handleListTasks(w http.ResponseWriter, r *http.Request) {
	tasks := []Task{
		{ID: 1, Title: "Task 1", Description: "Description for Task 1", Completed: false},
		{ID: 2, Title: "Task 2", Description: "Description for Task 2", Completed: true},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tasks)
}

func main() {
	http.HandleFunc("/", handlerRoot)
	http.HandleFunc("/ping", handlePing)
	http.HandleFunc("GET  /tasks", handleListTasks)
	fmt.Println("Server is running on http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}
