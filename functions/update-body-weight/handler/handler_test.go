package handler

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// test: Auth
func TestUnhappyUnauthenticated(t *testing.T) {
	mock := &mockDynamo{}
	h := NewHandler(mock, "TestTable")

	resp, err := h.HandleRequest(context.Background(), makeRequestWithContext(
		http.MethodPatch,
		validRequestBody,
		&events.APIGatewayProxyRequestContext{},
	))

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, resp.StatusCode)
	}

	body := parseErrorResponseBody(t, resp.Body)
	if body["error"] != "not authorised" {
		t.Errorf("expected error 'not authorised', got %q", body["error"])
	}
}

func TestUnhappyMissingClaimsSub(t *testing.T) {
	mock := &mockDynamo{}
	h := NewHandler(mock, "TestTable")

	req := events.APIGatewayProxyRequest{
		RequestContext: events.APIGatewayProxyRequestContext{
			Authorizer: map[string]any{
				"jwt": map[string]any{
					"claims": map[string]any{},
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


func TestUnhappyMissingJWTClaims(t *testing.T) {
	mock := &mockDynamo{}
	h := NewHandler(mock, "TestTable")

	req := events.APIGatewayProxyRequest{
		RequestContext: events.APIGatewayProxyRequestContext{
			Authorizer: map[string]any{
				"jwt": map[string]any{},
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

func TestUnhappyMissingJWTKey(t *testing.T) {
	mock := &mockDynamo{}
	h := NewHandler(mock, "TestTable")

	req := events.APIGatewayProxyRequest{
		RequestContext: events.APIGatewayProxyRequestContext{
			Authorizer: map[string]any{},
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

// TEST: Missing table name
func TestUnhappyMissingTableName(t *testing.T) {
	mock := &mockDynamo{}
	h := NewHandler(mock, "")

	resp, err := h.HandleRequest(context.Background(), makeRequest(
		http.MethodPatch,
		validRequestBody,
	))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, resp.StatusCode)
	}
	body := parseErrorResponseBody(t, resp.Body)
	if body["error"] != "table name not configured" {
		t.Errorf("expected error 'table name not configured', got %q", body["error"])
	}
}

func TestUnhappyInvalidBody(t *testing.T) {
	mock := &mockDynamo{}
	h := NewHandler(mock, "mockTable")

	resp, err := h.HandleRequest(context.Background(), makeRequestWithString(
		http.MethodPatch,
		"",
	))

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, resp.StatusCode)
	}

	body := parseErrorResponseBody(t, resp.Body)
	expectedError := "request body is empty"
	if body["error"] != expectedError {
		t.Errorf("expected error %q, got %q", expectedError, body["error"])
	}
}

func TestUnhappyInvalidJSONBody(t *testing.T) {
	mock := &mockDynamo{}
	h := NewHandler(mock, "mockTable")

	resp, err := h.HandleRequest(context.Background(), makeRequestWithString(
		http.MethodPatch,
		"invalidJson",
	))

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, resp.StatusCode)
	}

	body := parseErrorResponseBody(t, resp.Body)
	expectedError := "invalid JSON"
	if body["error"] != expectedError {
		t.Errorf("expected error %q, got %q", expectedError, body["error"])
	}
}

func TestUnhappyInvalidRequestBody(t *testing.T) {
	mock := &mockDynamo{}
	h := NewHandler(mock, "mockTable")

	resp, err := h.HandleRequest(context.Background(), makeRequestWithString(
		http.MethodPatch,
		"{ \"foo\": \"bar\" }",
	))

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, resp.StatusCode)
	}

	body := parseErrorResponseBody(t, resp.Body)
	expectedError := "invalid request body"
	if body["error"] != expectedError {
		t.Errorf("expected error %q, got %q", expectedError, body["error"])
	}
}

func TestUnhappyDynamoDBQueryError(t *testing.T) {
	mock := &mockDynamo{}
	mock.MockQuery = func(
		ctx context.Context,
		params *dynamodb.QueryInput,
		optFns ...func(*dynamodb.Options),
	) (*dynamodb.QueryOutput, error) {
		return nil, errors.New("Ooopsy!")
	}
	h := NewHandler(mock, "mockTable")

	resp, err := h.HandleRequest(context.Background(), makeRequest(
		http.MethodPatch,
		validRequestBody,
	))

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, resp.StatusCode)
	}

	body := parseErrorResponseBody(t, resp.Body)
	expectedError := "failed to save workouts"
	if body["error"] != expectedError {
		t.Errorf("expected error %q, got %q", expectedError, body["error"])
	}
}

func TestUnhappyDynamoDBQueryReturnsMoreThanOneValue(t *testing.T) {
	mock := &mockDynamo{}
	mock.MockQuery = func(
		ctx context.Context,
		params *dynamodb.QueryInput,
		optFns ...func(*dynamodb.Options),
	) (*dynamodb.QueryOutput, error) {
		return &dynamodb.QueryOutput{
			Count: 10,
		}, nil
	}
	h := NewHandler(mock, "mockTable")

	resp, err := h.HandleRequest(context.Background(), makeRequest(
		http.MethodPatch,
		validRequestBody,
	))

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, resp.StatusCode)
	}

	body := parseErrorResponseBody(t, resp.Body)
	expectedError := "failed to save workouts"
	if body["error"] != expectedError {
		t.Errorf("expected error %q, got %q", expectedError, body["error"])
	}
}

func TestUnhappyDynamoDBQueryUnmarshalFails(t *testing.T) {
	mock := &mockDynamo{}
	mock.MockQuery = func(
		ctx context.Context,
		params *dynamodb.QueryInput,
		optFns ...func(*dynamodb.Options),
	) (*dynamodb.QueryOutput, error) {
		return &dynamodb.QueryOutput{
			Items: []map[string]types.AttributeValue{
				{
					"userId": &types.AttributeValueMemberBOOL{ Value: true },
					"sk": &types.AttributeValueMemberN{ Value: "10" },
				},
			},
			Count: 1,
		}, nil
	}
	h := NewHandler(mock, "mockTable")

	resp, err := h.HandleRequest(context.Background(), makeRequest(
		http.MethodPatch,
		validRequestBody,
	))

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, resp.StatusCode)
	}

	body := parseErrorResponseBody(t, resp.Body)
	expectedError := "failed to save workouts"
	if body["error"] != expectedError {
		t.Errorf("expected error %q, got %q", expectedError, body["error"])
	}
}

func TestHappySavesWorkout(t *testing.T) {
	mock := &mockDynamo{}
	mock.MockQuery = func(
		ctx context.Context,
		params *dynamodb.QueryInput,
		optFns ...func(*dynamodb.Options),
	) (*dynamodb.QueryOutput, error) {
		return &dynamodb.QueryOutput{
			Items: []map[string]types.AttributeValue{
				{
					"userId": &types.AttributeValueMemberS{ Value: "UserID" },
					"sk": &types.AttributeValueMemberS{ Value: "WorkoutID" },
				},
			},
			Count: 1,
		}, nil
	}
	h := NewHandler(mock, "mockTable")

	resp, err := h.HandleRequest(context.Background(), makeRequest(
		http.MethodPatch,
		validRequestBody,
	))

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, resp.StatusCode)
	}
}
