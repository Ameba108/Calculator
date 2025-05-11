package main

import (
	"log"
	"net/http"
)

func main() {
	db, err := InitDB()
	if err != nil {
		log.Fatal("Failed to initialize database:", err)
	}
	defer db.Close()

	authHandler := &AuthHandler{db: db}
	orchestrator := &Orchestrator{db: db}

	http.HandleFunc("/api/v1/register", authHandler.Register)
	http.HandleFunc("/api/v1/login", authHandler.Login)
	http.HandleFunc("/api/v1/calculate", AuthMiddleware(db, orchestrator.CalculateHandler))
	http.HandleFunc("/api/v1/expressions", AuthMiddleware(db, orchestrator.ExpressionsHandler))
	http.HandleFunc("/api/v1/expressions/", AuthMiddleware(db, orchestrator.ExpressionHandler))
	http.HandleFunc("/internal/task", AuthMiddleware(db, orchestrator.TaskHandler))

	log.Println("Orchestrator is running on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
