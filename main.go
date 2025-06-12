package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
)

var (
	tasks     []Task
	idCounter int
	mu        sync.Mutex
	filePath  = "tasks.json"
	count     = 0
)

type Task struct {
	ID        int    `json:"id"`
	Title     string `json:"title"`
	Completed bool   `json:"completed"`
}

// Load tasks from file
func loadTasks() error {
	file, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			tasks = []Task{}
			return nil
		}
		return err
	}
	err = json.Unmarshal(file, &tasks)
	if err != nil {
		return err
	}

	for _, t := range tasks {
		if t.ID > idCounter {
			idCounter = t.ID
		}
	}
	return nil
}

// Save tasks to file
func saveTasks() error {
	data, err := json.MarshalIndent(tasks, "", " ")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(filePath, data, 0644)
}

// Middleware: CORS + Content-Type
func HeaderMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("Content-Type", "application/json")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// Get all tasks
func GetTasks(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()
	json.NewEncoder(w).Encode(tasks)
}

// Get single task by ID
func GetSingleTask(w http.ResponseWriter, r *http.Request, id int) {
	mu.Lock()
	defer mu.Unlock()
	for _, task := range tasks {
		if task.ID == id {
			json.NewEncoder(w).Encode(task)
			return
		}
	}
	http.Error(w, "Task not found", http.StatusNotFound)
}

// Add new task
func AddTask(w http.ResponseWriter, r *http.Request) {
	var t Task
	if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	mu.Lock()
	idCounter++
	t.ID = idCounter
	tasks = append(tasks, t)
	_ = saveTasks()
	mu.Unlock()

	count++
	log.Println("Task created:", count)
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(t)
}

// Update task by ID
func UpdateTask(w http.ResponseWriter, r *http.Request, id int) {
	var updated Task
	if err := json.NewDecoder(r.Body).Decode(&updated); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	mu.Lock()
	defer mu.Unlock()
	for i, task := range tasks {
		if task.ID == id {
			tasks[i].Title = updated.Title
			tasks[i].Completed = updated.Completed
			_ = saveTasks()
			json.NewEncoder(w).Encode(tasks[i])
			return
		}
	}
	http.Error(w, "Task not found", http.StatusNotFound)
}

// Delete task by ID
func DeleteTask(w http.ResponseWriter, r *http.Request, id int) {
	mu.Lock()
	defer mu.Unlock()
	for i, task := range tasks {
		if task.ID == id {
			tasks = append(tasks[:i], tasks[i+1:]...)
			_ = saveTasks()
			w.WriteHeader(http.StatusNoContent)
			return
		}
	}
	http.Error(w, "Task not found", http.StatusNotFound)
}

// Helper to extract ID from path
func getIDFromPath(path string) (int, error) {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) < 2 {
		return 0, http.ErrMissingFile
	}
	return strconv.Atoi(parts[1])
}

// Handle /tasks and /tasks/{id}
func TaskHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/tasks" {
		switch r.Method {
		case http.MethodGet:
			GetTasks(w, r)
		case http.MethodPost:
			AddTask(w, r)
		default:
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		}
		return
	}

	// Handle /tasks/{id}
	id, err := getIDFromPath(r.URL.Path)
	if err != nil {
		http.Error(w, "Invalid or missing ID", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodGet:
		GetSingleTask(w, r, id)
	case http.MethodPut:
		UpdateTask(w, r, id)
	case http.MethodDelete:
		DeleteTask(w, r, id)
	default:
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}

func DefaultHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Welcome to Task API"})
}

func main() {
	if err := loadTasks(); err != nil {
		log.Fatalf("Failed to load tasks: %v", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", DefaultHandler)
	mux.HandleFunc("/tasks", TaskHandler)
	mux.HandleFunc("/tasks/", TaskHandler)

	wrapped := HeaderMiddleware(mux)

	log.Println("Server started on http://localhost:8080")
	if err := http.ListenAndServe(":8080", wrapped); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
