package main

import (
	"log"
	"task-manager-api/db"

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
}
