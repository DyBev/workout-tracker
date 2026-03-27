package main

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// mockDynamo implements DynamoGetter for testing.
type mockDynamo struct {
	getFunc   func(ctx context.Context, params *dynamodb.GetItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error)
	calls     int
	lastInput *dynamodb.GetItemInput
}

func (m *mockDynamo) GetItem(ctx context.Context, params *dynamodb.GetItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error) {
	m.calls++
	m.lastInput = params
	if m.getFunc != nil {
		return m.getFunc(ctx, params, optFns...)
	}
	return &dynamodb.GetItemOutput{}, nil
}

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
				"jwt": map[string]interface{}{},
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

// ── Tests: Table name not configured ────────────────────────────────────────

func TestRejectsEmptyTableName(t *testing.T) {
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
		t.Errorf("expected 'table name not configured', got %q", body["error"])
	}
}

// ── Tests: No profile exists ────────────────────────────────────────────────

func TestReturnsNoneWhenNoProfileExists(t *testing.T) {
	mock := &mockDynamo{
		getFunc: func(ctx context.Context, params *dynamodb.GetItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error) {
			return &dynamodb.GetItemOutput{Item: nil}, nil
		},
	}
	h := NewHandler(mock, "UserProfiles")

	resp, err := h.HandleRequest(context.Background(), makeAuthedRequest("user-123"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}
	body := parseResponseBody(t, resp.Body)
	if body["status"] != "none" {
		t.Errorf("expected status 'none', got %q", body["status"])
	}
	// subscriptionSource should not be present when status is "none"
	if _, exists := body["subscriptionSource"]; exists {
		t.Error("expected subscriptionSource to be absent when status is 'none'")
	}
}

// ── Tests: Manual premium user ──────────────────────────────────────────────

func TestReturnsActiveForManualPremiumUser(t *testing.T) {
	mock := &mockDynamo{
		getFunc: func(ctx context.Context, params *dynamodb.GetItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error) {
			return &dynamodb.GetItemOutput{
				Item: map[string]types.AttributeValue{
					"userId":             &types.AttributeValueMemberS{Value: "user-123"},
					"subscriptionStatus": &types.AttributeValueMemberS{Value: "active"},
					"subscriptionSource": &types.AttributeValueMemberS{Value: "manual"},
					"createdAt":          &types.AttributeValueMemberS{Value: "2026-03-24T00:00:00Z"},
					"updatedAt":          &types.AttributeValueMemberS{Value: "2026-03-24T00:00:00Z"},
				},
			}, nil
		},
	}
	h := NewHandler(mock, "UserProfiles")

	resp, err := h.HandleRequest(context.Background(), makeAuthedRequest("user-123"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}
	body := parseResponseBody(t, resp.Body)
	if body["status"] != "active" {
		t.Errorf("expected status 'active', got %q", body["status"])
	}
	if body["subscriptionSource"] != "manual" {
		t.Errorf("expected subscriptionSource 'manual', got %q", body["subscriptionSource"])
	}
	// Manual subscriptions should not have currentPeriodEnd
	if _, exists := body["currentPeriodEnd"]; exists {
		t.Error("expected currentPeriodEnd to be absent for manual subscription")
	}
}

// ── Tests: Stripe premium user ──────────────────────────────────────────────

func TestReturnsActiveForStripePremiumUser(t *testing.T) {
	mock := &mockDynamo{
		getFunc: func(ctx context.Context, params *dynamodb.GetItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error) {
			return &dynamodb.GetItemOutput{
				Item: map[string]types.AttributeValue{
					"userId":             &types.AttributeValueMemberS{Value: "user-456"},
					"stripeCustomerId":   &types.AttributeValueMemberS{Value: "cus_abc123"},
					"subscriptionId":     &types.AttributeValueMemberS{Value: "sub_xyz789"},
					"subscriptionStatus": &types.AttributeValueMemberS{Value: "active"},
					"subscriptionSource": &types.AttributeValueMemberS{Value: "stripe"},
					"currentPeriodEnd":   &types.AttributeValueMemberN{Value: "1774560000"},
					"createdAt":          &types.AttributeValueMemberS{Value: "2026-03-01T00:00:00Z"},
					"updatedAt":          &types.AttributeValueMemberS{Value: "2026-03-24T00:00:00Z"},
				},
			}, nil
		},
	}
	h := NewHandler(mock, "UserProfiles")

	resp, err := h.HandleRequest(context.Background(), makeAuthedRequest("user-456"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}
	body := parseResponseBody(t, resp.Body)
	if body["status"] != "active" {
		t.Errorf("expected status 'active', got %q", body["status"])
	}
	if body["subscriptionSource"] != "stripe" {
		t.Errorf("expected subscriptionSource 'stripe', got %q", body["subscriptionSource"])
	}
	// Stripe subscriptions should include currentPeriodEnd
	periodEnd, exists := body["currentPeriodEnd"]
	if !exists {
		t.Fatal("expected currentPeriodEnd to be present for Stripe subscription")
	}
	if periodEnd.(float64) != 1774560000 {
		t.Errorf("expected currentPeriodEnd 1774560000, got %v", periodEnd)
	}
}

// ── Tests: Stripe past_due subscription ─────────────────────────────────────

func TestReturnsPastDueForStripeSubscription(t *testing.T) {
	mock := &mockDynamo{
		getFunc: func(ctx context.Context, params *dynamodb.GetItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error) {
			return &dynamodb.GetItemOutput{
				Item: map[string]types.AttributeValue{
					"userId":             &types.AttributeValueMemberS{Value: "user-789"},
					"subscriptionStatus": &types.AttributeValueMemberS{Value: "past_due"},
					"subscriptionSource": &types.AttributeValueMemberS{Value: "stripe"},
					"currentPeriodEnd":   &types.AttributeValueMemberN{Value: "1774560000"},
					"createdAt":          &types.AttributeValueMemberS{Value: "2026-03-01T00:00:00Z"},
					"updatedAt":          &types.AttributeValueMemberS{Value: "2026-03-24T00:00:00Z"},
				},
			}, nil
		},
	}
	h := NewHandler(mock, "UserProfiles")

	resp, err := h.HandleRequest(context.Background(), makeAuthedRequest("user-789"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}
	body := parseResponseBody(t, resp.Body)
	if body["status"] != "past_due" {
		t.Errorf("expected status 'past_due', got %q", body["status"])
	}
	if body["subscriptionSource"] != "stripe" {
		t.Errorf("expected subscriptionSource 'stripe', got %q", body["subscriptionSource"])
	}
}

// ── Tests: Cancelled subscription ───────────────────────────────────────────

func TestReturnsCancelledSubscription(t *testing.T) {
	mock := &mockDynamo{
		getFunc: func(ctx context.Context, params *dynamodb.GetItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error) {
			return &dynamodb.GetItemOutput{
				Item: map[string]types.AttributeValue{
					"userId":             &types.AttributeValueMemberS{Value: "user-000"},
					"subscriptionStatus": &types.AttributeValueMemberS{Value: "cancelled"},
					"subscriptionSource": &types.AttributeValueMemberS{Value: "stripe"},
					"createdAt":          &types.AttributeValueMemberS{Value: "2026-01-01T00:00:00Z"},
					"updatedAt":          &types.AttributeValueMemberS{Value: "2026-03-24T00:00:00Z"},
				},
			}, nil
		},
	}
	h := NewHandler(mock, "UserProfiles")

	resp, err := h.HandleRequest(context.Background(), makeAuthedRequest("user-000"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}
	body := parseResponseBody(t, resp.Body)
	if body["status"] != "cancelled" {
		t.Errorf("expected status 'cancelled', got %q", body["status"])
	}
}

// ── Tests: Key correctness ──────────────────────────────────────────────────

func TestUsesCorrectKey(t *testing.T) {
	mock := &mockDynamo{}
	h := NewHandler(mock, "UserProfiles")

	h.HandleRequest(context.Background(), makeAuthedRequest("user-xyz"))

	if mock.calls != 1 {
		t.Fatalf("expected 1 DynamoDB call, got %d", mock.calls)
	}

	key := mock.lastInput.Key
	userIDVal, ok := key["userId"].(*types.AttributeValueMemberS)
	if !ok {
		t.Fatal("expected userId in key to be a string attribute")
	}
	if userIDVal.Value != "user-xyz" {
		t.Errorf("expected userId 'user-xyz', got %q", userIDVal.Value)
	}

	if *mock.lastInput.TableName != "UserProfiles" {
		t.Errorf("expected table name 'UserProfiles', got %q", *mock.lastInput.TableName)
	}
}

// ── Tests: DynamoDB error ─────────────────────────────────────────────────────

func TestDynamoDBGetError(t *testing.T) {
	mock := &mockDynamo{
		getFunc: func(ctx context.Context, params *dynamodb.GetItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error) {
			return nil, errors.New("service unavailable")
		},
	}
	h := NewHandler(mock, "UserProfiles")

	resp, err := h.HandleRequest(context.Background(), makeAuthedRequest("user-123"))
	if err == nil {
		t.Fatal("expected error to be returned from HandleRequest")
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, resp.StatusCode)
	}
	body := parseResponseBody(t, resp.Body)
	if body["error"] != "failed to read subscription status" {
		t.Errorf("expected 'failed to read subscription status', got %q", body["error"])
	}
}

// ── Tests: Response format ──────────────────────────────────────────────────

func TestResponseHasJSONContentType(t *testing.T) {
	mock := &mockDynamo{}
	h := NewHandler(mock, "UserProfiles")

	resp, _ := h.HandleRequest(context.Background(), makeAuthedRequest("user-123"))
	if resp.Headers["Content-Type"] != "application/json" {
		t.Errorf("expected Content-Type 'application/json', got %q", resp.Headers["Content-Type"])
	}
}

func TestErrorResponseHasJSONContentType(t *testing.T) {
	mock := &mockDynamo{}
	h := NewHandler(mock, "UserProfiles")

	resp, _ := h.HandleRequest(context.Background(), makeUnauthedRequest())
	if resp.Headers["Content-Type"] != "application/json" {
		t.Errorf("expected Content-Type 'application/json', got %q", resp.Headers["Content-Type"])
	}
}
