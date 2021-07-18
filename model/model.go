package model

import (
	"time"
)

type Todo struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Completed bool      `json:"completed"`
	CreatedAt time.Time `json:"created_at"`
}

type DBHandler interface {
	GetTodos(sessionId string) []*Todo
	AddTodo(sessionId, name string) *Todo
	RemoveTodo(sessionId string, id int) bool
	PatchTodo(sessionId string, id int, completed bool) bool
	Close()
}

var handler DBHandler

func NewDBHandler(filepath string) DBHandler {
	// return newMemoryHandler()
	return newSqliteHandler(filepath)
}
