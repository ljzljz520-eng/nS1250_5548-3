package account

import "errors"

type Status string

const (
	StatusActive   Status = "active"
	StatusInactive Status = "inactive"
)

var (
	ErrAlreadyExists        = errors.New("worker already exists")
	ErrNotFound             = errors.New("worker not found")
	ErrConfirmationRequired = errors.New("deletion confirmation required")
	ErrInvalidWorker        = errors.New("invalid worker")
)

type Worker struct {
	Name       string `json:"name" validate:"required,min=2,max=80"`
	EmployeeID string `json:"employee_id" validate:"required,alphanumunicode,min=3,max=32"`
	Phone      string `json:"phone" validate:"required,e164"`
	Email      string `json:"email" validate:"required,email,max=254"`
	Team       string `json:"team" validate:"required,min=2,max=80"`
	Status     Status `json:"status" validate:"required,oneof=active inactive"`
}

type UpdateWorker struct {
	Name   string `json:"name" validate:"required,min=2,max=80"`
	Phone  string `json:"phone" validate:"required,e164"`
	Email  string `json:"email" validate:"required,email,max=254"`
	Team   string `json:"team" validate:"required,min=2,max=80"`
	Status Status `json:"status" validate:"required,oneof=active inactive"`
}
