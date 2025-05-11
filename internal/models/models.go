package models

import "time"

type User struct {
	ID       int
	Login    string
	Password string
}

type Expression struct {
	ID     string
	UserID int
	Expr   string
	Status string
	Result float64
}

type Task struct {
	ID            string
	UserID        int
	Arg1          float64
	Arg2          float64
	Operation     string
	OperationTime time.Duration
}
