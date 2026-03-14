package main

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

type mockDynamo struct {
	putItemFunc func(ctx context.Context, params *dynamodb.PutItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error)
	called      bool
	lastInput   *dynamodb.PutItemInput
}

func (m *mockDynamo) PutItem(ctx context.Context, params *dynamodb.PutItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error) {
	m.called = true
	m.lastInput = params
	if m.putItemFunc != nil {
		return m.putItemFunc(ctx, params, optFns...)
	}
	return &dynamodb.PutItemOutput{}, nil
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

	if mock.called {
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

// ── Tests: Successful save ──────────────────────────────────────────────────

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

	if !mock.called {
		t.Fatal("expected DynamoDB PutItem to be called")
	}
	if *mock.lastInput.TableName != "TestTable" {
		t.Errorf("expected table name 'TestTable', got %q", *mock.lastInput.TableName)
	}

	// Verify key attributes are in the DynamoDB item.
	item := mock.lastInput.Item
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

	if !mock.called {
		t.Fatal("expected DynamoDB PutItem to be called")
	}
}

// ── Tests: Nullable fields ──────────────────────────────────────────────────

func TestNullableFieldsAreAccepted(t *testing.T) {
	// templateId, completedAt, and bodyWeight can be null.
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
	// Sets can have null reps and null weight.
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

// ── Tests: DynamoDB failure ─────────────────────────────────────────────────

func TestDynamoDBPutItemError(t *testing.T) {
	mock := &mockDynamo{
		putItemFunc: func(ctx context.Context, params *dynamodb.PutItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error) {
			return nil, errors.New("service unavailable")
		},
	}
	h := NewHandler(mock, "TestTable")

	resp, err := h.HandleRequest(context.Background(), makeRequest(http.MethodPost, validWorkoutJSON()))
	// The handler returns the DynamoDB error as the Lambda error for observability.
	if err == nil {
		t.Fatal("expected error to be returned")
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, resp.StatusCode)
	}
	body := parseResponseBody(t, resp.Body)
	if body["error"] != "failed to save workout" {
		t.Errorf("expected 'failed to save workout', got %q", body["error"])
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

	item := mock.lastInput.Item
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

	if *mock.lastInput.TableName != tableName {
		t.Errorf("expected table name %q, got %q", tableName, *mock.lastInput.TableName)
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
