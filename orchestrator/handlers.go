package main

import (
	mod "calc/internal/models"
	"encoding/json"
	"fmt"
	"net/http"
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

	expr, err := govaluate.NewEvaluableExpression(req.Expression)
	if err != nil {
		http.Error(w, "Invalid expression", http.StatusBadRequest)
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

func (o *Orchestrator) evaluateExpression(id string, expr *govaluate.EvaluableExpression, userID int) {
	result, err := expr.Evaluate(nil)
	if err != nil {
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

	json.NewEncoder(w).Encode(expressions)
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
