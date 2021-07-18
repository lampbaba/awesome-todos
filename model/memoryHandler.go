package model

import (
	"time"
)

type memoryHandler struct {
	todoMap map[int]*Todo
}

func newMemoryHandler() DBHandler {
	m := &memoryHandler{}
	m.todoMap = make(map[int]*Todo)
	return m
}

func (m *memoryHandler) GetTodos(sessionId string) []*Todo {
	list := []*Todo{}
	for _, value := range m.todoMap {
		list = append(list, value)
	}
	return list
}

func (m *memoryHandler) AddTodo(sessionId, name string) *Todo {
	id := len(m.todoMap) + 1
	m.todoMap[id] = &Todo{ID: id, Name: name, Completed: false, CreatedAt: time.Now()}
	return m.todoMap[id]
}

func (m *memoryHandler) RemoveTodo(sessionId string, id int) bool {
	_, ok := m.todoMap[id]
	if ok {
		delete(m.todoMap, id)
	}
	return ok
}

func (m *memoryHandler) PatchTodo(sessionId string, id int, completed bool) bool {
	todo, ok := m.todoMap[id]
	if ok {
		todo.Completed = completed
	}
	return ok
}

func (m *memoryHandler) Close() {

}
