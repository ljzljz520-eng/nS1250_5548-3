package httpapi

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"

	"warehouse-workers/internal/account"
)

type API struct {
	service *account.Service
}

type errorBody struct {
	Error apiError `json:"error"`
}

type apiError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func New(service *account.Service) *API {
	return &API{service: service}
}

func (a *API) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", a.health)
	mux.HandleFunc("GET /workers", a.listWorkers)
	mux.HandleFunc("POST /workers", a.createWorker)
	mux.HandleFunc("GET /workers/{employeeID}", a.getWorker)
	mux.HandleFunc("PUT /workers/{employeeID}", a.updateWorker)
	mux.HandleFunc("DELETE /workers/{employeeID}", a.deleteWorker)
	return recoverRequests(mux)
}

func (a *API) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (a *API) listWorkers(w http.ResponseWriter, r *http.Request) {
	workers, err := a.service.List(r.Context())
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"workers": workers})
}

func (a *API) createWorker(w http.ResponseWriter, r *http.Request) {
	var worker account.Worker
	if err := decodeJSON(r, &worker); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "request body must be a single valid worker object")
		return
	}
	created, err := a.service.Create(r.Context(), worker)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

func (a *API) getWorker(w http.ResponseWriter, r *http.Request) {
	worker, err := a.service.Get(r.Context(), r.PathValue("employeeID"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, worker)
}

func (a *API) updateWorker(w http.ResponseWriter, r *http.Request) {
	var update account.UpdateWorker
	if err := decodeJSON(r, &update); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "request body must be a single valid update object")
		return
	}
	worker, err := a.service.Update(r.Context(), r.PathValue("employeeID"), update)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, worker)
}

func (a *API) deleteWorker(w http.ResponseWriter, r *http.Request) {
	confirmed, _ := strconv.ParseBool(r.URL.Query().Get("confirm"))
	if err := a.service.Delete(r.Context(), r.PathValue("employeeID"), confirmed); err != nil {
		writeServiceError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func decodeJSON(r *http.Request, destination any) error {
	decoder := json.NewDecoder(io.LimitReader(r.Body, 1<<20))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(destination); err != nil {
		return err
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return errors.New("request body must contain one JSON value")
	}
	return nil
}

func writeServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, account.ErrInvalidWorker):
		writeError(w, http.StatusBadRequest, "validation_failed", "worker fields are invalid")
	case errors.Is(err, account.ErrConfirmationRequired):
		writeError(w, http.StatusBadRequest, "confirmation_required", "set confirm=true to delete the worker")
	case errors.Is(err, account.ErrAlreadyExists):
		writeError(w, http.StatusConflict, "worker_exists", "employee ID already exists")
	case errors.Is(err, account.ErrNotFound):
		writeError(w, http.StatusNotFound, "record_not_found", "worker record does not exist")
	default:
		writeError(w, http.StatusInternalServerError, "internal_error", "unexpected server error")
	}
}

func recoverRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if recover() != nil {
				writeError(w, http.StatusInternalServerError, "internal_error", "unexpected server error")
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, errorBody{Error: apiError{Code: code, Message: message}})
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
