package main

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// mockDynamo implements DynamoUpdater for testing.
type mockDynamo struct {
	updateFunc func(ctx context.Context, params *dynamodb.UpdateItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.UpdateItemOutput, error)
	calls      int
	lastInput  *dynamodb.UpdateItemInput
}

func (m *mockDynamo) UpdateItem(ctx context.Context, params *dynamodb.UpdateItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.UpdateItemOutput, error) {
	m.calls++
	m.lastInput = params
	if m.updateFunc != nil {
		return m.updateFunc(ctx, params, optFns...)
	}
	return &dynamodb.UpdateItemOutput{}, nil
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

func fixedTime() time.Time {
	return time.Date(2026, 3, 20, 12, 0, 0, 0, time.UTC)
}

// ── Tests: Auth ──────────────────────────────────────────────────────────────

func TestRejectsUnauthenticated(t *testing.T) {
	mock := &mockDynamo{}
	h := NewHandler(mock, "TestTable")

	resp, err := h.HandleRequest(context.Background(), makeUnauthedRequest(`{"savedExerciseId":"abc","archive":true}`))
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

func TestRejectsMissingSavedExerciseId(t *testing.T) {
	mock := &mockDynamo{}
	h := NewHandler(mock, "TestTable")

	resp, err := h.HandleRequest(context.Background(), makeAuthedRequest("user-123", `{"archive":true}`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, resp.StatusCode)
	}
	body := parseResponseBody(t, resp.Body)
	if body["error"] != "missing required field: savedExerciseId" {
		t.Errorf("expected 'missing required field: savedExerciseId', got %q", body["error"])
	}
}

// ── Tests: Archive (archive=true) ───────────────────────────────────────────

func TestArchiveExercise(t *testing.T) {
	mock := &mockDynamo{}
	h := NewHandler(mock, "TestTable")
	h.now = fixedTime

	resp, err := h.HandleRequest(context.Background(), makeAuthedRequest("user-123", `{"savedExerciseId":"abc123","archive":true}`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}
	body := parseResponseBody(t, resp.Body)
	if body["message"] != "exercise archived" {
		t.Errorf("expected message 'exercise archived', got %q", body["message"])
	}

	// Verify UpdateExpression sets archivedAt and updatedAt.
	if mock.lastInput.UpdateExpression == nil {
		t.Fatal("expected UpdateExpression to be set")
	}
	expr := *mock.lastInput.UpdateExpression
	if expr != "SET archivedAt = :archivedAt, updatedAt = :updatedAt" {
		t.Errorf("unexpected UpdateExpression: %q", expr)
	}

	// Verify archivedAt value is the fixed time.
	archivedAtVal, ok := mock.lastInput.ExpressionAttributeValues[":archivedAt"].(*types.AttributeValueMemberS)
	if !ok {
		t.Fatal("expected :archivedAt to be a string attribute")
	}
	expectedTime := fixedTime().UTC().Format(time.RFC3339Nano)
	if archivedAtVal.Value != expectedTime {
		t.Errorf("expected archivedAt %q, got %q", expectedTime, archivedAtVal.Value)
	}

	// Verify updatedAt value is the fixed time.
	updatedAtVal, ok := mock.lastInput.ExpressionAttributeValues[":updatedAt"].(*types.AttributeValueMemberS)
	if !ok {
		t.Fatal("expected :updatedAt to be a string attribute")
	}
	if updatedAtVal.Value != expectedTime {
		t.Errorf("expected updatedAt %q, got %q", expectedTime, updatedAtVal.Value)
	}
}

// ── Tests: Restore (archive=false) ──────────────────────────────────────────

func TestRestoreExercise(t *testing.T) {
	mock := &mockDynamo{}
	h := NewHandler(mock, "TestTable")
	h.now = fixedTime

	resp, err := h.HandleRequest(context.Background(), makeAuthedRequest("user-123", `{"savedExerciseId":"abc123","archive":false}`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}
	body := parseResponseBody(t, resp.Body)
	if body["message"] != "exercise restored" {
		t.Errorf("expected message 'exercise restored', got %q", body["message"])
	}

	// Verify UpdateExpression removes archivedAt and sets updatedAt.
	if mock.lastInput.UpdateExpression == nil {
		t.Fatal("expected UpdateExpression to be set")
	}
	expr := *mock.lastInput.UpdateExpression
	if expr != "SET updatedAt = :updatedAt REMOVE archivedAt" {
		t.Errorf("unexpected UpdateExpression: %q", expr)
	}

	// Verify there is no :archivedAt in expression values.
	if _, ok := mock.lastInput.ExpressionAttributeValues[":archivedAt"]; ok {
		t.Error("expected no :archivedAt expression value for restore")
	}

	// Verify updatedAt value is the fixed time.
	updatedAtVal, ok := mock.lastInput.ExpressionAttributeValues[":updatedAt"].(*types.AttributeValueMemberS)
	if !ok {
		t.Fatal("expected :updatedAt to be a string attribute")
	}
	expectedTime := fixedTime().UTC().Format(time.RFC3339Nano)
	if updatedAtVal.Value != expectedTime {
		t.Errorf("expected updatedAt %q, got %q", expectedTime, updatedAtVal.Value)
	}
}

// ── Tests: Key correctness ──────────────────────────────────────────────────

func TestUsesCorrectKey(t *testing.T) {
	mock := &mockDynamo{}
	h := NewHandler(mock, "TestTable")

	h.HandleRequest(context.Background(), makeAuthedRequest("user-xyz", `{"savedExerciseId":"abc123","archive":true}`))

	key := mock.lastInput.Key
	userIDVal, ok := key["userId"].(*types.AttributeValueMemberS)
	if !ok {
		t.Fatal("expected userId in key to be a string attribute")
	}
	if userIDVal.Value != "user-xyz" {
		t.Errorf("expected userId 'user-xyz', got %q", userIDVal.Value)
	}

	skVal, ok := key["sk"].(*types.AttributeValueMemberS)
	if !ok {
		t.Fatal("expected sk in key to be a string attribute")
	}
	if skVal.Value != "EXERCISE#abc123" {
		t.Errorf("expected sk 'EXERCISE#abc123', got %q", skVal.Value)
	}
}

// ── Tests: Condition expression ─────────────────────────────────────────────

func TestConditionExpressionChecksItemExists(t *testing.T) {
	mock := &mockDynamo{}
	h := NewHandler(mock, "TestTable")

	h.HandleRequest(context.Background(), makeAuthedRequest("user-123", `{"savedExerciseId":"abc123","archive":true}`))

	if mock.lastInput.ConditionExpression == nil {
		t.Fatal("expected ConditionExpression to be set")
	}
	if *mock.lastInput.ConditionExpression != "attribute_exists(sk)" {
		t.Errorf("expected ConditionExpression 'attribute_exists(sk)', got %q", *mock.lastInput.ConditionExpression)
	}
}

// ── Tests: 404 when item not found ──────────────────────────────────────────

func TestReturns404WhenItemNotFound(t *testing.T) {
	mock := &mockDynamo{
		updateFunc: func(ctx context.Context, params *dynamodb.UpdateItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.UpdateItemOutput, error) {
			return nil, &types.ConditionalCheckFailedException{Message: stringPtr("The conditional request failed")}
		},
	}
	h := NewHandler(mock, "TestTable")

	resp, err := h.HandleRequest(context.Background(), makeAuthedRequest("user-123", `{"savedExerciseId":"nonexistent","archive":true}`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, resp.StatusCode)
	}
	body := parseResponseBody(t, resp.Body)
	if body["error"] != "exercise not found" {
		t.Errorf("expected error 'exercise not found', got %q", body["error"])
	}
}

func stringPtr(s string) *string {
	return &s
}

// ── Tests: DynamoDB error ─────────────────────────────────────────────────────

func TestDynamoDBUpdateError(t *testing.T) {
	mock := &mockDynamo{
		updateFunc: func(ctx context.Context, params *dynamodb.UpdateItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.UpdateItemOutput, error) {
			return nil, errors.New("service unavailable")
		},
	}
	h := NewHandler(mock, "TestTable")

	resp, err := h.HandleRequest(context.Background(), makeAuthedRequest("user-123", `{"savedExerciseId":"abc123","archive":true}`))
	if err == nil {
		t.Fatal("expected error to be returned from HandleRequest")
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, resp.StatusCode)
	}
	body := parseResponseBody(t, resp.Body)
	if body["error"] != "failed to update exercise" {
		t.Errorf("expected 'failed to update exercise', got %q", body["error"])
	}
}

// ── Tests: Response format ──────────────────────────────────────────────────

func TestResponseHasJSONContentType(t *testing.T) {
	mock := &mockDynamo{}
	h := NewHandler(mock, "TestTable")

	resp, _ := h.HandleRequest(context.Background(), makeAuthedRequest("user-123", `{"savedExerciseId":"abc123","archive":true}`))
	if resp.Headers["Content-Type"] != "application/json" {
		t.Errorf("expected Content-Type 'application/json', got %q", resp.Headers["Content-Type"])
	}
}

func TestErrorResponseHasJSONContentType(t *testing.T) {
	mock := &mockDynamo{}
	h := NewHandler(mock, "TestTable")

	resp, _ := h.HandleRequest(context.Background(), makeUnauthedRequest(""))
	if resp.Headers["Content-Type"] != "application/json" {
		t.Errorf("expected Content-Type 'application/json', got %q", resp.Headers["Content-Type"])
	}
}
