package main

import (
	"log"
	"net/http"
	"task-manager-api/db"
	"task-manager-api/handlers"
	_ "task-manager-api/handlers"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

func main() {
	var err error
	err = godotenv.Load(".env")
	if err != nil {
		log.Fatal(err)
	}
	err = db.Init()
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("Error closing database %s", err)
		}
	}()

	r := mux.NewRouter()

	r.HandleFunc("/tasks", handlers.HandelTasks).Methods("GET", "POST")
	// http.HandleFunc("/tasks", handlers.HandelTasks)
	r.HandleFunc("/tasks/{id:[0-9]+}", handlers.HandleTaskByID).Methods("GET", "DELETE", "PUT")
	log.Fatal(http.ListenAndServe(":5000", r))
}
