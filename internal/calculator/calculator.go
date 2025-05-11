package main

import (
	"encoding/json"
	"net/http"
	"time"
)

type CalculationRequest struct {
	Arg1      float64 `json:"arg1"`
	Arg2      float64 `json:"arg2"`
	Operation string  `json:"operation"`
}

func calculateHandler(w http.ResponseWriter, r *http.Request) {
	var req CalculationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	var result float64
	switch req.Operation {
	case "+":
		result = req.Arg1 + req.Arg2
	case "-":
		result = req.Arg1 - req.Arg2
	case "*":
		result = req.Arg1 * req.Arg2
	case "/":
		result = req.Arg1 / req.Arg2
	default:
		http.Error(w, "Invalid operation", http.StatusBadRequest)
		return
	}
	time.Sleep(time.Duration(1 * time.Second))
	json.NewEncoder(w).Encode(result)
}
