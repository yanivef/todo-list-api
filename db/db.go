package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"sync"
	"task-manager-api/models"
	task "task-manager-api/models"
	"task-manager-api/utils"
	_ "task-manager-api/utils"

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

// get task from database by given ID
func GetTask(id int) (task.Task, error) {
	var err error
	var task task.Task
	query := `SELECT id, name, done, created_at, updated_at FROM tasks WHERE id = $1`

	row := DB.QueryRow(query, id)

	err = row.Scan(&task.ID, &task.Name, &task.Done, &task.CreatedAt, &task.UpdatedAt)
	if err != nil {
		// check if the error suggests that no row was found with the given ID
		if err == sql.ErrNoRows {
			return task, fmt.Errorf("task with ID %d was not found", id)
		}
		return task, err
	}

	return task, nil

}

func DeleteTask(id int) error {
	var err error
	exists, err := TaskExists(id)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("task ID: %v not found", id)
	}
	query := `DELETE FROM tasks WHERE id=$1`
	_, err = DB.Exec(query, id)
	if err != nil {
		return err
	}
	log.Printf("Task %v deleted successfully", id)
	return nil
}

func UpdateTask(id int, updates map[string]interface{}) error {
	if exists, _ := TaskExists(id); !exists {
		return fmt.Errorf("task ID %v was not found", id)
	}
	setClause := ""
	args := []interface{}{} // init empty slice

	i := 1
	for key, val := range updates {
		if setClause != "" {
			setClause += ", "
		}
		setClause += fmt.Sprintf("%s = $%d", key, i)
		args = append(args, val)
		i++
	}

	args = append(args, id)
	query := fmt.Sprintf(`UPDATE tasks SET %s WHERE id =$%d`, setClause, i)
	_, err := DB.Exec(query, args...)
	if err != nil {
		return err
	}
	return nil

}

func GetMultipleTasks(ids []int) ([]task.Task, error) {
	var wg sync.WaitGroup
	errsChan := make(chan error, len(ids))
	tasksChan := make(chan task.Task, len(ids))
	tasks := make([]task.Task, 0, len(ids))

	for _, id := range ids {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			task, err := GetTask(id)
			if err != nil {
				errsChan <- err
				return
			}
			tasksChan <- task
		}(id)
	}

	go func() {
		wg.Wait()
		close(tasksChan)
		close(errsChan)
	}()

	for {
		select {
		case task, ok := <-tasksChan:
			if ok {
				tasks = append(tasks, task)
			} else {
				return tasks, nil
			}
		case err, ok := <-errsChan:
			if ok {
				return nil, err
			}
		}
	}
}

// USER FUNC

func CreateUser(user models.Users) error {
	username, password, email := user.Username, user.Password, user.Email

	if username == "" || password == "" || email == "" {
		mockReq := `example of request:
		{
		"username":"name",
		"password":"pass",
		"email":"example@example.com"
		}`

		return fmt.Errorf("invalid request format\n%v", mockReq)
	}
	if !utils.IsValidEmail(email) {
		return fmt.Errorf("invalid email format")
	}

	exists, err := utils.IsEmailExists(email, DB)
	if err != nil {
		return fmt.Errorf("error query db for user email")
	}
	if exists {
		return fmt.Errorf("email already exists")
	}

	hashPassword, err := utils.HashPassword(password)
	if err != nil {
		log.Printf("failed to hash password")
		return fmt.Errorf("failed to hash password")
	}

	query := `INSERT INTO users (username, pass, email) VALUES($1, $2, $3)`
	_, err = DB.Exec(query, username, hashPassword, email)
	if err != nil {
		log.Printf("error creating user: %s", err)
		return fmt.Errorf("error creating user: %s", err)
	}
	log.Printf("new user created: %s, %s", user.Username, user.Email)
	return nil
}

func GetUserByEmail(email string) (models.Users, error) {
	var err error
	var user models.Users
	query := `SELECT username, pass, email FROM users WHERE email = $1`

	row := DB.QueryRow(query, email)

	err = row.Scan(&user.Username, &user.Password, &user.Email)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("User email: %s not found in database", email)
			return user, fmt.Errorf("user email: %s not found in database", email)
		}
		log.Printf("Database error: %s", err.Error())
		return user, fmt.Errorf("database error: %s", err.Error())
	}
	return user, nil
}
