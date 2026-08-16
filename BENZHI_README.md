# Warehouse Worker Accounts

A Go 1.25 HTTP service for warehouse worker account management. It stores data in memory and exposes create, list, detail, edit, and confirmed-delete operations.

## Run locally

From the module root:

```bash
go mod download
go run ./cmd/server -addr :8080
```

The example server starts with employee `WH-1001`. Check it with:

```bash
curl http://localhost:8080/healthz
curl http://localhost:8080/workers/WH-1001
```

Run every business-path test from the module root:

```bash
go test -count=1 ./...
```

The deleted-worker detail regression currently exposes the intentionally injected defect: the expected `404 record_not_found` response is replaced by `500 internal_error`. The same test also verifies that another worker remains queryable.

## Data model

Each worker has these fields:

| Field | JSON name | Rule |
| --- | --- | --- |
| Name | `name` | 2-80 characters |
| Employee ID | `employee_id` | 3-32 letters or digits, unique |
| Phone | `phone` | E.164 format |
| Email | `email` | Valid email address |
| Team | `team` | 2-80 characters |
| Status | `status` | `active` or `inactive` |

Data lasts for the lifetime of the process.

## API

All JSON responses use `Content-Type: application/json`. Errors have the form:

```json
{"error":{"code":"validation_failed","message":"worker fields are invalid"}}
```

### Create a worker

`POST /workers`

```bash
curl -i http://localhost:8080/workers \
  -H 'Content-Type: application/json' \
  -d '{"name":"Chen Yu","employee_id":"WH2001","phone":"+8613900000001","email":"chen.yu@example.com","team":"Packing B","status":"active"}'
```

Returns `201`. Duplicate employee IDs return `409 worker_exists`, and invalid fields return `400 validation_failed`.

### List workers

`GET /workers` returns `200` with workers ordered by employee ID.

```bash
curl http://localhost:8080/workers
```

### Get worker details

`GET /workers/{employee_id}` returns `200` with one worker. A missing worker should return `404 record_not_found`.

```bash
curl http://localhost:8080/workers/WH2001
```

### Edit a worker

`PUT /workers/{employee_id}` keeps the employee ID from the path and replaces the other fields.

```bash
curl -i -X PUT http://localhost:8080/workers/WH2001 \
  -H 'Content-Type: application/json' \
  -d '{"name":"Chen Yu","phone":"+8613900000001","email":"chen.yu@example.com","team":"Dispatch C","status":"inactive"}'
```

Returns `200`, `404 record_not_found`, or `400 validation_failed`.

### Delete a worker

Deletion must be explicitly confirmed with `confirm=true`.

```bash
curl -i -X DELETE 'http://localhost:8080/workers/WH2001?confirm=true'
```

Returns `204`. Without confirmation it returns `400 confirmation_required`; an unknown employee ID returns `404 record_not_found`.
