package db

import (
	task "task-manager-api/models"

	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
)

var DB *sql.DB

func Init() error {
	connStr := fmt.Sprintf("user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("DB_NAME"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_USERNAME"))

	var err error
	DB, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Printf("Error opening database connection: %s", err)
		return err
	}

	err = DB.Ping()
	if err != nil {
		log.Printf("Error pinging database: %s", err)
		return err
	}
	log.Println("Successfully connected to the database")
	return nil
}

func Close() error {
	if DB != nil {
		return DB.Close() // returns error if couldn't close DB connection
	}
	return nil // no need to close, DB is nil
}

func TaskExists(id int) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM tasks WHERE id=$1)` // $1 is a placeholder
	err := DB.QueryRow(query, id).Scan(&exists)
	return exists, err
}

func InsertTask(task task.Task) error {
	var err error
	id, name, done := task.ID, task.Name, task.Done

	query := `INSERT INTO tasks (id, name, done) VALUES($1, $2, $3)`

	_, err = DB.Exec(query, id, name, done)
	if err != nil {
		log.Printf("Error inserting task: %s", err)
		return err
	} else {
		log.Printf("New task inserted to the DB, task details: %v", task)
	}

	return nil
}
