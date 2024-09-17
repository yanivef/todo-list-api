package handlers

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"task-manager-api/db"
	task "task-manager-api/models"
	"time"
)

// POST
func CreateTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to parse body", http.StatusBadRequest)
		return
	}
	var data map[string]interface{}
	err = json.Unmarshal(body, &data)
	if err != nil {
		http.Error(w, "Failed to parse json", http.StatusBadRequest)
		return
	}

	if data["id"] == nil || data["name"] == nil {
		http.Error(w, "ID and name field are mandatory", http.StatusBadRequest)
		return
	}
	var done bool

	if data["done"] != nil {
		done = data["done"].(bool)
	} else {
		done = false
	}
	// TYPE ASSERTIONS
	id := int(data["id"].(float64)) // JSON numbers are parsed as float64
	name := (data["name"].(string))

	if name == "" {
		http.Error(w, "name must not be empty", http.StatusBadRequest)
		return
	}
	if exists, _ := db.TaskExists(id); exists {
		http.Error(w, "Task ID is taken", http.StatusBadRequest)
		return
	}
	// PARSE THE TIME FROM STRING
	var created_at time.Time
	var updated_at time.Time

	if data["created_at"] != nil {
		created_at, err = time.Parse(time.RFC3339, data["created_at"].(string))
		if err != nil {
			http.Error(w, "Invalid created_at format", http.StatusBadRequest)
			return
		}
	} else {
		created_at = time.Now()
	}
	if data["updated_at"] != nil {
		updated_at, err = time.Parse(time.RFC3339, data["updated_at"].(string))
		if err != nil {
			http.Error(w, "Invalid updated_at format", http.StatusBadRequest)
			return
		}
	} else {
		updated_at = time.Now()
	}

	// CHECK updated_at value is after or equal -> created_at
	if t := updated_at.Before(created_at); t {
		http.Error(w, "Invalid updated_at value (DATE CANT BE BEFORE created_at VALUE)", http.StatusBadRequest)
		return
	}

	new_task := task.Task{
		ID:        id,
		Name:      name,
		Done:      done,
		CreatedAt: created_at,
		UpdatedAt: updated_at,
	}
	log.Printf("POST request to create task: %v", new_task)
	// adding new task to DB
	err = db.InsertTask(new_task)
	if err != nil {
		log.Printf("Failed adding new task to database, task details: %v", new_task)
		http.Error(w, "Failed adding new task to database", http.StatusInternalServerError) // CHECK IF THIS IS THE CORRECT ERROR
		return
	}
}

// PUT
func UpdateTask(w http.ResponseWriter, r *http.Request) {

}

// GET
func GetTask(w http.ResponseWriter, r *http.Request) {

}

func GetAllTasks(w http.ResponseWriter, r *http.Request) {

}

// DELETE
func DeleteTask(w http.ResponseWriter, r *http.Request) {

}
