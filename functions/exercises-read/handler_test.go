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

// mockDynamo implements DynamoQuerier for testing.
type mockDynamo struct {
	queryFunc func(ctx context.Context, params *dynamodb.QueryInput, optFns ...func(*dynamodb.Options)) (*dynamodb.QueryOutput, error)
	calls     int
	lastInput *dynamodb.QueryInput
}

func (m *mockDynamo) Query(ctx context.Context, params *dynamodb.QueryInput, optFns ...func(*dynamodb.Options)) (*dynamodb.QueryOutput, error) {
	m.calls++
	m.lastInput = params
	if m.queryFunc != nil {
		return m.queryFunc(ctx, params, optFns...)
	}
	return &dynamodb.QueryOutput{Items: []map[string]types.AttributeValue{}}, nil
}

// makeAuthedRequest builds an APIGatewayProxyRequest with a fake Cognito JWT
// authorizer context for the given user ID.
func makeAuthedRequest(userID string) events.APIGatewayProxyRequest {
	return events.APIGatewayProxyRequest{
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

func makeUnauthedRequest() events.APIGatewayProxyRequest {
	return events.APIGatewayProxyRequest{}
}

func parseResponseBody(t *testing.T, body string) map[string]any {
	t.Helper()
	var m map[string]any
	if err := json.Unmarshal([]byte(body), &m); err != nil {
		t.Fatalf("failed to parse response body: %v", err)
	}
	return m
}

// buildExerciseItems constructs n DynamoDB attribute maps that look like SavedExercise items.
func buildExerciseItems(userID string, n int) []map[string]types.AttributeValue {
	items := make([]map[string]types.AttributeValue, n)
	for i := 0; i < n; i++ {
		id := fmt.Sprintf("ex-%03d", i)
		items[i] = map[string]types.AttributeValue{
			"userId":          &types.AttributeValueMemberS{Value: userID},
			"sk":              &types.AttributeValueMemberS{Value: fmt.Sprintf("EXERCISE#%s", id)},
			"savedExerciseId": &types.AttributeValueMemberS{Value: id},
			"name":            &types.AttributeValueMemberS{Value: fmt.Sprintf("Exercise %d", i)},
			"note":            &types.AttributeValueMemberS{Value: ""},
			"tags":            &types.AttributeValueMemberL{Value: []types.AttributeValue{}},
			"createdAt":       &types.AttributeValueMemberS{Value: fmt.Sprintf("2026-03-%02dT10:00:00.000Z", i+1)},
			"updatedAt":       &types.AttributeValueMemberS{Value: fmt.Sprintf("2026-03-%02dT10:00:00.000Z", i+1)},
		}
	}
	return items
}

// ── Tests: Auth ──────────────────────────────────────────────────────────────

func TestRejectsUnauthenticated(t *testing.T) {
	mock := &mockDynamo{}
	h := NewHandler(mock, "TestTable")

	resp, err := h.HandleRequest(context.Background(), makeUnauthedRequest())
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

func TestRejectsMissingJWTClaims(t *testing.T) {
	mock := &mockDynamo{}
	h := NewHandler(mock, "TestTable")

	req := events.APIGatewayProxyRequest{
		RequestContext: events.APIGatewayProxyRequestContext{
			Authorizer: map[string]interface{}{
				"jwt": map[string]interface{}{
					// no "claims" key
				},
			},
		},
	}

	resp, err := h.HandleRequest(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, resp.StatusCode)
	}
}

func TestRejectsMissingTableName(t *testing.T) {
	mock := &mockDynamo{}
	h := NewHandler(mock, "")

	resp, err := h.HandleRequest(context.Background(), makeAuthedRequest("user-123"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, resp.StatusCode)
	}
	body := parseResponseBody(t, resp.Body)
	if body["error"] != "table name not configured" {
		t.Errorf("expected error 'table name not configured', got %q", body["error"])
	}
}

// ── Tests: Empty result ───────────────────────────────────────────────────────

func TestReturnsEmptyArrayWhenNoExercises(t *testing.T) {
	mock := &mockDynamo{
		queryFunc: func(ctx context.Context, params *dynamodb.QueryInput, optFns ...func(*dynamodb.Options)) (*dynamodb.QueryOutput, error) {
			return &dynamodb.QueryOutput{Items: []map[string]types.AttributeValue{}}, nil
		},
	}
	h := NewHandler(mock, "TestTable")

	resp, err := h.HandleRequest(context.Background(), makeAuthedRequest("user-123"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}

	body := parseResponseBody(t, resp.Body)
	exercisesRaw, ok := body["exercises"].([]any)
	if !ok {
		t.Fatalf("expected 'exercises' array in response, got %T", body["exercises"])
	}
	if len(exercisesRaw) != 0 {
		t.Errorf("expected 0 exercises, got %d", len(exercisesRaw))
	}
}

// ── Tests: Successful read ───────────────────────────────────────────────────

func TestReturnsExercisesCorrectly(t *testing.T) {
	archived := "2026-03-10T12:00:00.000Z"
	items := []map[string]types.AttributeValue{
		{
			"userId":          &types.AttributeValueMemberS{Value: "user-123"},
			"sk":              &types.AttributeValueMemberS{Value: "EXERCISE#abc123"},
			"savedExerciseId": &types.AttributeValueMemberS{Value: "abc123"},
			"name":            &types.AttributeValueMemberS{Value: "Bench Press"},
			"note":            &types.AttributeValueMemberS{Value: "Keep elbows tucked"},
			"tags": &types.AttributeValueMemberL{
				Value: []types.AttributeValue{
					&types.AttributeValueMemberS{Value: "chest"},
					&types.AttributeValueMemberS{Value: "push"},
				},
			},
			"archivedAt": &types.AttributeValueMemberS{Value: archived},
			"createdAt":  &types.AttributeValueMemberS{Value: "2026-03-01T10:00:00.000Z"},
			"updatedAt":  &types.AttributeValueMemberS{Value: "2026-03-15T14:30:00.000Z"},
		},
	}

	mock := &mockDynamo{
		queryFunc: func(ctx context.Context, params *dynamodb.QueryInput, optFns ...func(*dynamodb.Options)) (*dynamodb.QueryOutput, error) {
			return &dynamodb.QueryOutput{Items: items}, nil
		},
	}
	h := NewHandler(mock, "TestTable")

	resp, err := h.HandleRequest(context.Background(), makeAuthedRequest("user-123"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}

	body := parseResponseBody(t, resp.Body)
	exercisesRaw, ok := body["exercises"].([]any)
	if !ok || len(exercisesRaw) != 1 {
		t.Fatalf("expected 1 exercise in response")
	}

	ex := exercisesRaw[0].(map[string]any)
	if ex["savedExerciseId"] != "abc123" {
		t.Errorf("expected savedExerciseId 'abc123', got %q", ex["savedExerciseId"])
	}
	if ex["name"] != "Bench Press" {
		t.Errorf("expected name 'Bench Press', got %q", ex["name"])
	}
	if ex["note"] != "Keep elbows tucked" {
		t.Errorf("expected note 'Keep elbows tucked', got %q", ex["note"])
	}
	tags, ok := ex["tags"].([]any)
	if !ok || len(tags) != 2 {
		t.Fatalf("expected 2 tags, got %v", ex["tags"])
	}
	if tags[0] != "chest" || tags[1] != "push" {
		t.Errorf("expected tags [chest, push], got %v", tags)
	}
	if ex["archivedAt"] != archived {
		t.Errorf("expected archivedAt %q, got %q", archived, ex["archivedAt"])
	}
	if ex["createdAt"] != "2026-03-01T10:00:00.000Z" {
		t.Errorf("expected createdAt '2026-03-01T10:00:00.000Z', got %q", ex["createdAt"])
	}
	if ex["updatedAt"] != "2026-03-15T14:30:00.000Z" {
		t.Errorf("expected updatedAt '2026-03-15T14:30:00.000Z', got %q", ex["updatedAt"])
	}
}

func TestReturnsMultipleExercises(t *testing.T) {
	items := buildExerciseItems("user-123", 5)
	mock := &mockDynamo{
		queryFunc: func(ctx context.Context, params *dynamodb.QueryInput, optFns ...func(*dynamodb.Options)) (*dynamodb.QueryOutput, error) {
			return &dynamodb.QueryOutput{Items: items}, nil
		},
	}
	h := NewHandler(mock, "TestTable")

	resp, err := h.HandleRequest(context.Background(), makeAuthedRequest("user-123"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}

	body := parseResponseBody(t, resp.Body)
	exercisesRaw, ok := body["exercises"].([]any)
	if !ok {
		t.Fatalf("expected 'exercises' array in response, got %T", body["exercises"])
	}
	if len(exercisesRaw) != 5 {
		t.Errorf("expected 5 exercises, got %d", len(exercisesRaw))
	}
}

// ── Tests: DynamoDB error ─────────────────────────────────────────────────────

func TestDynamoDBQueryError(t *testing.T) {
	mock := &mockDynamo{
		queryFunc: func(ctx context.Context, params *dynamodb.QueryInput, optFns ...func(*dynamodb.Options)) (*dynamodb.QueryOutput, error) {
			return nil, errors.New("service unavailable")
		},
	}
	h := NewHandler(mock, "TestTable")

	resp, err := h.HandleRequest(context.Background(), makeAuthedRequest("user-123"))
	if err == nil {
		t.Fatal("expected error to be returned from HandleRequest")
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, resp.StatusCode)
	}
	body := parseResponseBody(t, resp.Body)
	if body["error"] != "failed to read exercises" {
		t.Errorf("expected 'failed to read exercises', got %q", body["error"])
	}
}

// ── Tests: DynamoDB query parameters ─────────────────────────────────────────

func TestQueryUsesCorrectTableName(t *testing.T) {
	mock := &mockDynamo{}
	h := NewHandler(mock, "MyExercisesTable")

	h.HandleRequest(context.Background(), makeAuthedRequest("user-123"))

	if *mock.lastInput.TableName != "MyExercisesTable" {
		t.Errorf("expected TableName 'MyExercisesTable', got %q", *mock.lastInput.TableName)
	}
}

func TestQueryUsesExercisePrefix(t *testing.T) {
	mock := &mockDynamo{}
	h := NewHandler(mock, "TestTable")

	h.HandleRequest(context.Background(), makeAuthedRequest("user-123"))

	prefixAttr, ok := mock.lastInput.ExpressionAttributeValues[":prefix"]
	if !ok {
		t.Fatal("expected :prefix expression attribute value")
	}
	prefixVal, ok := prefixAttr.(*types.AttributeValueMemberS)
	if !ok {
		t.Fatal("expected :prefix to be a string attribute")
	}
	if prefixVal.Value != "EXERCISE#" {
		t.Errorf("expected :prefix = 'EXERCISE#', got %q", prefixVal.Value)
	}
}

func TestQueryScopedToCurrentUser(t *testing.T) {
	mock := &mockDynamo{}
	h := NewHandler(mock, "TestTable")

	h.HandleRequest(context.Background(), makeAuthedRequest("user-xyz"))

	uidAttr, ok := mock.lastInput.ExpressionAttributeValues[":uid"]
	if !ok {
		t.Fatal("expected :uid expression attribute value")
	}
	uidVal, ok := uidAttr.(*types.AttributeValueMemberS)
	if !ok {
		t.Fatal("expected :uid to be a string attribute")
	}
	if uidVal.Value != "user-xyz" {
		t.Errorf("expected :uid = 'user-xyz', got %q", uidVal.Value)
	}
}

// ── Tests: Response format ────────────────────────────────────────────────────

func TestResponseHasJSONContentType(t *testing.T) {
	mock := &mockDynamo{}
	h := NewHandler(mock, "TestTable")

	resp, _ := h.HandleRequest(context.Background(), makeAuthedRequest("user-123"))
	if resp.Headers["Content-Type"] != "application/json" {
		t.Errorf("expected Content-Type 'application/json', got %q", resp.Headers["Content-Type"])
	}
}

func TestErrorResponseHasJSONContentType(t *testing.T) {
	mock := &mockDynamo{}
	h := NewHandler(mock, "TestTable")

	resp, _ := h.HandleRequest(context.Background(), makeUnauthedRequest())
	if resp.Headers["Content-Type"] != "application/json" {
		t.Errorf("expected Content-Type 'application/json', got %q", resp.Headers["Content-Type"])
	}
}
