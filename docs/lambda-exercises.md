# Saved Exercises - Lambda API Endpoints

Specification for 3 Lambda functions that provide the backend API for the saved exercises feature. All functions live under `functions/` and follow the same patterns as `save-workout` and `read-workout`.

## Shared Conventions

- **DynamoDB table:** Same single-table as workouts (TABLE_NAME env var)
- **Key design:** PK = `userId` (string, from JWT `sub` claim), SK = `EXERCISE#<savedExerciseId>`
- **Auth:** JWT authorizer via API Gateway. Extract `sub` from `req.RequestContext.Authorizer["jwt"]["claims"]["sub"]`
- **Region:** `eu-west-2` (hardcoded in `config.LoadDefaultConfig`)
- **Response helpers:** Each function defines its own `response()` and `errorBody()` (duplicated per function, same as existing pattern)
- **Build:** `make <function-name>` — the makefile auto-discovers directories under `functions/`
- **Go module:** Root `go.mod` at `dybev.uk/workout-tracker`, shared by all functions

## DynamoDB Item Schema (SavedExercise)

```json
{
  "userId":           "string (PK - user's Cognito sub)",
  "sk":               "string (SK - EXERCISE#<savedExerciseId>)",
  "savedExerciseId":  "string",
  "name":             "string",
  "note":             "string",
  "tags":             ["string"],
  "archivedAt":       "string | null (ISO 8601 timestamp or omitted)",
  "createdAt":        "string (ISO 8601)",
  "updatedAt":        "string (ISO 8601)"
}
```

## Go Type Definition

Each function should define its own types (duplicated, not shared). The `SavedExercise` struct:

```go
type SavedExercise struct {
    UserID          string   `json:"userId,omitempty" dynamodbav:"userId"`
    SK              string   `json:"sk,omitempty" dynamodbav:"sk"`
    SavedExerciseID string   `json:"savedExerciseId" dynamodbav:"savedExerciseId"`
    Name            string   `json:"name" dynamodbav:"name"`
    Note            string   `json:"note" dynamodbav:"note"`
    Tags            []string `json:"tags" dynamodbav:"tags"`
    ArchivedAt      *string  `json:"archivedAt" dynamodbav:"archivedAt"`
    CreatedAt       string   `json:"createdAt" dynamodbav:"createdAt"`
    UpdatedAt       string   `json:"updatedAt" dynamodbav:"updatedAt"`
}
```

---

## 1. exercises-read (GET)

**Path:** `functions/exercises-read/`
**Files:** `main.go`, `types.go`, `handler.go`, `handler_test.go`
**HTTP Method:** GET
**API Route:** `/exercises/read`

### Behaviour

- Query DynamoDB with `userId = :uid AND begins_with(sk, :prefix)` where `:prefix = "EXERCISE#"`
- No pagination — return all exercises (saved exercises per user should be manageable)
- Returns ALL exercises including archived (frontend filters)
- No `ScanIndexForward` preference needed (order doesn't matter)

### DynamoDB Interface

```go
type DynamoQuerier interface {
    Query(ctx context.Context, params *dynamodb.QueryInput, optFns ...func(*dynamodb.Options)) (*dynamodb.QueryOutput, error)
}
```

### Response (200 OK)

```json
{
  "exercises": [
    {
      "savedExerciseId": "abc123",
      "name": "Bench Press",
      "note": "Keep elbows tucked",
      "tags": ["chest", "push"],
      "archivedAt": null,
      "createdAt": "2026-03-01T10:00:00.000Z",
      "updatedAt": "2026-03-15T14:30:00.000Z"
    }
  ]
}
```

### Error Responses

- `401` — missing/invalid auth
- `500` — DynamoDB error (return the error from HandleRequest for Lambda logging)

### Tests

- Rejects unauthenticated request
- Rejects missing JWT claims
- Rejects missing table name
- Returns empty array when no exercises
- Returns exercises correctly (verify field unmarshalling)
- Handles DynamoDB error
- Query uses correct table name
- Query uses EXERCISE# prefix
- Query is scoped to current user

---

## 2. exercises-save (POST)

**Path:** `functions/exercises-save/`
**Files:** `main.go`, `types.go`, `handler.go`, `handler_test.go`
**HTTP Method:** POST
**API Route:** `/exercises/save`

### Behaviour

- Accepts a JSON array of `SavedExercise` objects in the request body
- Sets `userId` from JWT `sub` claim on each exercise
- Generates SK as `EXERCISE#<savedExerciseId>` for each exercise
- Validates required fields: `savedExerciseId`, `name`, `createdAt`, `updatedAt`
- Uses `BatchWriteItem` with chunking (max 25) and exponential backoff retry (same pattern as `save-workout`)
- This is used for both creating new exercises AND updating existing ones (PutItem is an upsert)

### DynamoDB Interface

```go
type DynamoBatchWriter interface {
    BatchWriteItem(ctx context.Context, params *dynamodb.BatchWriteItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.BatchWriteItemOutput, error)
}
```

### Request Body

```json
[
  {
    "savedExerciseId": "abc123",
    "name": "Bench Press",
    "note": "Keep elbows tucked",
    "tags": ["chest", "push"],
    "archivedAt": null,
    "createdAt": "2026-03-01T10:00:00.000Z",
    "updatedAt": "2026-03-15T14:30:00.000Z"
  }
]
```

### Additional Types

```go
type SaveResult struct {
    SavedExerciseID string `json:"savedExerciseId"`
    Status          string `json:"status"`
    Error           string `json:"error,omitempty"`
}
```

### Response (201 Created)

```json
{
  "message": "batch complete",
  "results": [
    { "savedExerciseId": "abc123", "status": "saved" }
  ]
}
```

### Error Responses

- `400` — invalid JSON, empty array, missing required fields
- `401` — missing/invalid auth
- `500` — DynamoDB error

### Tests

- Rejects unauthenticated request
- Rejects empty body
- Rejects empty array
- Rejects exercise missing savedExerciseId
- Rejects exercise missing name
- Rejects exercise missing createdAt
- Rejects exercise missing updatedAt
- Saves single exercise successfully
- Saves multiple exercises successfully
- Sets userId from JWT on all exercises
- Generates correct SK (EXERCISE# prefix)
- Handles DynamoDB error
- Reports failed exercises in results

---

## 3. exercises-archive (PATCH)

**Path:** `functions/exercises-archive/`
**Files:** `main.go`, `types.go`, `handler.go`, `handler_test.go`
**HTTP Method:** PATCH
**API Route:** `/exercises/archive`

### Behaviour

- Accepts a JSON body with `savedExerciseId` and `archive` (boolean)
- If `archive = true`: UpdateItem to SET `archivedAt` to current ISO timestamp and `updatedAt` to current ISO timestamp
- If `archive = false`: UpdateItem to REMOVE `archivedAt` and SET `updatedAt` to current ISO timestamp
- Key: `userId` (from JWT) + `sk = EXERCISE#<savedExerciseId>`
- Uses a condition expression `attribute_exists(sk)` to ensure the item exists

### DynamoDB Interface

```go
type DynamoUpdater interface {
    UpdateItem(ctx context.Context, params *dynamodb.UpdateItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.UpdateItemOutput, error)
}
```

### Request Body

```json
{
  "savedExerciseId": "abc123",
  "archive": true
}
```

### Additional Types

```go
type ArchiveRequest struct {
    SavedExerciseID string `json:"savedExerciseId"`
    Archive         bool   `json:"archive"`
}
```

### Response (200 OK)

```json
{
  "message": "exercise archived"
}
```

Or for restore:

```json
{
  "message": "exercise restored"
}
```

### Error Responses

- `400` — invalid JSON, missing `savedExerciseId`
- `401` — missing/invalid auth
- `404` — exercise not found (condition check failed)
- `500` — DynamoDB error

### Tests

- Rejects unauthenticated request
- Rejects empty body
- Rejects missing savedExerciseId
- Archives exercise (archive=true) — verify UpdateExpression sets archivedAt and updatedAt
- Restores exercise (archive=false) — verify UpdateExpression removes archivedAt and sets updatedAt
- Uses correct key (userId + EXERCISE#savedExerciseId)
- Condition expression checks item exists
- Returns 404 when item not found (ConditionalCheckFailedException)
- Handles DynamoDB error

---

## Frontend API Client

The frontend API client is already implemented at `packages/app/src/services/exerciseApi.ts` with three functions:

- `readExercises()` — GET `/exercises/read`
- `saveExercises(exercises)` — POST `/exercises/save`
- `archiveExercise(savedExerciseId, archive)` — PATCH `/exercises/archive`

All use Bearer token auth with `idToken` from Cognito session.
