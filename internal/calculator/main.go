package main

import (
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/calculate", calculateHandler)

	log.Println("Calculator is running on :8081")
	log.Fatal(http.ListenAndServe(":8081", nil))
}
