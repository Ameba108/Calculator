package main

import (
	mod "calc/internal/models"
	"database/sql"
	"fmt"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

type DB struct {
	*sql.DB
}

func InitDB() (*DB, error) {
	db, err := sql.Open("sqlite3", "./db/calculator.db")
	if err != nil {
		log.Fatal(err)
	}
	_, err = db.Exec(`
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		login TEXT UNIQUE NOT NULL,
		password TEXT NOT NULL
	);

	CREATE TABLE IF NOT EXISTS expressions (
		id TEXT PRIMARY KEY,
		user_id INTEGER NOT NULL,
		expr TEXT NOT NULL,
		status TEXT NOT NULL,
		result REAL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY(user_id) REFERENCES users(id)
	);

	CREATE TABLE IF NOT EXISTS tasks (
		id TEXT PRIMARY KEY,
		user_id INTEGER NOT NULL,
		arg1 REAL,
		arg2 REAL,
		operation TEXT,
		operation_time INTEGER,
		FOREIGN KEY(user_id) REFERENCES users(id)
	);`)
	if err != nil {
		log.Fatal(err)
	}
	return &DB{db}, nil
}

func (db *DB) GetExpressionByID(userID int, id string) (*mod.Expression, error) {
	var expr mod.Expression
	row := db.QueryRow(
		"SELECT id, user_id, expr, status, result FROM expressions WHERE id = ? AND user_id = ?",
		id, userID,
	)

	err := row.Scan(&expr.ID, &expr.UserID, &expr.Expr, &expr.Status, &expr.Result)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("expression not found")
		}
		return nil, err
	}
	return &expr, nil
}

func (db *DB) SaveExpression(expr *mod.Expression) error {
	_, err := db.Exec(
		"INSERT INTO expressions (id, user_id, expr, status) VALUES (?, ?, ?, ?)",
		expr.ID, expr.UserID, expr.Expr, expr.Status,
	)
	return err
}

func (db *DB) GetExpressions(userID int) ([]*mod.Expression, error) {
	rows, err := db.Query(`
    SELECT id, user_id, expr, status, result FROM expressions WHERE user_id = ?`,
		userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var expressions []*mod.Expression
	for rows.Next() {
		var expr mod.Expression
		err := rows.Scan(
			&expr.ID,
			&expr.UserID,
			&expr.Expr,
			&expr.Status,
			&expr.Result)
		if err != nil {
			return nil, err
		}
		expr.UserID = userID
		expressions = append(expressions, &expr)
	}
	return expressions, nil
}
