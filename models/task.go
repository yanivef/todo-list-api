package models

import "time"

type Task struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description`
	Status      bool      `json:"status"`
	OwnerEmail  string    `json:"owner_email`
	Owner       *Users    `json:"owner,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
