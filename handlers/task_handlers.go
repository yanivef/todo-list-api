package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"task-manager-api/db"
	"task-manager-api/models"

	"github.com/gorilla/mux"
)

func HandelTasks(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		// GetTasks(w, r)
		handleGetTasks(w, r)
	case http.MethodPost:
		CreateTask(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// function to create task
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

	// TYPE ASSERTIONS
	id, idOK := data["id"].(float64) // JSON numbers are parsed as float64
	name, nameOK := data["name"].(string)
	done, doneOK := data["done"].(bool)

	if name == "" || !idOK || !nameOK {
		http.Error(w, "ID and name required and name must not be empty", http.StatusBadRequest)
		return
	}

	if !doneOK {
		done = false
	}

	if exists, _ := db.TaskExists(int(id)); exists {
		http.Error(w, "Task ID is taken", http.StatusBadRequest)
		return
	}

	// create new task with the given details, set created_at, update_at to current time as default creation
	new_task := models.Task{
		ID:        int(id),
		Name:      name,
		Done:      done,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	log.Printf("POST request to create task: %v", new_task)

	// adding new task to DB
	err = db.InsertTask(new_task)
	if err != nil {
		log.Printf("Failed adding new task to database, task details: %v", new_task)
		http.Error(w, "Failed adding new task to database", http.StatusInternalServerError) // CHECK IF THIS IS THE CORRECT ERROR
		return
	}
	w.WriteHeader(http.StatusCreated) // indicate successful creation
}

// PUT
func UpdateTaskByID(w http.ResponseWriter, r *http.Request, id int) {
	if r.Method != http.MethodPut {
		http.Error(w, "Only PUT method is allowed", http.StatusMethodNotAllowed)
		return
	}
	body, err := io.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		http.Error(w, "Couldn't read body", http.StatusBadRequest)
		return
	}

	var data map[string]interface{}
	err = json.Unmarshal(body, &data)
	if err != nil {
		http.Error(w, "Couldn't parse json", http.StatusBadRequest)
		return
	}
	if len(data) == 0 {
		http.Error(w, "Request body cannot be empty", http.StatusBadRequest)
		return
	}

	var updates = make(map[string]interface{})

	if name, ok := data["name"].(string); ok {
		if name != "" {
			updates["name"] = name
		}
	}

	if done, ok := data["done"].(bool); ok {
		updates["done"] = done
	}

	// if created_atStr, ok := data["created_at"].(string); ok {
	// 	created_at, err := time.Parse(time.RFC3339, created_atStr)
	// 	if err != nil {
	// 		http.Error(w, "Invalid created_at format", http.StatusBadRequest)
	// 		return
	// 	}
	// 	updates["created_at"] = created_at
	// }

	// if updated_atStr, ok := data["updated_at"].(string); ok {
	// 	updated_at, err := time.Parse(time.RFC3339, updated_atStr)
	// 	if err != nil {
	// 		http.Error(w, "Invalid updated_at format", http.StatusBadRequest)
	// 		return
	// 	}
	// 	updates["updated_at"] = updated_at
	// }

	updates["updated_at"] = time.Now()

	err = db.UpdateTask(id, updates)
	if err != nil {
		http.Error(w, "Couldn't read task: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusAccepted)

}

// GET
func GetTaskByID(w http.ResponseWriter, r *http.Request, id int) {
	var task models.Task
	var err error

	task, err = db.GetTask(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	out, err := json.Marshal(task)
	if err != nil {
		http.Error(w, "Couldn't parse task", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(out)
}

func GetTasks(w http.ResponseWriter, r *http.Request) {
	// GET REQUEST -> return list of all tasks from database
	if r.Method != http.MethodGet {
		http.Error(w, "Only GET method is allowed", http.StatusMethodNotAllowed)
		return
	}
	tasks, err := db.GetAllTasks()
	if err != nil {
		http.Error(w, "Couldn't fetch tasks from database", http.StatusInternalServerError)
		return
	}
	out, err := json.Marshal(tasks)
	if err != nil {
		http.Error(w, "Couldn't parse tasks", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(out)
}

// DELETE
func DeleteTaskByID(w http.ResponseWriter, r *http.Request, id int) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Only DELETE method is allowed", http.StatusMethodNotAllowed)
		return
	}
	err := db.DeleteTask(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// handle [GET, DELETE, PUT] requests by task ID
func HandleTaskByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr, idOK := vars["id"]
	if !idOK {
		http.Error(w, "Parameter ID not found", http.StatusBadRequest)
		return
	}

	var err error
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid task ID", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodGet:
		GetTaskByID(w, r, id)

	case http.MethodDelete:
		DeleteTaskByID(w, r, id)
	case http.MethodPut:
		UpdateTaskByID(w, r, id)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return

	}
}

// handle get request to tasks URL, if id(one or more) provided then trigger 'GetMultipleTasksByID', otherwise, trigger 'GetTasks'
func handleGetTasks(w http.ResponseWriter, r *http.Request) {
	_, idOK := r.URL.Query()["id"]

	if idOK {
		GetMultipleTasksByID(w, r)
	} else {
		GetTasks(w, r)
	}
}

// handles multiple tasks id requests, passing ids to 'GetMultipleTasks' function
func GetMultipleTasksByID(w http.ResponseWriter, r *http.Request) {
	idsStr, idOK := r.URL.Query()["id"]

	if !idOK {
		http.Error(w, "No id provided", http.StatusBadRequest)
		return
	}

	var ids []int
	for _, val := range idsStr {
		id, err := strconv.Atoi(val)
		if err != nil {
			http.Error(w, "Invalid ID format", http.StatusBadRequest)
			return
		}
		ids = append(ids, id)
	}
	mulTasks, err := db.GetMultipleTasks(ids)

	if err != nil {
		http.Error(w, "Error fetching task: "+err.Error(), http.StatusNotFound)
		return
	}

	out, err := json.Marshal(mulTasks)

	if err != nil {
		http.Error(w, "Failed parse tasks do json", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(out)
}

func CreateUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	defer r.Body.Close()

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed parsing body", http.StatusBadRequest)
		return
	}

	var credentials map[string]interface{}

	err = json.Unmarshal(body, &credentials)
	if err != nil {
		http.Error(w, "Failed parsing JSON", http.StatusInternalServerError)
		return
	}

	username, usernameOK := credentials["username"].(string)
	password, passwordOK := credentials["password"].(string)
	email, emailOK := credentials["email"].(string)
	if !usernameOK || !passwordOK || !emailOK {
		http.Error(w, "please make sure to provide: username, password, email", http.StatusBadRequest)
		return
	}

	var user models.Users = models.Users{Username: username, Password: password, Email: email}

	err = db.CreateUser(user)
	if err != nil {
		if strings.Contains(err.Error(), "invalid request format") {
			http.Error(w, err.Error(), http.StatusBadRequest)
		} else if strings.Contains(err.Error(), "invalid email format") || strings.Contains(err.Error(), "email already exists") {
			http.Error(w, err.Error(), http.StatusConflict)
		} else {
			http.Error(w, "internal server error: "+err.Error(), http.StatusInternalServerError)
		}
		return
	}

	var response map[string]string = map[string]string{
		"message": fmt.Sprintf("New user created: %v", user.Username),
	}

	jsonRes, err := json.Marshal(response)
	if err != nil {
		http.Error(w, "failed to create response", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated) // indicate successful creation
	w.Write(jsonRes)
}
