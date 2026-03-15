package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// mockDynamo implements DynamoBatchWriter for testing.
type mockDynamo struct {
	batchWriteFunc func(ctx context.Context, params *dynamodb.BatchWriteItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.BatchWriteItemOutput, error)
	calls          int
	lastInput      *dynamodb.BatchWriteItemInput
}

func (m *mockDynamo) BatchWriteItem(ctx context.Context, params *dynamodb.BatchWriteItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.BatchWriteItemOutput, error) {
	m.calls++
	m.lastInput = params
	if m.batchWriteFunc != nil {
		return m.batchWriteFunc(ctx, params, optFns...)
	}
	return &dynamodb.BatchWriteItemOutput{}, nil
}

func validWorkoutJSON() string {
	reps := 10
	weight := 80.0
	w := Workout{
		UserID:    "user-123",
		SK:        "WORKOUT#2026-03-14T10:00:00.000Z#wkt-abc",
		WorkoutID: "wkt-abc",
		StartedAt: "2026-03-14T10:00:00.000Z",
		Notes:     "Great session",
		Tags:      []string{"push", "chest"},
		BodyWeight: &BodyWeight{
			Value: 75.5,
			Unit:  "kg",
		},
		Exercises: []WorkoutExercise{
			{
				ExerciseID: "ex-1",
				Name:       "Bench Press",
				Order:      0,
				Sets: []WorkoutSet{
					{SetID: "s-1", Order: 0, Reps: &reps, Weight: &weight, WeightUnit: "kg"},
					{SetID: "s-2", Order: 1, Reps: &reps, Weight: nil, WeightUnit: "lbs"},
				},
			},
		},
		CreatedAt: "2026-03-14T10:00:00.000Z",
		UpdatedAt: "2026-03-14T10:30:00.000Z",
	}
	b, _ := json.Marshal(w)
	return string(b)
}

func minimalWorkoutJSON() string {
	w := Workout{
		UserID:    "user-456",
		SK:        "WORKOUT#2026-03-14T12:00:00.000Z#wkt-def",
		WorkoutID: "wkt-def",
		StartedAt: "2026-03-14T12:00:00.000Z",
		Notes:     "",
		Tags:      []string{},
		Exercises: []WorkoutExercise{},
		CreatedAt: "2026-03-14T12:00:00.000Z",
		UpdatedAt: "2026-03-14T12:00:00.000Z",
	}
	b, _ := json.Marshal(w)
	return string(b)
}

func validWorkoutArrayJSON() string {
	reps := 10
	weight := 80.0
	workouts := []Workout{
		{
			UserID:    "user-123",
			SK:        "WORKOUT#2026-03-14T10:00:00.000Z#wkt-abc",
			WorkoutID: "wkt-abc",
			StartedAt: "2026-03-14T10:00:00.000Z",
			Exercises: []WorkoutExercise{
				{
					ExerciseID: "ex-1",
					Name:       "Bench Press",
					Order:      0,
					Sets: []WorkoutSet{
						{SetID: "s-1", Order: 0, Reps: &reps, Weight: &weight, WeightUnit: "kg"},
					},
				},
			},
			CreatedAt: "2026-03-14T10:00:00.000Z",
			UpdatedAt: "2026-03-14T10:30:00.000Z",
		},
		{
			UserID:    "user-123",
			SK:        "WORKOUT#2026-03-14T11:00:00.000Z#wkt-def",
			WorkoutID: "wkt-def",
			StartedAt: "2026-03-14T11:00:00.000Z",
			Exercises: []WorkoutExercise{},
			CreatedAt: "2026-03-14T11:00:00.000Z",
			UpdatedAt: "2026-03-14T11:30:00.000Z",
		},
	}
	b, _ := json.Marshal(workouts)
	return string(b)
}

func makeRequest(method, body string) events.APIGatewayProxyRequest {
	return events.APIGatewayProxyRequest{
		HTTPMethod: method,
		Body:       body,
	}
}

func parseResponseBody(t *testing.T, body string) map[string]string {
	t.Helper()
	var m map[string]string
	if err := json.Unmarshal([]byte(body), &m); err != nil {
		t.Fatalf("failed to parse response body: %v", err)
	}
	return m
}

func parseBatchResponseBody(t *testing.T, body string) (string, []SaveResult) {
	t.Helper()
	var m struct {
		Message string       `json:"message"`
		Results []SaveResult `json:"results"`
	}
	if err := json.Unmarshal([]byte(body), &m); err != nil {
		t.Fatalf("failed to parse batch response body: %v", err)
	}
	return m.Message, m.Results
}

// ── Tests: HTTP method handling ─────────────────────────────────────────────

func TestRejectsNonPostMethods(t *testing.T) {
	mock := &mockDynamo{}
	h := NewHandler(mock, "TestTable")

	methods := []string{http.MethodGet, http.MethodPut, http.MethodDelete, http.MethodPatch}
	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			resp, err := h.HandleRequest(context.Background(), makeRequest(method, ""))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if resp.StatusCode != http.StatusMethodNotAllowed {
				t.Errorf("expected status %d, got %d", http.StatusMethodNotAllowed, resp.StatusCode)
			}
			body := parseResponseBody(t, resp.Body)
			if body["error"] != "method not allowed" {
				t.Errorf("expected error 'method not allowed', got %q", body["error"])
			}
		})
	}

	if mock.calls > 0 {
		t.Error("DynamoDB should not have been called for non-POST methods")
	}
}

// ── Tests: Request body validation ──────────────────────────────────────────

func TestRejectsEmptyBody(t *testing.T) {
	mock := &mockDynamo{}
	h := NewHandler(mock, "TestTable")

	resp, err := h.HandleRequest(context.Background(), makeRequest(http.MethodPost, ""))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, resp.StatusCode)
	}
	body := parseResponseBody(t, resp.Body)
	if body["error"] != "request body is empty" {
		t.Errorf("expected 'request body is empty', got %q", body["error"])
	}
}

func TestRejectsInvalidJSON(t *testing.T) {
	mock := &mockDynamo{}
	h := NewHandler(mock, "TestTable")

	resp, err := h.HandleRequest(context.Background(), makeRequest(http.MethodPost, "{not valid json"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, resp.StatusCode)
	}
}

func TestRejectsMissingRequiredFields(t *testing.T) {
	tests := []struct {
		name    string
		payload string
		wantErr string
	}{
		{
			name:    "missing userId",
			payload: `{"sk":"WORKOUT#t#id","workoutId":"id","startedAt":"t","createdAt":"t","updatedAt":"t"}`,
			wantErr: "missing required fields: userId",
		},
		{
			name:    "missing sk",
			payload: `{"userId":"u","workoutId":"id","startedAt":"t","createdAt":"t","updatedAt":"t"}`,
			wantErr: "missing required fields: sk",
		},
		{
			name:    "missing workoutId",
			payload: `{"userId":"u","sk":"s","startedAt":"t","createdAt":"t","updatedAt":"t"}`,
			wantErr: "missing required fields: workoutId",
		},
		{
			name:    "missing startedAt",
			payload: `{"userId":"u","sk":"s","workoutId":"id","createdAt":"t","updatedAt":"t"}`,
			wantErr: "missing required fields: startedAt",
		},
		{
			name:    "missing createdAt",
			payload: `{"userId":"u","sk":"s","workoutId":"id","startedAt":"t","updatedAt":"t"}`,
			wantErr: "missing required fields: createdAt",
		},
		{
			name:    "missing updatedAt",
			payload: `{"userId":"u","sk":"s","workoutId":"id","startedAt":"t","createdAt":"t"}`,
			wantErr: "missing required fields: updatedAt",
		},
		{
			name:    "multiple missing",
			payload: `{"startedAt":"t","createdAt":"t","updatedAt":"t"}`,
			wantErr: "missing required fields: userId, sk, workoutId",
		},
	}

	mock := &mockDynamo{}
	h := NewHandler(mock, "TestTable")

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			resp, err := h.HandleRequest(context.Background(), makeRequest(http.MethodPost, tc.payload))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if resp.StatusCode != http.StatusBadRequest {
				t.Errorf("expected status %d, got %d", http.StatusBadRequest, resp.StatusCode)
			}
			body := parseResponseBody(t, resp.Body)
			if body["error"] != tc.wantErr {
				t.Errorf("expected error %q, got %q", tc.wantErr, body["error"])
			}
		})
	}
}

// ── Tests: Exercise and set validation ──────────────────────────────────────

func TestRejectsExerciseMissingFields(t *testing.T) {
	tests := []struct {
		name    string
		payload string
		wantErr string
	}{
		{
			name: "missing exerciseId",
			payload: `{
				"userId":"u","sk":"s","workoutId":"id","startedAt":"t","createdAt":"t","updatedAt":"t",
				"exercises":[{"name":"Bench","order":0,"sets":[]}]
			}`,
			wantErr: "exercises[0]: missing exerciseId",
		},
		{
			name: "missing name",
			payload: `{
				"userId":"u","sk":"s","workoutId":"id","startedAt":"t","createdAt":"t","updatedAt":"t",
				"exercises":[{"exerciseId":"e1","order":0,"sets":[]}]
			}`,
			wantErr: "exercises[0]: missing name",
		},
	}

	mock := &mockDynamo{}
	h := NewHandler(mock, "TestTable")

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			resp, err := h.HandleRequest(context.Background(), makeRequest(http.MethodPost, tc.payload))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if resp.StatusCode != http.StatusBadRequest {
				t.Errorf("expected status %d, got %d", http.StatusBadRequest, resp.StatusCode)
			}
			body := parseResponseBody(t, resp.Body)
			if body["error"] != tc.wantErr {
				t.Errorf("expected error %q, got %q", tc.wantErr, body["error"])
			}
		})
	}
}

func TestRejectsSetMissingFields(t *testing.T) {
	tests := []struct {
		name    string
		payload string
		wantErr string
	}{
		{
			name: "missing setId",
			payload: `{
				"userId":"u","sk":"s","workoutId":"id","startedAt":"t","createdAt":"t","updatedAt":"t",
				"exercises":[{"exerciseId":"e1","name":"Bench","order":0,"sets":[{"order":0,"weightUnit":"kg"}]}]
			}`,
			wantErr: "exercises[0].sets[0]: missing setId",
		},
		{
			name: "invalid weightUnit",
			payload: `{
				"userId":"u","sk":"s","workoutId":"id","startedAt":"t","createdAt":"t","updatedAt":"t",
				"exercises":[{"exerciseId":"e1","name":"Bench","order":0,"sets":[{"setId":"s1","order":0,"weightUnit":"stones"}]}]
			}`,
			wantErr: `exercises[0].sets[0]: weightUnit must be "kg" or "lbs"`,
		},
	}

	mock := &mockDynamo{}
	h := NewHandler(mock, "TestTable")

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			resp, err := h.HandleRequest(context.Background(), makeRequest(http.MethodPost, tc.payload))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if resp.StatusCode != http.StatusBadRequest {
				t.Errorf("expected status %d, got %d", http.StatusBadRequest, resp.StatusCode)
			}
			body := parseResponseBody(t, resp.Body)
			if body["error"] != tc.wantErr {
				t.Errorf("expected error %q, got %q", tc.wantErr, body["error"])
			}
		})
	}
}

// ── Tests: BodyWeight validation ────────────────────────────────────────────

func TestRejectsInvalidBodyWeightUnit(t *testing.T) {
	mock := &mockDynamo{}
	h := NewHandler(mock, "TestTable")

	payload := `{
		"userId":"u","sk":"s","workoutId":"id","startedAt":"t","createdAt":"t","updatedAt":"t",
		"bodyWeight":{"value":80,"unit":"stones"}
	}`

	resp, err := h.HandleRequest(context.Background(), makeRequest(http.MethodPost, payload))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, resp.StatusCode)
	}
	body := parseResponseBody(t, resp.Body)
	expected := `bodyWeight.unit must be "kg" or "lbs"`
	if body["error"] != expected {
		t.Errorf("expected error %q, got %q", expected, body["error"])
	}
}

// ── Tests: Successful save (single workout) ─────────────────────────────────

func TestSaveFullWorkout(t *testing.T) {
	mock := &mockDynamo{}
	h := NewHandler(mock, "TestTable")

	resp, err := h.HandleRequest(context.Background(), makeRequest(http.MethodPost, validWorkoutJSON()))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Errorf("expected status %d, got %d", http.StatusCreated, resp.StatusCode)
	}

	body := parseResponseBody(t, resp.Body)
	if body["message"] != "workout saved" {
		t.Errorf("expected message 'workout saved', got %q", body["message"])
	}
	if body["workoutId"] != "wkt-abc" {
		t.Errorf("expected workoutId 'wkt-abc', got %q", body["workoutId"])
	}

	if mock.calls != 1 {
		t.Fatalf("expected BatchWriteItem to be called once, got %d", mock.calls)
	}
	reqs, ok := mock.lastInput.RequestItems["TestTable"]
	if !ok || len(reqs) == 0 {
		t.Fatal("expected request items for TestTable")
	}

	// Verify key attributes are in the DynamoDB item.
	item := mock.lastInput.RequestItems["TestTable"][0].PutRequest.Item
	if item["userId"] == nil {
		t.Error("expected userId in DynamoDB item")
	}
	if item["sk"] == nil {
		t.Error("expected sk in DynamoDB item")
	}
	if item["workoutId"] == nil {
		t.Error("expected workoutId in DynamoDB item")
	}
	if item["exercises"] == nil {
		t.Error("expected exercises in DynamoDB item")
	}
	if item["bodyWeight"] == nil {
		t.Error("expected bodyWeight in DynamoDB item")
	}
	if item["tags"] == nil {
		t.Error("expected tags in DynamoDB item")
	}
}

func TestSaveMinimalWorkout(t *testing.T) {
	mock := &mockDynamo{}
	h := NewHandler(mock, "TestTable")

	resp, err := h.HandleRequest(context.Background(), makeRequest(http.MethodPost, minimalWorkoutJSON()))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Errorf("expected status %d, got %d", http.StatusCreated, resp.StatusCode)
	}

	body := parseResponseBody(t, resp.Body)
	if body["workoutId"] != "wkt-def" {
		t.Errorf("expected workoutId 'wkt-def', got %q", body["workoutId"])
	}

	if mock.calls != 1 {
		t.Fatalf("expected BatchWriteItem to be called once, got %d", mock.calls)
	}
}

// ── Tests: Nullable fields ──────────────────────────────────────────────────

func TestNullableFieldsAreAccepted(t *testing.T) {
	payload := `{
		"userId":"u","sk":"s","workoutId":"id","startedAt":"t","createdAt":"t","updatedAt":"t",
		"templateId": null,
		"completedAt": null,
		"bodyWeight": null,
		"notes": "",
		"tags": [],
		"exercises": []
	}`

	mock := &mockDynamo{}
	h := NewHandler(mock, "TestTable")

	resp, err := h.HandleRequest(context.Background(), makeRequest(http.MethodPost, payload))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Errorf("expected status %d, got %d", http.StatusCreated, resp.StatusCode)
	}
}

func TestNullableRepsAndWeight(t *testing.T) {
	payload := `{
		"userId":"u","sk":"s","workoutId":"id","startedAt":"t","createdAt":"t","updatedAt":"t",
		"exercises":[{
			"exerciseId":"e1","name":"Bench","order":0,
			"sets":[{"setId":"s1","order":0,"reps":null,"weight":null,"weightUnit":"kg"}]
		}]
	}`

	mock := &mockDynamo{}
	h := NewHandler(mock, "TestTable")

	resp, err := h.HandleRequest(context.Background(), makeRequest(http.MethodPost, payload))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Errorf("expected status %d, got %d", http.StatusCreated, resp.StatusCode)
	}
}

// ── Tests: DynamoDB failure (single workout) ────────────────────────────────

func TestDynamoDBBatchWriteError(t *testing.T) {
	mock := &mockDynamo{
		batchWriteFunc: func(ctx context.Context, params *dynamodb.BatchWriteItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.BatchWriteItemOutput, error) {
			return nil, errors.New("service unavailable")
		},
	}
	h := NewHandler(mock, "TestTable")

	resp, err := h.HandleRequest(context.Background(), makeRequest(http.MethodPost, validWorkoutJSON()))
	if err == nil {
		t.Fatal("expected error to be returned")
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, resp.StatusCode)
	}
}

// ── Tests: DynamoDB item shape ──────────────────────────────────────────────

func TestDynamoDBItemContainsAllFields(t *testing.T) {
	mock := &mockDynamo{}
	h := NewHandler(mock, "TestTable")

	_, err := h.HandleRequest(context.Background(), makeRequest(http.MethodPost, validWorkoutJSON()))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	item := mock.lastInput.RequestItems["TestTable"][0].PutRequest.Item
	expectedKeys := []string{
		"userId", "sk", "workoutId", "startedAt", "notes",
		"tags", "bodyWeight", "exercises", "createdAt", "updatedAt",
	}

	for _, key := range expectedKeys {
		if item[key] == nil {
			t.Errorf("expected key %q in DynamoDB item, but it was missing", key)
		}
	}
}

func TestTableNameIsPassedCorrectly(t *testing.T) {
	tableName := "MyCustomTable"
	mock := &mockDynamo{}
	h := NewHandler(mock, tableName)

	_, err := h.HandleRequest(context.Background(), makeRequest(http.MethodPost, validWorkoutJSON()))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, ok := mock.lastInput.RequestItems[tableName]; !ok {
		t.Errorf("expected request items for table %q, got keys: %v", tableName, mock.lastInput.RequestItems)
	}
}

// ── Tests: Response format ──────────────────────────────────────────────────

func TestResponseHasJSONContentType(t *testing.T) {
	mock := &mockDynamo{}
	h := NewHandler(mock, "TestTable")

	resp, _ := h.HandleRequest(context.Background(), makeRequest(http.MethodPost, validWorkoutJSON()))
	ct := resp.Headers["Content-Type"]
	if ct != "application/json" {
		t.Errorf("expected Content-Type 'application/json', got %q", ct)
	}
}

func TestErrorResponseHasJSONContentType(t *testing.T) {
	mock := &mockDynamo{}
	h := NewHandler(mock, "TestTable")

	resp, _ := h.HandleRequest(context.Background(), makeRequest(http.MethodGet, ""))
	ct := resp.Headers["Content-Type"]
	if ct != "application/json" {
		t.Errorf("expected Content-Type 'application/json', got %q", ct)
	}
}

// ── Tests: Batch (array) requests ───────────────────────────────────────────

func TestSaveBatchWorkouts(t *testing.T) {
	mock := &mockDynamo{}
	h := NewHandler(mock, "TestTable")

	resp, err := h.HandleRequest(context.Background(), makeRequest(http.MethodPost, validWorkoutArrayJSON()))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Errorf("expected status %d, got %d", http.StatusCreated, resp.StatusCode)
	}

	msg, results := parseBatchResponseBody(t, resp.Body)
	if msg != "batch complete" {
		t.Errorf("expected message 'batch complete', got %q", msg)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	if results[0].WorkoutID != "wkt-abc" || results[0].Status != "saved" {
		t.Errorf("results[0]: expected saved wkt-abc, got %+v", results[0])
	}
	if results[1].WorkoutID != "wkt-def" || results[1].Status != "saved" {
		t.Errorf("results[1]: expected saved wkt-def, got %+v", results[1])
	}

	// Both workouts should be in a single BatchWriteItem call.
	if mock.calls != 1 {
		t.Errorf("expected 1 BatchWriteItem call, got %d", mock.calls)
	}
	if len(mock.lastInput.RequestItems["TestTable"]) != 2 {
		t.Errorf("expected 2 write requests, got %d", len(mock.lastInput.RequestItems["TestTable"]))
	}
}

func TestBatchRejectsEmptyArray(t *testing.T) {
	mock := &mockDynamo{}
	h := NewHandler(mock, "TestTable")

	resp, err := h.HandleRequest(context.Background(), makeRequest(http.MethodPost, "[]"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, resp.StatusCode)
	}
	body := parseResponseBody(t, resp.Body)
	if body["error"] != "workout array is empty" {
		t.Errorf("expected 'workout array is empty', got %q", body["error"])
	}
}

func TestBatchRejectsInvalidJSON(t *testing.T) {
	mock := &mockDynamo{}
	h := NewHandler(mock, "TestTable")

	resp, err := h.HandleRequest(context.Background(), makeRequest(http.MethodPost, "[not valid]"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, resp.StatusCode)
	}
}

func TestBatchRejectsValidationFailureInOneWorkout(t *testing.T) {
	payload := `[
		{"userId":"u","sk":"s","workoutId":"w1","startedAt":"t","createdAt":"t","updatedAt":"t"},
		{"sk":"s","workoutId":"w2","startedAt":"t","createdAt":"t","updatedAt":"t"}
	]`

	mock := &mockDynamo{}
	h := NewHandler(mock, "TestTable")

	resp, err := h.HandleRequest(context.Background(), makeRequest(http.MethodPost, payload))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, resp.StatusCode)
	}
	body := parseResponseBody(t, resp.Body)
	expected := "workouts[1]: missing required fields: userId"
	if body["error"] != expected {
		t.Errorf("expected error %q, got %q", expected, body["error"])
	}
	if mock.calls > 0 {
		t.Error("DynamoDB should not have been called when validation fails")
	}
}

func TestBatchDynamoDBError(t *testing.T) {
	mock := &mockDynamo{
		batchWriteFunc: func(ctx context.Context, params *dynamodb.BatchWriteItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.BatchWriteItemOutput, error) {
			return nil, errors.New("service unavailable")
		},
	}
	h := NewHandler(mock, "TestTable")

	resp, err := h.HandleRequest(context.Background(), makeRequest(http.MethodPost, validWorkoutArrayJSON()))
	if err == nil {
		t.Fatal("expected error to be returned")
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, resp.StatusCode)
	}
}

func TestBatchResponseHasJSONContentType(t *testing.T) {
	mock := &mockDynamo{}
	h := NewHandler(mock, "TestTable")

	resp, _ := h.HandleRequest(context.Background(), makeRequest(http.MethodPost, validWorkoutArrayJSON()))
	ct := resp.Headers["Content-Type"]
	if ct != "application/json" {
		t.Errorf("expected Content-Type 'application/json', got %q", ct)
	}
}

func TestSingleWorkoutStillWorks(t *testing.T) {
	mock := &mockDynamo{}
	h := NewHandler(mock, "TestTable")

	resp, err := h.HandleRequest(context.Background(), makeRequest(http.MethodPost, validWorkoutJSON()))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Errorf("expected status %d, got %d", http.StatusCreated, resp.StatusCode)
	}

	body := parseResponseBody(t, resp.Body)
	if body["message"] != "workout saved" {
		t.Errorf("expected message 'workout saved', got %q", body["message"])
	}
	if body["workoutId"] != "wkt-abc" {
		t.Errorf("expected workoutId 'wkt-abc', got %q", body["workoutId"])
	}
}

// ── Tests: UnprocessedItems retry ───────────────────────────────────────────

func TestRetriesUnprocessedItems(t *testing.T) {
	callCount := 0
	mock := &mockDynamo{
		batchWriteFunc: func(ctx context.Context, params *dynamodb.BatchWriteItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.BatchWriteItemOutput, error) {
			callCount++
			// First call: return one item as unprocessed.
			if callCount == 1 {
				return &dynamodb.BatchWriteItemOutput{
					UnprocessedItems: params.RequestItems,
				}, nil
			}
			// Second call (retry): succeed.
			return &dynamodb.BatchWriteItemOutput{}, nil
		},
	}
	h := NewHandler(mock, "TestTable")

	resp, err := h.HandleRequest(context.Background(), makeRequest(http.MethodPost, validWorkoutJSON()))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Errorf("expected status %d, got %d", http.StatusCreated, resp.StatusCode)
	}
	if callCount != 2 {
		t.Errorf("expected 2 BatchWriteItem calls (1 initial + 1 retry), got %d", callCount)
	}
	body := parseResponseBody(t, resp.Body)
	if body["message"] != "workout saved" {
		t.Errorf("expected message 'workout saved', got %q", body["message"])
	}
}

func TestUnprocessedItemsExhaustRetries(t *testing.T) {
	callCount := 0
	mock := &mockDynamo{
		batchWriteFunc: func(ctx context.Context, params *dynamodb.BatchWriteItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.BatchWriteItemOutput, error) {
			callCount++
			// Always return items as unprocessed.
			return &dynamodb.BatchWriteItemOutput{
				UnprocessedItems: params.RequestItems,
			}, nil
		},
	}
	h := NewHandler(mock, "TestTable")

	resp, err := h.HandleRequest(context.Background(), makeRequest(http.MethodPost, validWorkoutJSON()))
	// Single workout path: unprocessed after retries returns an error.
	if err == nil {
		t.Fatal("expected error for exhausted retries on single workout")
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, resp.StatusCode)
	}
	// 1 initial + maxRetries retries = 4 total calls.
	if callCount != 4 {
		t.Errorf("expected 4 BatchWriteItem calls (1 + 3 retries), got %d", callCount)
	}
}

func TestBatchUnprocessedItemsReportsFailures(t *testing.T) {
	callCount := 0
	mock := &mockDynamo{
		batchWriteFunc: func(ctx context.Context, params *dynamodb.BatchWriteItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.BatchWriteItemOutput, error) {
			callCount++
			requests := params.RequestItems["TestTable"]
			// Always return the second item as unprocessed (if present).
			var unprocessed []types.WriteRequest
			for _, req := range requests {
				if v, ok := req.PutRequest.Item["workoutId"].(*types.AttributeValueMemberS); ok && v.Value == "wkt-def" {
					unprocessed = append(unprocessed, req)
				}
			}
			if len(unprocessed) > 0 {
				return &dynamodb.BatchWriteItemOutput{
					UnprocessedItems: map[string][]types.WriteRequest{
						"TestTable": unprocessed,
					},
				}, nil
			}
			return &dynamodb.BatchWriteItemOutput{}, nil
		},
	}
	h := NewHandler(mock, "TestTable")

	resp, err := h.HandleRequest(context.Background(), makeRequest(http.MethodPost, validWorkoutArrayJSON()))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Errorf("expected status %d, got %d", http.StatusCreated, resp.StatusCode)
	}

	_, results := parseBatchResponseBody(t, resp.Body)
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	if results[0].Status != "saved" {
		t.Errorf("results[0]: expected 'saved', got %q", results[0].Status)
	}
	if results[1].Status != "error" {
		t.Errorf("results[1]: expected 'error', got %q", results[1].Status)
	}
	if results[1].Error != "failed to save workout" {
		t.Errorf("results[1]: expected error 'failed to save workout', got %q", results[1].Error)
	}
}

// ── Tests: Chunking for >25 items ───────────────────────────────────────────

func TestChunksLargeBatches(t *testing.T) {
	callCount := 0
	var requestSizes []int
	mock := &mockDynamo{
		batchWriteFunc: func(ctx context.Context, params *dynamodb.BatchWriteItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.BatchWriteItemOutput, error) {
			callCount++
			requestSizes = append(requestSizes, len(params.RequestItems["TestTable"]))
			return &dynamodb.BatchWriteItemOutput{}, nil
		},
	}
	h := NewHandler(mock, "TestTable")

	// Build 30 valid workouts.
	workouts := make([]Workout, 30)
	for i := range workouts {
		workouts[i] = Workout{
			UserID:    "user-123",
			SK:        fmt.Sprintf("WORKOUT#2026-03-14T10:00:00.000Z#wkt-%03d", i),
			WorkoutID: fmt.Sprintf("wkt-%03d", i),
			StartedAt: "2026-03-14T10:00:00.000Z",
			CreatedAt: "2026-03-14T10:00:00.000Z",
			UpdatedAt: "2026-03-14T10:00:00.000Z",
		}
	}
	body, _ := json.Marshal(workouts)

	resp, err := h.HandleRequest(context.Background(), makeRequest(http.MethodPost, string(body)))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Errorf("expected status %d, got %d", http.StatusCreated, resp.StatusCode)
	}

	// 30 items should be split into 2 BatchWriteItem calls: 25 + 5.
	if callCount != 2 {
		t.Errorf("expected 2 BatchWriteItem calls, got %d", callCount)
	}
	if len(requestSizes) != 2 || requestSizes[0] != 25 || requestSizes[1] != 5 {
		t.Errorf("expected chunk sizes [25, 5], got %v", requestSizes)
	}

	_, results := parseBatchResponseBody(t, resp.Body)
	if len(results) != 30 {
		t.Errorf("expected 30 results, got %d", len(results))
	}
	for i, r := range results {
		if r.Status != "saved" {
			t.Errorf("results[%d]: expected 'saved', got %q", i, r.Status)
		}
	}
}
