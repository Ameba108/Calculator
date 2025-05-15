package main

import (
	mod "calc/internal/models"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/Knetic/govaluate"
)

var (
	expressions = make(map[string]*mod.Expression) // Хранилище выражений
	tasks       = make(map[string]*mod.Task)       // Хранилище задач
	mutex       = &sync.Mutex{}
)

type Orchestrator struct {
	db *DB
}

func (o *Orchestrator) CalculateHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(int)

	var req struct {
		Expression string `json:"expression"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if !isValidMathExpression(req.Expression) {
		http.Error(w, "Invalid expression: only numbers, +, -, *, /, (), and spaces are allowed", http.StatusBadRequest)
		return
	}
	if containsDivisionByZero(req.Expression) {
		http.Error(w, "Invalid expression: division by zero is not allowed", http.StatusBadRequest)
		return
	}

	expr, err := govaluate.NewEvaluableExpression(req.Expression)
	if err != nil {
		http.Error(w, "Invalid expression syntax", http.StatusBadRequest)
		return
	}

	id := fmt.Sprintf("%d", time.Now().UnixNano())
	expression := &mod.Expression{
		ID:     id,
		UserID: userID,
		Expr:   req.Expression,
		Status: "pending",
	}

	if err := o.db.SaveExpression(expression); err != nil {
		http.Error(w, "Failed to save expression", http.StatusInternalServerError)
		return
	}

	go o.evaluateExpression(id, expr, userID)

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"id": id})
}

func isValidMathExpression(expr string) bool {
	allowedChars := `0123456789+-*/(). `
	for _, char := range expr {
		if !strings.ContainsRune(allowedChars, char) {
			return false
		}
	}
	return true
}

func containsDivisionByZero(expr string) bool {
	return strings.Contains(expr, "/0") || strings.Contains(expr, "/ 0")
}

func (o *Orchestrator) evaluateExpression(id string, expr *govaluate.EvaluableExpression, userID int) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Expression evaluation failed (ID: %s): %v", id, r)
			o.db.Exec("UPDATE expressions SET status = ? WHERE id = ? AND user_id = ?", "error", id, userID)
		}
	}()

	result, err := expr.Evaluate(nil)
	if err != nil {
		log.Printf("Expression evaluation error (ID: %s): %v", id, err)
		o.db.Exec("UPDATE expressions SET status = ? WHERE id = ? AND user_id = ?", "error", id, userID)
		return
	}

	if _, ok := result.(float64); !ok {
		log.Printf("Expression returned non-float result (ID: %s)", id)
		o.db.Exec("UPDATE expressions SET status = ? WHERE id = ? AND user_id = ?", "error", id, userID)
		return
	}

	o.db.Exec("UPDATE expressions SET status = ?, result = ? WHERE id = ? AND user_id = ?",
		"done", result.(float64), id, userID)
}

func (o *Orchestrator) ExpressionsHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(int)

	expressions, err := o.db.GetExpressions(userID)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	safeExpressions := make([]*mod.Expression, 0, len(expressions))
	for _, expr := range expressions {
		if expr.Status == "error" {
			safeExpressions = append(safeExpressions, expr)
			continue
		}

		if _, err := govaluate.NewEvaluableExpression(expr.Expr); err != nil {
			expr.Status = "error"
		}
		safeExpressions = append(safeExpressions, expr)
	}

	json.NewEncoder(w).Encode(safeExpressions)
}

func (o *Orchestrator) ExpressionHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Path[len("/api/v1/expressions/"):]
	userID := r.Context().Value("user_id").(int)

	expr, err := o.db.GetExpressionByID(userID, id)
	if err != nil {
		http.Error(w, "Expression not found", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(expr)
}

func (o *Orchestrator) TaskHandler(w http.ResponseWriter, r *http.Request) {
	mutex.Lock()
	defer mutex.Unlock()

	switch r.Method {
	case http.MethodGet:
		for _, task := range tasks {
			if task.Arg1 != 0 && task.Arg2 != 0 && task.Operation != "" {
				json.NewEncoder(w).Encode(map[string]interface{}{
					"task": task,
				})
				return
			}
		}
		http.Error(w, "No tasks available", http.StatusNotFound)

	case http.MethodPost:
		var req struct {
			ID     string  `json:"id"`
			Result float64 `json:"result"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusUnprocessableEntity)
			return
		}

		task, exists := tasks[req.ID]
		if !exists {
			http.Error(w, "Task not found", http.StatusNotFound)
			return
		}

		task.Arg1 = req.Result
		task.Arg2 = 0
		task.Operation = ""

		for _, expr := range expressions {
			if expr.ID == task.ID {
				expr.Status = "done"
				expr.Result = req.Result
				break
			}
		}

		w.WriteHeader(http.StatusOK)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}
