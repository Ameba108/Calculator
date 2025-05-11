package main

import (
	"bytes"
	mod "calc/internal/models"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

type Agent struct {
	orchestratorURL string
	calculatorURL   string
	token           string
}

func NewAgent(orchestratorURL, calculatorURL, token string) *Agent {
	return &Agent{
		orchestratorURL: orchestratorURL,
		calculatorURL:   calculatorURL,
		token:           token,
	}
}

func (a *Agent) fetchTask() (*mod.Task, error) {
	req, err := http.NewRequest("GET", a.orchestratorURL+"/internal/task", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", a.token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch task: %s", resp.Status)
	}

	var task mod.Task
	if err := json.NewDecoder(resp.Body).Decode(&task); err != nil {
		return nil, err
	}

	return &task, nil
}

func (a *Agent) executeTask(task *mod.Task) float64 {
	reqBody, _ := json.Marshal(map[string]interface{}{
		"arg1":      task.Arg1,
		"arg2":      task.Arg2,
		"operation": task.Operation,
	})

	resp, err := http.Post(a.calculatorURL+"/calculate", "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		log.Println("Error sending task to calculator:", err)
		return 0
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Println("Error:", resp.Status)
		return 0
	}

	var result float64
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		log.Println("Error decoding calculator response:", err)
		return 0
	}

	return result
}

func (a *Agent) sendResult(taskID string, result float64) error {
	reqBody, _ := json.Marshal(map[string]interface{}{
		"id":     taskID,
		"result": result,
	})

	req, err := http.NewRequest("POST", a.orchestratorURL+"/internal/task", bytes.NewBuffer(reqBody))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", a.token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to send result: %s", resp.Status)
	}

	return nil
}

func (a *Agent) Run() {
	for {
		task, err := a.fetchTask()
		if err != nil {
			log.Println("Error fetching task:", err)
			time.Sleep(5 * time.Second)
			continue
		}

		result := a.executeTask(task)
		if err := a.sendResult(task.ID, result); err != nil {
			log.Println("Error sending result:", err)
		}
	}
}
