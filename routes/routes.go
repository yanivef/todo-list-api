package routes

import (
	"net/http"
	"task-manager-api/handlers"
	"task-manager-api/utils"

	"github.com/gorilla/mux"
)

func NewRouter() *mux.Router {

	r := mux.NewRouter()
	// r.HandleFunc("/tasks", handlers.HandleTasks).Methods("GET", "POST", "DELETE", "PUT")
	r.HandleFunc("/users", handlers.CreateUser).Methods("POST")
	r.HandleFunc("/login", handlers.Login).Methods("POST")
	r.Handle("/tasks", utils.JWTAuthMiddleware(http.HandlerFunc(handlers.HandleTasks))).Methods("GET", "POST", "DELETE", "PUT")

	return r

}
