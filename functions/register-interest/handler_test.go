package main

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider/types"
)

// mockCognito implements CognitoUserCreator for testing.
type mockCognito struct {
	createFunc func(ctx context.Context, params *cognitoidentityprovider.AdminCreateUserInput, optFns ...func(*cognitoidentityprovider.Options)) (*cognitoidentityprovider.AdminCreateUserOutput, error)
	calls      int
	lastInput  *cognitoidentityprovider.AdminCreateUserInput
}

func (m *mockCognito) AdminCreateUser(ctx context.Context, params *cognitoidentityprovider.AdminCreateUserInput, optFns ...func(*cognitoidentityprovider.Options)) (*cognitoidentityprovider.AdminCreateUserOutput, error) {
	m.calls++
	m.lastInput = params
	if m.createFunc != nil {
		return m.createFunc(ctx, params, optFns...)
	}
	return &cognitoidentityprovider.AdminCreateUserOutput{}, nil
}

func makeRequest(body string) events.APIGatewayProxyRequest {
	return events.APIGatewayProxyRequest{
		Body: body,
	}
}

func makeHtmxRequest(body string) events.APIGatewayProxyRequest {
	return events.APIGatewayProxyRequest{
		Body:    body,
		Headers: map[string]string{"HX-Request": "true"},
	}
}

func parseJSONBody(t *testing.T, body string) map[string]any {
	t.Helper()
	var m map[string]any
	if err := json.Unmarshal([]byte(body), &m); err != nil {
		t.Fatalf("failed to parse response body: %v", err)
	}
	return m
}

// ── Tests: Config ───────────────────────────────────────────────────────────

func TestRejectsMissingUserPoolID(t *testing.T) {
	mock := &mockCognito{}
	h := NewHandler(mock, "", true)

	resp, err := h.HandleRequest(context.Background(), makeRequest("email=test@example.com"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, resp.StatusCode)
	}
	body := parseJSONBody(t, resp.Body)
	if body["error"] != "user pool ID not configured" {
		t.Errorf("expected 'user pool ID not configured', got %q", body["error"])
	}
	if mock.calls > 0 {
		t.Error("Cognito should not have been called when user pool ID is empty")
	}
}

// ── Tests: Request body validation ──────────────────────────────────────────

func TestRejectsEmptyBody(t *testing.T) {
	mock := &mockCognito{}
	h := NewHandler(mock, "eu-west-2_TestPool", true)

	resp, err := h.HandleRequest(context.Background(), makeRequest(""))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, resp.StatusCode)
	}
	body := parseJSONBody(t, resp.Body)
	if body["error"] != "request body is empty" {
		t.Errorf("expected 'request body is empty', got %q", body["error"])
	}
}

func TestRejectsMissingEmail(t *testing.T) {
	mock := &mockCognito{}
	h := NewHandler(mock, "eu-west-2_TestPool", true)

	resp, err := h.HandleRequest(context.Background(), makeRequest("foo=bar"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, resp.StatusCode)
	}
	body := parseJSONBody(t, resp.Body)
	if body["error"] != "missing required field: email" {
		t.Errorf("expected 'missing required field: email', got %q", body["error"])
	}
}

func TestRejectsInvalidEmail(t *testing.T) {
	mock := &mockCognito{}
	h := NewHandler(mock, "eu-west-2_TestPool", true)

	cases := []string{
		"email=notanemail",
		"email=@example.com",
		"email=test@",
		"email=test@.com",
		"email=test@com.",
	}

	for _, body := range cases {
		resp, err := h.HandleRequest(context.Background(), makeRequest(body))
		if err != nil {
			t.Fatalf("unexpected error for %s: %v", body, err)
		}
		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("expected status %d for %s, got %d", http.StatusBadRequest, body, resp.StatusCode)
		}
		parsed := parseJSONBody(t, resp.Body)
		if parsed["error"] != "invalid email address" {
			t.Errorf("expected 'invalid email address' for %s, got %q", body, parsed["error"])
		}
	}
}

func TestRejectsInvalidJSON(t *testing.T) {
	mock := &mockCognito{}
	h := NewHandler(mock, "eu-west-2_TestPool", true)

	// Send an invalid form-encoded body (malformed percent-encoding)
	resp, err := h.HandleRequest(context.Background(), makeRequest("%zz"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, resp.StatusCode)
	}
}

// ── Tests: Successful registration (JSON) ───────────────────────────────────

func TestSuccessfulRegistrationJSON(t *testing.T) {
	mock := &mockCognito{}
	h := NewHandler(mock, "eu-west-2_TestPool", true)

	resp, err := h.HandleRequest(context.Background(), makeRequest("email=test@example.com"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}
	if resp.Headers["Content-Type"] != "application/json" {
		t.Errorf("expected Content-Type 'application/json', got %q", resp.Headers["Content-Type"])
	}
	body := parseJSONBody(t, resp.Body)
	if body["message"] != "thanks for registering your interest" {
		t.Errorf("expected message 'thanks for registering your interest', got %q", body["message"])
	}
	if mock.calls != 1 {
		t.Errorf("expected 1 Cognito call, got %d", mock.calls)
	}
}

// ── Tests: Successful registration (htmx) ──────────────────────────────────

func TestSuccessfulRegistrationHtmx(t *testing.T) {
	mock := &mockCognito{}
	h := NewHandler(mock, "eu-west-2_TestPool", true)

	resp, err := h.HandleRequest(context.Background(), makeHtmxRequest("email=test@example.com"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}
	if resp.Headers["Content-Type"] != "text/html" {
		t.Errorf("expected Content-Type 'text/html', got %q", resp.Headers["Content-Type"])
	}
	if !strings.Contains(resp.Body, "register-success") {
		t.Errorf("expected HTML response with register-success class, got %q", resp.Body)
	}
	if !strings.Contains(resp.Body, "Check your email") {
		t.Errorf("expected HTML response with success message, got %q", resp.Body)
	}
}

func TestHtmxErrorReturnsHTML(t *testing.T) {
	mock := &mockCognito{}
	h := NewHandler(mock, "eu-west-2_TestPool", true)

	resp, err := h.HandleRequest(context.Background(), makeHtmxRequest(""))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, resp.StatusCode)
	}
	if resp.Headers["Content-Type"] != "text/html" {
		t.Errorf("expected Content-Type 'text/html', got %q", resp.Headers["Content-Type"])
	}
	if !strings.Contains(resp.Body, "register-error") {
		t.Errorf("expected HTML response with register-error class, got %q", resp.Body)
	}
}

// ── Tests: Email normalisation ──────────────────────────────────────────────

func TestNormalisesEmail(t *testing.T) {
	mock := &mockCognito{}
	h := NewHandler(mock, "eu-west-2_TestPool", true)

	// Use percent-encoding to include leading/trailing spaces in the value
	h.HandleRequest(context.Background(), makeRequest("email=%20%20Test@Example.COM%20%20"))

	if mock.lastInput == nil {
		t.Fatal("expected Cognito to be called")
	}
	if *mock.lastInput.Username != "test@example.com" {
		t.Errorf("expected username 'test@example.com', got %q", *mock.lastInput.Username)
	}
}

// ── Tests: Cognito input correctness ────────────────────────────────────────

func TestSetsCorrectUserAttributes(t *testing.T) {
	mock := &mockCognito{}
	h := NewHandler(mock, "eu-west-2_TestPool", true)

	h.HandleRequest(context.Background(), makeRequest("email=test@example.com"))

	if mock.lastInput == nil {
		t.Fatal("expected Cognito to be called")
	}

	attrs := mock.lastInput.UserAttributes
	if len(attrs) != 2 {
		t.Fatalf("expected 2 user attributes, got %d", len(attrs))
	}

	emailAttr := findAttr(attrs, "email")
	if emailAttr == nil || *emailAttr.Value != "test@example.com" {
		t.Errorf("expected email attribute 'test@example.com', got %v", emailAttr)
	}

	verifiedAttr := findAttr(attrs, "email_verified")
	if verifiedAttr == nil || *verifiedAttr.Value != "true" {
		t.Errorf("expected email_verified attribute 'true', got %v", verifiedAttr)
	}
}

func TestSetsEmailDeliveryMedium(t *testing.T) {
	mock := &mockCognito{}
	h := NewHandler(mock, "eu-west-2_TestPool", true)

	h.HandleRequest(context.Background(), makeRequest("email=test@example.com"))

	if mock.lastInput == nil {
		t.Fatal("expected Cognito to be called")
	}
	if len(mock.lastInput.DesiredDeliveryMediums) != 1 {
		t.Fatalf("expected 1 delivery medium, got %d", len(mock.lastInput.DesiredDeliveryMediums))
	}
	if mock.lastInput.DesiredDeliveryMediums[0] != types.DeliveryMediumTypeEmail {
		t.Errorf("expected EMAIL delivery medium, got %v", mock.lastInput.DesiredDeliveryMediums[0])
	}
}

func TestUsesCorrectUserPoolID(t *testing.T) {
	mock := &mockCognito{}
	h := NewHandler(mock, "eu-west-2_MyPool", true)

	h.HandleRequest(context.Background(), makeRequest("email=test@example.com"))

	if mock.lastInput == nil {
		t.Fatal("expected Cognito to be called")
	}
	if *mock.lastInput.UserPoolId != "eu-west-2_MyPool" {
		t.Errorf("expected user pool ID 'eu-west-2_MyPool', got %q", *mock.lastInput.UserPoolId)
	}
}

// ── Tests: User already exists ──────────────────────────────────────────────

func TestReturnsSuccessWhenUserAlreadyExists(t *testing.T) {
	mock := &mockCognito{
		createFunc: func(ctx context.Context, params *cognitoidentityprovider.AdminCreateUserInput, optFns ...func(*cognitoidentityprovider.Options)) (*cognitoidentityprovider.AdminCreateUserOutput, error) {
			return nil, &types.UsernameExistsException{Message: stringPtr("User already exists")}
		},
	}
	h := NewHandler(mock, "eu-west-2_TestPool", true)

	resp, err := h.HandleRequest(context.Background(), makeRequest("email=existing@example.com"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}
	body := parseJSONBody(t, resp.Body)
	if body["message"] != "thanks for registering your interest" {
		t.Errorf("expected message 'thanks for registering your interest', got %q", body["message"])
	}
}

func TestReturnsHtmxSuccessWhenUserAlreadyExists(t *testing.T) {
	mock := &mockCognito{
		createFunc: func(ctx context.Context, params *cognitoidentityprovider.AdminCreateUserInput, optFns ...func(*cognitoidentityprovider.Options)) (*cognitoidentityprovider.AdminCreateUserOutput, error) {
			return nil, &types.UsernameExistsException{Message: stringPtr("User already exists")}
		},
	}
	h := NewHandler(mock, "eu-west-2_TestPool", true)

	resp, err := h.HandleRequest(context.Background(), makeHtmxRequest("email=existing@example.com"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}
	if resp.Headers["Content-Type"] != "text/html" {
		t.Errorf("expected Content-Type 'text/html', got %q", resp.Headers["Content-Type"])
	}
	if !strings.Contains(resp.Body, "register-success") {
		t.Errorf("expected HTML success response, got %q", resp.Body)
	}
}

// ── Tests: Cognito error ────────────────────────────────────────────────────

func TestCognitoError(t *testing.T) {
	mock := &mockCognito{
		createFunc: func(ctx context.Context, params *cognitoidentityprovider.AdminCreateUserInput, optFns ...func(*cognitoidentityprovider.Options)) (*cognitoidentityprovider.AdminCreateUserOutput, error) {
			return nil, errors.New("service unavailable")
		},
	}
	h := NewHandler(mock, "eu-west-2_TestPool", true)

	resp, err := h.HandleRequest(context.Background(), makeRequest("email=test@example.com"))
	if err == nil {
		t.Fatal("expected error to be returned from HandleRequest")
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, resp.StatusCode)
	}
	body := parseJSONBody(t, resp.Body)
	if body["error"] != "failed to register interest" {
		t.Errorf("expected 'failed to register interest', got %q", body["error"])
	}
}

// ── Tests: Response format ──────────────────────────────────────────────────

func TestJSONResponseHasJSONContentType(t *testing.T) {
	mock := &mockCognito{}
	h := NewHandler(mock, "eu-west-2_TestPool", true)

	resp, _ := h.HandleRequest(context.Background(), makeRequest("email=test@example.com"))
	if resp.Headers["Content-Type"] != "application/json" {
		t.Errorf("expected Content-Type 'application/json', got %q", resp.Headers["Content-Type"])
	}
}

func TestHtmxResponseHasHTMLContentType(t *testing.T) {
	mock := &mockCognito{}
	h := NewHandler(mock, "eu-west-2_TestPool", true)

	resp, _ := h.HandleRequest(context.Background(), makeHtmxRequest("email=test@example.com"))
	if resp.Headers["Content-Type"] != "text/html" {
		t.Errorf("expected Content-Type 'text/html', got %q", resp.Headers["Content-Type"])
	}
}

// ── Helpers ─────────────────────────────────────────────────────────────────

func findAttr(attrs []types.AttributeType, name string) *types.AttributeType {
	for _, a := range attrs {
		if *a.Name == name {
			return &a
		}
	}
	return nil
}

func stringPtr(s string) *string {
	return &s
}
