package models

type Users struct {
	Username string `json: "username"`
	Password string `json: "-"`
	Email    string `json: "email"`
}
