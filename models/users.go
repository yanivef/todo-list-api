package models

type Users struct {
	ID       int    `json: "id"`
	Username string `json: "username"`
	Password string `json: "-"`
	Email    string `json: "email"`
}
