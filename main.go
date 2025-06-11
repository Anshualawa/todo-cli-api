package main
import (
	"net/http"
	"encoding/json"
	"log"
	"io/ioutil"
	"os"
	"sync"
)

var (
	tasks []Task
	idCounter int
	mu sync.Mutex
	filePath = "tasks.json"
	count = 0
)

// Load tasks from file
func loadTasks() error {
	file,err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			tasks = []Task{}
			return nil 
		}
		return err
	}
	err = json.Unmarshal(file,&tasks)
	if err != nil {
		return err
	}

	// Set idCounter to max exiting ID
	for _, t := range tasks {
		if t.ID > idCounter {
			idCounter = t.ID
		}
	}
	return nil 
}

// Save tasks to file 
func saveTasks() error {
	data,err := json.MarshalIndent(tasks,""," ")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(filePath,data,0644)
}


// Middleware : CORS + Content-Type
func HeaderMiddleware(next http.Handler) http.Handler{
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin","*")
		w.Header().Set("Access-Control-Allow-Methods","GET, POST, OPTIONS,PUT")
		w.Header().Set("Access-Control-Allow-Headers","Contetn-Type")
		w.Header().Set("Content-Type","application/json")

		if r.Method == http.MethodOptions{
			w.WriteHeader(http.StatusNoContent)
			return
		}

		// Call the next handler
		next.ServeHTTP(w,r)
	})
}


// Get tasks
func GetTasks(w http.ResponseWriter, r *http.Request){
	mu.Lock()
	defer mu.Unlock()
	json.NewEncoder(w).Encode(tasks)
}

// Add tasks
func AddTask(w http.ResponseWriter, r *http.Request){
	var t Task
	if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
		http.Error(w,err.Error(), http.StatusBadRequest)
		return
	}

	mu.Lock()
	idCounter++
	t.ID = idCounter
	tasks = append(tasks,t)
	_ = saveTasks()
	mu.Unlock()
	
	count++
	log.Println("task:",count)
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(t)
}
	
func main(){
	if err := loadTasks(); err != nil {
		log.Fatalf("Failed to load tasks: %v",err)
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/", DefaultHandler)
	mux.HandleFunc("/tasks", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
			case http.MethodGet:
				GetTasks(w,r)

			case http.MethodPost:
				AddTask(w,r)
			default:
				http.Error(w,"Method Not Allowed", http.StatusMethodNotAllowed)
			}
		})


	wrapped := HeaderMiddleware(mux)

	log.Println("Server start on http://localhost:8080")
	
	if err := http.ListenAndServe(":8080",wrapped); err != nil {
		log.Printf("Server failed to start: %v\n",err)
	}
}


type Task struct {
        ID              int     `json:"id"`
        Title           string  `json:"title"`
        Completed       bool    `json:"completed"`
}

func DefaultHandler(w http.ResponseWriter, r *http.Request) {
        var task Task
        w.WriteHeader(http.StatusOK)
        json.NewEncoder(w).Encode(task)
}
