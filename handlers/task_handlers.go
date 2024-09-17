package main

import (
	"encoding/json"
	"io"
	"net/http"
)

func CreateTask(w http.ResponseWriter, r *http.Request) {
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
	id, name, done, created_at, updated_at := data["id"], data["name"], data["done"], data["created_at"], data["updated_at"]
	var new_task Task := Task{ID: id, Name: name, Done: done, CreatedAt: created_at, UpdatedAt: updated_at}
	

}

func UpdateTask(w http.ResponseWriter, r *http.Request) {

}

func GetTask(w http.ResponseWriter, r *http.Request) {

}

func DeleteTask(w http.ResponseWriter, r *http.Request) {

}
