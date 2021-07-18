package model

import (
	"database/sql"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type sqliteHandler struct {
	db *sql.DB
}

func (s *sqliteHandler) GetTodos(sessionId string) []*Todo {
	todos := []*Todo{}
	rows, err := s.db.Query(
		`SELECT id, name, completed, createdAt
		   FROM todos
		  WHERE sessionId = ?
	`, sessionId)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	for rows.Next() {
		var todo Todo
		rows.Scan(&todo.ID, &todo.Name, &todo.Completed, &todo.CreatedAt)
		todos = append(todos, &todo)
	}

	return todos
}

func (s *sqliteHandler) AddTodo(sessionId, name string) *Todo {
	statement, err := s.db.Prepare(
		`INSERT INTO todos ( sessionId
			               , name
			               , completed
						   , createdAt
		) VALUES (?, ?, ?, datetime('now'))
	`)
	if err != nil {
		panic(err)
	}
	defer statement.Close()

	result, err := statement.Exec(sessionId, name, false)
	if err != nil {
		panic(err)
	}
	// 발급된 아이디를 가져온다.
	id, err := result.LastInsertId()
	if err != nil {
		panic(err)
	}

	var todo Todo
	todo.ID = int(id)
	todo.Name = name
	todo.Completed = false
	todo.CreatedAt = time.Now()
	return &todo
}

func (s *sqliteHandler) RemoveTodo(sessionId string, id int) bool {
	statement, err := s.db.Prepare(
		`DELETE
		   FROM todos
		  WHERE id = ?
		    AND sessionId = ?
	`)
	if err != nil {
		panic(err)
	}
	defer statement.Close()

	result, err := statement.Exec(id, sessionId)
	if err != nil {
		panic(err)
	}

	cnt, err := result.RowsAffected()
	if err != nil {
		panic(err)
	}

	return cnt == 1
}
func (s *sqliteHandler) PatchTodo(sessionId string, id int, completed bool) bool {
	statement, err := s.db.Prepare(`
		UPDATE todos
		   SET completed = ?
		 WHERE id = ?
		   AND sessionId = ?
	`)
	if err != nil {
		panic(err)
	}
	defer statement.Close()

	result, err := statement.Exec(completed, id, sessionId)
	if err != nil {
		panic(err)
	}
	cnt, err := result.RowsAffected()
	if err != nil {
		panic(err)
	}

	return cnt == 1
}

func (s *sqliteHandler) Close() {
	s.db.Close()
}

func newSqliteHandler(filepath string) DBHandler {
	database, err := sql.Open("sqlite3", filepath)
	if err != nil {
		panic(err)
	}

	statement, err := database.Prepare(
		`CREATE TABLE IF NOT EXISTS todos (
			id        INTEGER PRIMARY KEY AUTOINCREMENT,
			sessionId STRING,
			name      TEXT,
			completed BOOLEAN,
			createdAt DATETIME
	);
	CREATE INDEX IF NOT EXISTS sessionIdIndexOnTodos On todos (sessionId ASC);
	`)
	if err != nil {
		panic(err)
	}
	statement.Exec()

	return &sqliteHandler{db: database}
}
