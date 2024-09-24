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
	"task-manager-api/utils"
)

func HandleTasks(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value("claims").(*utils.CustomClaims)
	if !ok {
		http.Error(w, "Could not extract user claims", http.StatusInternalServerError)
		return
	}
	userEmail := claims.Email

	switch r.Method {
	case http.MethodGet:
		handleGetTasks(w, r, userEmail)
	case http.MethodPost:
		CreateTask(w, r, userEmail)
	case http.MethodDelete:
		DeleteTaskByID(w, r, userEmail)
	case http.MethodPut:
		UpdateTaskByID(w, r, userEmail)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// function to create task
func CreateTask(w http.ResponseWriter, r *http.Request, userEmail string) {
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
	name, nameOK := data["name"].(string)
	desc, descOK := data["description"].(string)
	status, statusOK := data["status"].(bool)

	if name == "" || !nameOK {
		http.Error(w, "Task name required", http.StatusBadRequest)
		return
	}

	if !descOK {
		desc = ""
	}

	if !statusOK {
		status = false
	}

	// create new task with the given details, set created_at, update_at to current time as default creation
	new_task := models.Task{
		Name:        name,
		Description: desc,
		Status:      status,
		OwnerEmail:  userEmail,
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
func UpdateTaskByID(w http.ResponseWriter, r *http.Request, userEmail string) {
	if r.Method != http.MethodPut {
		http.Error(w, "Only PUT method is allowed", http.StatusMethodNotAllowed)
		return
	}
	ids, idOK := r.URL.Query()["id"]
	if !idOK || len(ids) == 0 {
		http.Error(w, "Task ID is required", http.StatusBadRequest)
		return
	}
	id, err := strconv.Atoi(ids[0]) // convert first ID from string to int
	if err != nil {
		http.Error(w, "Invalid ID format", http.StatusBadRequest)
		return
	}

	// check if task exists
	exists, err := db.TaskExists(id, userEmail)
	if err != nil {
		http.Error(w, "Failed to check task existence: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if !exists {
		http.Error(w, "Task does not exist, ID: "+ids[0], http.StatusBadRequest)
		return
	}

	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Couldn't read body", http.StatusBadRequest)
		return
	}

	// unmarshal JSON body to a map
	var data map[string]interface{}
	err = json.Unmarshal(body, &data)
	if err != nil {
		http.Error(w, "Couldn't parse json", http.StatusBadRequest)
		return
	}

	// check if body contains fields to update
	if len(data) == 0 {
		http.Error(w, "Request body cannot be empty", http.StatusBadRequest)
		return
	}

	updates := make(map[string]interface{})

	if name, ok := data["name"].(string); ok && name != "" {
		updates["name"] = name
	}

	if status, ok := data["status"].(bool); ok {
		updates["status"] = status
	}

	updates["updated_at"] = time.Now()

	err = db.UpdateTask(id, userEmail, updates)
	if err != nil {
		http.Error(w, "Couldn't read task: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf("Task with ID: %d successfully updated", id)))
}

func GetTasks(w http.ResponseWriter, r *http.Request, userEmail string) {
	// GET REQUEST -> return list of all tasks from database
	if r.Method != http.MethodGet {
		http.Error(w, "Only GET method is allowed", http.StatusMethodNotAllowed)
		return
	}
	tasks, err := db.GetAllTasks(userEmail)
	if err != nil {
		http.Error(w, "Couldn't fetch tasks from database", http.StatusInternalServerError)
		return
	}

	if tasks == nil {
		tasks = []models.Task{} // in order to return response: [] instead of null -> ensures tasks is empty slice and not nil
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
func DeleteTaskByID(w http.ResponseWriter, r *http.Request, userEmail string) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Only DELETE method is allowed", http.StatusMethodNotAllowed)
		return
	}
	// retrieve the ID from the URL query parameter
	ids, idsOK := r.URL.Query()["id"]
	if !idsOK || len(ids) == 0 {
		http.Error(w, "No ID parameter was found", http.StatusBadRequest)
		return
	}
	// convert the first ID to an integer
	id, err := strconv.Atoi(ids[0])
	if err != nil {
		http.Error(w, "Invalid ID format", http.StatusBadRequest)
		return
	}
	// delete task from the database
	err = db.DeleteTask(id, userEmail)
	if err != nil {
		res := fmt.Sprintf("Error deleting task from database, Error: %s", err.Error())
		http.Error(w, res, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// handle get request to tasks URL, if ID(one or more) provided then trigger 'GetMultipleTasksByID', otherwise, trigger 'GetTasks'
func handleGetTasks(w http.ResponseWriter, r *http.Request, userEmail string) {
	_, idOK := r.URL.Query()["id"]

	// if id parameter exists, get task by id, otherwise retrieve all tasks

	if idOK {
		GetMultipleTasksByID(w, r, userEmail)
	} else {
		GetTasks(w, r, userEmail)
	}
}

// handles multiple tasks ID requests, passing IDs to 'GetMultipleTasks' function
func GetMultipleTasksByID(w http.ResponseWriter, r *http.Request, userEmail string) {
	idsStr, _ := r.URL.Query()["id"]

	var ids []int
	for _, val := range idsStr {
		id, err := strconv.Atoi(val)
		if err != nil {
			http.Error(w, "Invalid ID format", http.StatusBadRequest)
			return
		}
		ids = append(ids, id)
	}
	mulTasks, err := db.GetMultipleTasks(ids, userEmail)

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

func Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	defer r.Body.Close()

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed parse body", http.StatusBadRequest)
		return
	}

	var cred map[string]string

	err = json.Unmarshal(body, &cred)
	if err != nil {
		http.Error(w, "failed parse JSON", http.StatusInternalServerError)
		return
	}

	email, emailOK := cred["email"]
	password, passwordOK := cred["password"]

	if !emailOK || !passwordOK {
		http.Error(w, "invalid request format", http.StatusBadRequest)
		return
	}

	user, err := db.GetUserByEmail(email)
	hashedPassword := user.Password
	credOK := utils.CheckPassword(hashedPassword, password)

	if err != nil || !credOK {
		if strings.Contains(err.Error(), "not found in database") {
			http.Error(w, "invalid email or password", http.StatusUnauthorized)
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// generate JWT token
	token, err := utils.GenerateToken(user)
	if err != nil {
		http.Error(w, "failed to generate token", http.StatusInternalServerError)
		return
	}

	response := map[string]string{"token": token}
	jsonRes, err := json.Marshal(response)
	if err != nil {
		http.Error(w, "failed to create response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonRes)
}
