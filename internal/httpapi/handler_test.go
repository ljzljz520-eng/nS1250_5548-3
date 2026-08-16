package httpapi_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"warehouse-workers/internal/account"
	"warehouse-workers/internal/httpapi"
)

type responseError struct {
	Error struct {
		Code string `json:"code"`
	} `json:"error"`
}

func TestWorkerLifecycleThroughAPI(t *testing.T) {
	handler := newHandler(nil)
	worker := account.Worker{
		Name:       "Chen Yu",
		EmployeeID: "WH2001",
		Phone:      "+8613900000001",
		Email:      "CHEN.YU@example.com",
		Team:       "Packing B",
		Status:     account.StatusActive,
	}

	created := exchange(t, handler, http.MethodPost, "/workers", worker)
	if created.Code != http.StatusCreated {
		t.Fatalf("create status = %d, want %d", created.Code, http.StatusCreated)
	}

	detail := exchange(t, handler, http.MethodGet, "/workers/WH2001", nil)
	if detail.Code != http.StatusOK {
		t.Fatalf("detail status = %d, want %d", detail.Code, http.StatusOK)
	}
	var got account.Worker
	decodeResponse(t, detail, &got)
	if got.Email != "chen.yu@example.com" || got.Team != "Packing B" {
		t.Fatalf("created worker = %+v, want normalized email and Packing B", got)
	}

	update := account.UpdateWorker{
		Name:   "Chen Yu",
		Phone:  "+8613900000001",
		Email:  "chen.yu@example.com",
		Team:   "Dispatch C",
		Status: account.StatusInactive,
	}
	updated := exchange(t, handler, http.MethodPut, "/workers/WH2001", update)
	if updated.Code != http.StatusOK {
		t.Fatalf("update status = %d, want %d", updated.Code, http.StatusOK)
	}
	decodeResponse(t, updated, &got)
	if got.Team != "Dispatch C" || got.Status != account.StatusInactive {
		t.Fatalf("updated worker = %+v, want Dispatch C and inactive", got)
	}
}

func TestDeleteRequiresConfirmation(t *testing.T) {
	handler := newHandler([]account.Worker{worker("WH3001", "Inbound A")})

	response := exchange(t, handler, http.MethodDelete, "/workers/WH3001", nil)
	if response.Code != http.StatusBadRequest {
		t.Fatalf("delete status = %d, want %d", response.Code, http.StatusBadRequest)
	}
	var got responseError
	decodeResponse(t, response, &got)
	if got.Error.Code != "confirmation_required" {
		t.Fatalf("delete error code = %q, want %q", got.Error.Code, "confirmation_required")
	}

	detail := exchange(t, handler, http.MethodGet, "/workers/WH3001", nil)
	if detail.Code != http.StatusOK {
		t.Fatalf("detail status after unconfirmed delete = %d, want %d", detail.Code, http.StatusOK)
	}
}

func TestDeletedWorkerDetailReturnsNotFoundAndExistingWorkerRemainsAvailable(t *testing.T) {
	handler := newHandler([]account.Worker{
		worker("WH4001", "Inbound A"),
		worker("WH4002", "Outbound B"),
	})

	deleted := exchange(t, handler, http.MethodDelete, "/workers/WH4001?confirm=true", nil)
	if deleted.Code != http.StatusNoContent {
		t.Fatalf("delete status = %d, want %d", deleted.Code, http.StatusNoContent)
	}

	detail := exchange(t, handler, http.MethodGet, "/workers/WH4001", nil)
	if detail.Code != http.StatusNotFound {
		t.Errorf("deleted worker detail status = %d, want %d", detail.Code, http.StatusNotFound)
	}
	var response responseError
	decodeResponse(t, detail, &response)
	if response.Error.Code != "record_not_found" {
		t.Errorf("deleted worker detail error code = %q, want %q", response.Error.Code, "record_not_found")
	}

	existing := exchange(t, handler, http.MethodGet, "/workers/WH4002", nil)
	if existing.Code != http.StatusOK {
		t.Fatalf("existing worker detail status = %d, want %d", existing.Code, http.StatusOK)
	}
	var got account.Worker
	decodeResponse(t, existing, &got)
	if got.EmployeeID != "WH4002" {
		t.Fatalf("existing worker employee ID = %q, want %q", got.EmployeeID, "WH4002")
	}
}

func TestCreateRejectsInvalidContactDetails(t *testing.T) {
	handler := newHandler(nil)
	invalid := account.Worker{
		Name:       "Chen Yu",
		EmployeeID: "WH5001",
		Phone:      "not-a-phone",
		Email:      "not-an-email",
		Team:       "Packing B",
		Status:     account.StatusActive,
	}

	response := exchange(t, handler, http.MethodPost, "/workers", invalid)
	if response.Code != http.StatusBadRequest {
		t.Fatalf("create status = %d, want %d", response.Code, http.StatusBadRequest)
	}
	var got responseError
	decodeResponse(t, response, &got)
	if got.Error.Code != "validation_failed" {
		t.Fatalf("create error code = %q, want %q", got.Error.Code, "validation_failed")
	}
}

func newHandler(initial []account.Worker) http.Handler {
	repository := account.NewMemoryRepository(initial)
	return httpapi.New(account.NewService(repository)).Handler()
}

func worker(employeeID, team string) account.Worker {
	return account.Worker{
		Name:       "Warehouse Worker",
		EmployeeID: employeeID,
		Phone:      "+8613900000000",
		Email:      employeeID + "@example.com",
		Team:       team,
		Status:     account.StatusActive,
	}
}

func exchange(t *testing.T, handler http.Handler, method, target string, body any) *httptest.ResponseRecorder {
	t.Helper()
	var encoded bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&encoded).Encode(body); err != nil {
			t.Fatalf("encode request: %v", err)
		}
	}
	request := httptest.NewRequest(method, target, &encoded)
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	return response
}

func decodeResponse(t *testing.T, response *httptest.ResponseRecorder, destination any) {
	t.Helper()
	if err := json.NewDecoder(response.Body).Decode(destination); err != nil {
		t.Fatalf("decode response: %v", err)
	}
}
