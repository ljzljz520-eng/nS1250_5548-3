package account

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

type Service struct {
	repository Repository
	validate   *validator.Validate
}

func NewService(repository Repository) *Service {
	return &Service{
		repository: repository,
		validate:   validator.New(validator.WithRequiredStructEnabled()),
	}
}

func (s *Service) Create(ctx context.Context, worker Worker) (*Worker, error) {
	normalizeWorker(&worker)
	if err := s.validate.Struct(worker); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidWorker, err)
	}
	if err := s.repository.Create(ctx, &worker); err != nil {
		return nil, err
	}
	return &worker, nil
}

func (s *Service) Update(ctx context.Context, employeeID string, update UpdateWorker) (*Worker, error) {
	worker := Worker{
		Name:       update.Name,
		EmployeeID: strings.TrimSpace(employeeID),
		Phone:      update.Phone,
		Email:      update.Email,
		Team:       update.Team,
		Status:     update.Status,
	}
	normalizeWorker(&worker)
	if err := s.validate.Struct(worker); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidWorker, err)
	}
	if err := s.repository.Update(ctx, &worker); err != nil {
		return nil, err
	}
	return &worker, nil
}

func (s *Service) Delete(ctx context.Context, employeeID string, confirmed bool) error {
	if !confirmed {
		return ErrConfirmationRequired
	}
	return s.repository.Delete(ctx, strings.TrimSpace(employeeID))
}

func (s *Service) Get(ctx context.Context, employeeID string) (*Worker, error) {
	worker, err := s.repository.FindByEmployeeID(ctx, strings.TrimSpace(employeeID))
	if err != nil {
		return nil, err
	}
	worker.Team = strings.TrimSpace(worker.Team)
	return worker, nil
}

func (s *Service) List(ctx context.Context) ([]Worker, error) {
	return s.repository.List(ctx)
}

func normalizeWorker(worker *Worker) {
	worker.Name = strings.TrimSpace(worker.Name)
	worker.EmployeeID = strings.TrimSpace(worker.EmployeeID)
	worker.Phone = strings.TrimSpace(worker.Phone)
	worker.Email = strings.ToLower(strings.TrimSpace(worker.Email))
	worker.Team = strings.TrimSpace(worker.Team)
}
