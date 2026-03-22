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

func makeAuthedRequest(userID string, body string) events.APIGatewayProxyRequest {
	return events.APIGatewayProxyRequest{
		Body: body,
		RequestContext: events.APIGatewayProxyRequestContext{
			Authorizer: map[string]interface{}{
				"jwt": map[string]interface{}{
					"claims": map[string]interface{}{
						"sub": userID,
					},
				},
			},
		},
	}
}

func makeUnauthedRequest(body string) events.APIGatewayProxyRequest {
	return events.APIGatewayProxyRequest{
		Body: body,
	}
}

func parseResponseBody(t *testing.T, body string) map[string]any {
	t.Helper()
	var m map[string]any
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

func validExerciseJSON() string {
	ex := SavedExercise{
		SavedExerciseID: "abc123",
		Name:            "Bench Press",
		Note:            "Keep elbows tucked",
		Tags:            []string{"chest", "push"},
		CreatedAt:       "2026-03-01T10:00:00.000Z",
		UpdatedAt:       "2026-03-15T14:30:00.000Z",
	}
	b, _ := json.Marshal(ex)
	return string(b)
}

func validExerciseArrayJSON() string {
	exercises := []SavedExercise{
		{
			SavedExerciseID: "abc123",
			Name:            "Bench Press",
			Note:            "Keep elbows tucked",
			Tags:            []string{"chest", "push"},
			CreatedAt:       "2026-03-01T10:00:00.000Z",
			UpdatedAt:       "2026-03-15T14:30:00.000Z",
		},
	}
	b, _ := json.Marshal(exercises)
	return string(b)
}

func multipleExercisesJSON() string {
	exercises := []SavedExercise{
		{
			SavedExerciseID: "abc123",
			Name:            "Bench Press",
			Note:            "Keep elbows tucked",
			Tags:            []string{"chest", "push"},
			CreatedAt:       "2026-03-01T10:00:00.000Z",
			UpdatedAt:       "2026-03-15T14:30:00.000Z",
		},
		{
			SavedExerciseID: "def456",
			Name:            "Squat",
			Note:            "Full depth",
			Tags:            []string{"legs"},
			CreatedAt:       "2026-03-02T10:00:00.000Z",
			UpdatedAt:       "2026-03-16T14:30:00.000Z",
		},
	}
	b, _ := json.Marshal(exercises)
	return string(b)
}

// ── Tests: Auth ──────────────────────────────────────────────────────────────

func TestRejectsUnauthenticated(t *testing.T) {
	mock := &mockDynamo{}
	h := NewHandler(mock, "TestTable")

	resp, err := h.HandleRequest(context.Background(), makeUnauthedRequest(validExerciseArrayJSON()))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, resp.StatusCode)
	}
	body := parseResponseBody(t, resp.Body)
	if body["error"] != "not authorised" {
		t.Errorf("expected error 'not authorised', got %q", body["error"])
	}
	if mock.calls > 0 {
		t.Error("DynamoDB should not have been called for unauthenticated request")
	}
}

// ── Tests: Request body validation ──────────────────────────────────────────

func TestRejectsEmptyBody(t *testing.T) {
	mock := &mockDynamo{}
	h := NewHandler(mock, "TestTable")

	resp, err := h.HandleRequest(context.Background(), makeAuthedRequest("user-123", ""))
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

func TestRejectsEmptyArray(t *testing.T) {
	mock := &mockDynamo{}
	h := NewHandler(mock, "TestTable")

	resp, err := h.HandleRequest(context.Background(), makeAuthedRequest("user-123", "[]"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, resp.StatusCode)
	}
	body := parseResponseBody(t, resp.Body)
	if body["error"] != "exercise array is empty" {
		t.Errorf("expected 'exercise array is empty', got %q", body["error"])
	}
}

func TestRejectsMissingSavedExerciseId(t *testing.T) {
	mock := &mockDynamo{}
	h := NewHandler(mock, "TestTable")

	payload := `[{"name":"Bench Press","createdAt":"t","updatedAt":"t"}]`
	resp, err := h.HandleRequest(context.Background(), makeAuthedRequest("user-123", payload))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, resp.StatusCode)
	}
	body := parseResponseBody(t, resp.Body)
	expected := "exercises[0]: missing required fields: savedExerciseId"
	if body["error"] != expected {
		t.Errorf("expected error %q, got %q", expected, body["error"])
	}
}

func TestRejectsMissingName(t *testing.T) {
	mock := &mockDynamo{}
	h := NewHandler(mock, "TestTable")

	payload := `[{"savedExerciseId":"abc","createdAt":"t","updatedAt":"t"}]`
	resp, err := h.HandleRequest(context.Background(), makeAuthedRequest("user-123", payload))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, resp.StatusCode)
	}
	body := parseResponseBody(t, resp.Body)
	expected := "exercises[0]: missing required fields: name"
	if body["error"] != expected {
		t.Errorf("expected error %q, got %q", expected, body["error"])
	}
}

func TestRejectsMissingCreatedAt(t *testing.T) {
	mock := &mockDynamo{}
	h := NewHandler(mock, "TestTable")

	payload := `[{"savedExerciseId":"abc","name":"Bench","updatedAt":"t"}]`
	resp, err := h.HandleRequest(context.Background(), makeAuthedRequest("user-123", payload))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, resp.StatusCode)
	}
	body := parseResponseBody(t, resp.Body)
	expected := "exercises[0]: missing required fields: createdAt"
	if body["error"] != expected {
		t.Errorf("expected error %q, got %q", expected, body["error"])
	}
}

func TestRejectsMissingUpdatedAt(t *testing.T) {
	mock := &mockDynamo{}
	h := NewHandler(mock, "TestTable")

	payload := `[{"savedExerciseId":"abc","name":"Bench","createdAt":"t"}]`
	resp, err := h.HandleRequest(context.Background(), makeAuthedRequest("user-123", payload))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, resp.StatusCode)
	}
	body := parseResponseBody(t, resp.Body)
	expected := "exercises[0]: missing required fields: updatedAt"
	if body["error"] != expected {
		t.Errorf("expected error %q, got %q", expected, body["error"])
	}
}

// ── Tests: Successful save ──────────────────────────────────────────────────

func TestSaveSingleExercise(t *testing.T) {
	mock := &mockDynamo{}
	h := NewHandler(mock, "TestTable")

	resp, err := h.HandleRequest(context.Background(), makeAuthedRequest("user-123", validExerciseArrayJSON()))
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
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].SavedExerciseID != "abc123" || results[0].Status != "saved" {
		t.Errorf("results[0]: expected saved abc123, got %+v", results[0])
	}

	if mock.calls != 1 {
		t.Fatalf("expected BatchWriteItem to be called once, got %d", mock.calls)
	}
}

func TestSaveMultipleExercises(t *testing.T) {
	mock := &mockDynamo{}
	h := NewHandler(mock, "TestTable")

	resp, err := h.HandleRequest(context.Background(), makeAuthedRequest("user-123", multipleExercisesJSON()))
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
	if results[0].SavedExerciseID != "abc123" || results[0].Status != "saved" {
		t.Errorf("results[0]: expected saved abc123, got %+v", results[0])
	}
	if results[1].SavedExerciseID != "def456" || results[1].Status != "saved" {
		t.Errorf("results[1]: expected saved def456, got %+v", results[1])
	}

	if mock.calls != 1 {
		t.Errorf("expected 1 BatchWriteItem call, got %d", mock.calls)
	}
	if len(mock.lastInput.RequestItems["TestTable"]) != 2 {
		t.Errorf("expected 2 write requests, got %d", len(mock.lastInput.RequestItems["TestTable"]))
	}
}

// ── Tests: userId and SK generation ─────────────────────────────────────────

func TestSetsUserIdFromJWT(t *testing.T) {
	mock := &mockDynamo{}
	h := NewHandler(mock, "TestTable")

	h.HandleRequest(context.Background(), makeAuthedRequest("user-abc", validExerciseArrayJSON()))

	item := mock.lastInput.RequestItems["TestTable"][0].PutRequest.Item
	userIDVal, ok := item["userId"].(*types.AttributeValueMemberS)
	if !ok {
		t.Fatal("expected userId to be a string attribute")
	}
	if userIDVal.Value != "user-abc" {
		t.Errorf("expected userId 'user-abc', got %q", userIDVal.Value)
	}
}

func TestGeneratesCorrectSK(t *testing.T) {
	mock := &mockDynamo{}
	h := NewHandler(mock, "TestTable")

	h.HandleRequest(context.Background(), makeAuthedRequest("user-123", validExerciseArrayJSON()))

	item := mock.lastInput.RequestItems["TestTable"][0].PutRequest.Item
	skVal, ok := item["sk"].(*types.AttributeValueMemberS)
	if !ok {
		t.Fatal("expected sk to be a string attribute")
	}
	if skVal.Value != "EXERCISE#abc123" {
		t.Errorf("expected sk 'EXERCISE#abc123', got %q", skVal.Value)
	}
}

// ── Tests: DynamoDB error ─────────────────────────────────────────────────────

func TestDynamoDBBatchWriteError(t *testing.T) {
	mock := &mockDynamo{
		batchWriteFunc: func(ctx context.Context, params *dynamodb.BatchWriteItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.BatchWriteItemOutput, error) {
			return nil, errors.New("service unavailable")
		},
	}
	h := NewHandler(mock, "TestTable")

	resp, err := h.HandleRequest(context.Background(), makeAuthedRequest("user-123", validExerciseArrayJSON()))
	if err == nil {
		t.Fatal("expected error to be returned")
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, resp.StatusCode)
	}
}

func TestReportsFailedExercises(t *testing.T) {
	callCount := 0
	mock := &mockDynamo{
		batchWriteFunc: func(ctx context.Context, params *dynamodb.BatchWriteItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.BatchWriteItemOutput, error) {
			callCount++
			requests := params.RequestItems["TestTable"]
			// Always return the second item as unprocessed (if present).
			var unprocessed []types.WriteRequest
			for _, req := range requests {
				if v, ok := req.PutRequest.Item["savedExerciseId"].(*types.AttributeValueMemberS); ok && v.Value == "def456" {
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

	resp, err := h.HandleRequest(context.Background(), makeAuthedRequest("user-123", multipleExercisesJSON()))
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
	if results[1].Error != "failed to save exercise" {
		t.Errorf("results[1]: expected error 'failed to save exercise', got %q", results[1].Error)
	}
}

// ── Tests: Table name ────────────────────────────────────────────────────────

func TestTableNameIsPassedCorrectly(t *testing.T) {
	mock := &mockDynamo{}
	h := NewHandler(mock, "MyCustomTable")

	h.HandleRequest(context.Background(), makeAuthedRequest("user-123", validExerciseArrayJSON()))

	if _, ok := mock.lastInput.RequestItems["MyCustomTable"]; !ok {
		t.Errorf("expected request items for table 'MyCustomTable', got keys: %v", mock.lastInput.RequestItems)
	}
}

// ── Tests: Response format ──────────────────────────────────────────────────

func TestResponseHasJSONContentType(t *testing.T) {
	mock := &mockDynamo{}
	h := NewHandler(mock, "TestTable")

	resp, _ := h.HandleRequest(context.Background(), makeAuthedRequest("user-123", validExerciseArrayJSON()))
	ct := resp.Headers["Content-Type"]
	if ct != "application/json" {
		t.Errorf("expected Content-Type 'application/json', got %q", ct)
	}
}

// ── Tests: Single object (not array) input ──────────────────────────────────

func TestSaveSingleExerciseObject(t *testing.T) {
	mock := &mockDynamo{}
	h := NewHandler(mock, "TestTable")

	resp, err := h.HandleRequest(context.Background(), makeAuthedRequest("user-123", validExerciseJSON()))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Errorf("expected status %d, got %d", http.StatusCreated, resp.StatusCode)
	}

	body := parseResponseBody(t, resp.Body)
	if body["message"] != "exercise saved" {
		t.Errorf("expected message 'exercise saved', got %q", body["message"])
	}
	if body["savedExerciseId"] != "abc123" {
		t.Errorf("expected savedExerciseId 'abc123', got %q", body["savedExerciseId"])
	}

	if mock.calls != 1 {
		t.Fatalf("expected BatchWriteItem to be called once, got %d", mock.calls)
	}
}

func TestSingleObjectSetsUserIdFromJWT(t *testing.T) {
	mock := &mockDynamo{}
	h := NewHandler(mock, "TestTable")

	h.HandleRequest(context.Background(), makeAuthedRequest("user-abc", validExerciseJSON()))

	item := mock.lastInput.RequestItems["TestTable"][0].PutRequest.Item
	userIDVal, ok := item["userId"].(*types.AttributeValueMemberS)
	if !ok {
		t.Fatal("expected userId to be a string attribute")
	}
	if userIDVal.Value != "user-abc" {
		t.Errorf("expected userId 'user-abc', got %q", userIDVal.Value)
	}
}

func TestSingleObjectGeneratesCorrectSK(t *testing.T) {
	mock := &mockDynamo{}
	h := NewHandler(mock, "TestTable")

	h.HandleRequest(context.Background(), makeAuthedRequest("user-123", validExerciseJSON()))

	item := mock.lastInput.RequestItems["TestTable"][0].PutRequest.Item
	skVal, ok := item["sk"].(*types.AttributeValueMemberS)
	if !ok {
		t.Fatal("expected sk to be a string attribute")
	}
	if skVal.Value != "EXERCISE#abc123" {
		t.Errorf("expected sk 'EXERCISE#abc123', got %q", skVal.Value)
	}
}

func TestSingleObjectRejectsMissingFields(t *testing.T) {
	mock := &mockDynamo{}
	h := NewHandler(mock, "TestTable")

	payload := `{"name":"Bench Press","createdAt":"t","updatedAt":"t"}`
	resp, err := h.HandleRequest(context.Background(), makeAuthedRequest("user-123", payload))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, resp.StatusCode)
	}
	body := parseResponseBody(t, resp.Body)
	expected := "missing required fields: savedExerciseId"
	if body["error"] != expected {
		t.Errorf("expected error %q, got %q", expected, body["error"])
	}
}

func TestSingleObjectDynamoDBError(t *testing.T) {
	mock := &mockDynamo{
		batchWriteFunc: func(ctx context.Context, params *dynamodb.BatchWriteItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.BatchWriteItemOutput, error) {
			return nil, errors.New("service unavailable")
		},
	}
	h := NewHandler(mock, "TestTable")

	resp, err := h.HandleRequest(context.Background(), makeAuthedRequest("user-123", validExerciseJSON()))
	if err == nil {
		t.Fatal("expected error to be returned")
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, resp.StatusCode)
	}
}

func TestSingleObjectUnprocessedReturnsError(t *testing.T) {
	mock := &mockDynamo{
		batchWriteFunc: func(ctx context.Context, params *dynamodb.BatchWriteItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.BatchWriteItemOutput, error) {
			// Always return items as unprocessed.
			return &dynamodb.BatchWriteItemOutput{
				UnprocessedItems: params.RequestItems,
			}, nil
		},
	}
	h := NewHandler(mock, "TestTable")

	resp, err := h.HandleRequest(context.Background(), makeAuthedRequest("user-123", validExerciseJSON()))
	if err == nil {
		t.Fatal("expected error for exhausted retries on single exercise")
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, resp.StatusCode)
	}
}

func TestSingleObjectStillWorksWithArrayInput(t *testing.T) {
	mock := &mockDynamo{}
	h := NewHandler(mock, "TestTable")

	// Array with one item should still return batch format.
	resp, err := h.HandleRequest(context.Background(), makeAuthedRequest("user-123", validExerciseArrayJSON()))
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
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].SavedExerciseID != "abc123" || results[0].Status != "saved" {
		t.Errorf("results[0]: expected saved abc123, got %+v", results[0])
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

	// Build 30 valid exercises.
	exercises := make([]SavedExercise, 30)
	for i := range exercises {
		exercises[i] = SavedExercise{
			SavedExerciseID: fmt.Sprintf("ex-%03d", i),
			Name:            fmt.Sprintf("Exercise %d", i),
			CreatedAt:       "2026-03-14T10:00:00.000Z",
			UpdatedAt:       "2026-03-14T10:00:00.000Z",
		}
	}
	body, _ := json.Marshal(exercises)

	resp, err := h.HandleRequest(context.Background(), makeAuthedRequest("user-123", string(body)))
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
