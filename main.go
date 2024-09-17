package main

import (
	"log"
	"net/http"
	"task-manager-api/db"
	"task-manager-api/handlers"
	_ "task-manager-api/handlers"

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

	http.HandleFunc("/tasks", handlers.CreateTask)
	//http.HandleFunc("/tasks", handlers.GetTask)

	log.Fatal(http.ListenAndServe(":5000", nil))
}
