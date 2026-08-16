package account

import (
	"context"
	"sort"
	"sync"
)

type Repository interface {
	Create(context.Context, *Worker) error
	Update(context.Context, *Worker) error
	Delete(context.Context, string) error
	FindByEmployeeID(context.Context, string) (*Worker, error)
	List(context.Context) ([]Worker, error)
}

type MemoryRepository struct {
	mu      sync.RWMutex
	workers map[string]Worker
}

func NewMemoryRepository(initial []Worker) *MemoryRepository {
	workers := make(map[string]Worker, len(initial))
	for _, worker := range initial {
		workers[worker.EmployeeID] = worker
	}
	return &MemoryRepository{workers: workers}
}

func (r *MemoryRepository) Create(_ context.Context, worker *Worker) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.workers[worker.EmployeeID]; exists {
		return ErrAlreadyExists
	}
	r.workers[worker.EmployeeID] = *worker
	return nil
}

func (r *MemoryRepository) Update(_ context.Context, worker *Worker) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.workers[worker.EmployeeID]; !exists {
		return ErrNotFound
	}
	r.workers[worker.EmployeeID] = *worker
	return nil
}

func (r *MemoryRepository) Delete(_ context.Context, employeeID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.workers[employeeID]; !exists {
		return ErrNotFound
	}
	delete(r.workers, employeeID)
	return nil
}

func (r *MemoryRepository) FindByEmployeeID(_ context.Context, employeeID string) (*Worker, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	worker, exists := r.workers[employeeID]
	if !exists {
		return nil, nil
	}
	copy := worker
	return &copy, nil
}

func (r *MemoryRepository) List(_ context.Context) ([]Worker, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	workers := make([]Worker, 0, len(r.workers))
	for _, worker := range r.workers {
		workers = append(workers, worker)
	}
	sort.Slice(workers, func(i, j int) bool {
		return workers[i].EmployeeID < workers[j].EmployeeID
	})
	return workers, nil
}
