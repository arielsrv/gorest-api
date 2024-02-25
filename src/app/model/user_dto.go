package model

import "time"

type UserDTO struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	Email  string `json:"email"`
	Gender string `json:"gender"`
	Status string `json:"status"`

	Posts []PostDTO `json:"posts"`
	Todos []TodoDTO `json:"todos"`
}

type PostDTO struct {
	ID     int    `json:"id"`
	UserID int    `json:"-"`
	Title  string `json:"title"`
	Body   string `json:"body"`

	Comments []CommentDTO `json:"comments"`
}

type TodoDTO struct {
	ID     int       `json:"id"`
	UserID int       `json:"_"`
	Title  string    `json:"title"`
	DueOn  time.Time `json:"due_on"`
	Status string    `json:"status"`
}

type CommentDTO struct {
	ID     int    `json:"id"`
	PostID int    `json:"_"`
	Name   string `json:"name"`
	Email  string `json:"email"`
	Body   string `json:"body"`
}
