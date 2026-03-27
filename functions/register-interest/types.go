package main

// RegisterRequest is the expected form body for the register interest endpoint.
type RegisterRequest struct {
	Email string `json:"email"`
}
