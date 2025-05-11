package main

import (
	"log"
)

func main() {
	agent := NewAgent("http://localhost:8080", "http://localhost:8081", "agent-secret-token")
	log.Println("Agent started")
	agent.Run()
}
