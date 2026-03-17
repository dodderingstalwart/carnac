package main

import (
	"encoding/json"
	"net/http"
)

// Joke represents a carnac joke
type Joke struct {
	ID int `json:"id"`
	Answer string `json:"answer"`
	Question string `json:"question"`
}

// Insult represents a carnac insult
type Insult struct {
	ID int `json:"id"`
	Insult string `json:"insult"`
}

// APIServer for handing HTTP requests
type APIServer struct {
	db *sql.DB
}

// ErrorResponse represents a response to an error
type ErrorResponse struct {
	Error string `json:"error"`
}

// NewAPIServer creates a new API server
func NewAPIServer(db *sql.DB) *APIServer {
	return &APIServer{db: db}
}

