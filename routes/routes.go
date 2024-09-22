package routes

import (
	"task-manager-api/handlers"

	"github.com/gorilla/mux"
)

func NewRouter() *mux.Router {

	r := mux.NewRouter()

	r.HandleFunc("/tasks", handlers.HandelTasks).Methods("GET", "POST")
	r.HandleFunc("/tasks/{id:[0-9]+}", handlers.HandleTaskByID).Methods("GET", "DELETE", "PUT")
	r.HandleFunc("/users", handlers.CreateUser).Methods("POST")
	r.HandleFunc("/login", handlers.Login).Methods("POST")

	return r

}
