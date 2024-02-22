package model

import "time"

type TodoResponse struct {
	ID     int       `json:"id"`
	UserID int       `json:"user_id"`
	Title  string    `json:"title"`
	DueOn  time.Time `json:"due_on"`
	Status string    `json:"status"`
}
