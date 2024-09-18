package main

import (
	"log"
	"net/http"
	"task-manager-api/db"
	_ "task-manager-api/handlers"
	"task-manager-api/routes"

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

	r := routes.NewRouter()
	log.Fatal(http.ListenAndServe(":5000", r))
}
