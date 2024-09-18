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
		os.Getenv("DB_USERNAME"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"))

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

// close database connection
func Close() error {
	if DB != nil {
		return DB.Close() // returns error if couldn't close DB connection
	}
	return nil // no need to close, DB is nil
}

// check if tasks exists in the database
func TaskExists(id int) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM tasks WHERE id=$1)` // $1 is a placeholder
	err := DB.QueryRow(query, id).Scan(&exists)
	return exists, err
}

// insert new task to the database
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

// gets all tasks from database
func GetAllTasks() ([]task.Task, error) {
	var err error
	var tasks []task.Task
	query := `SELECT id, name, done, created_at, updated_at FROM tasks`

	rows, err := DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close() // close database cursor

	for rows.Next() {
		var t task.Task
		err = rows.Scan(&t.ID, &t.Name, &t.Done, &t.CreatedAt, &t.UpdatedAt)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, t)
	}
	return tasks, nil
}
