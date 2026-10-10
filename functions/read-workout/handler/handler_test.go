package handler

import (
	"context"
	"errors"
	"fmt"
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

	resp, err := h.HandleRequest(context.Background(), makeAuthedRequest("user-123", nil))
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

// TEST: Successful read
func TestReturnsFirstPageWithNoSKParam(t *testing.T) {
	items := buildWorkoutItems("user-123", 20)
	mock := &mockDynamo{
		queryFunc: func(ctx context.Context, params *dynamodb.QueryInput, optFns ...func(*dynamodb.Options)) (*dynamodb.QueryOutput, error) {
			return &dynamodb.QueryOutput{Items: items}, nil
		},
	}
	h := NewHandler(mock, "TestTable")

	resp, err := h.HandleRequest(context.Background(), makeAuthedRequest("user-123", nil))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}

	body := parseResponseBody(t, resp.Body)
	workoutsRaw, ok := body["workouts"].([]any)
	if !ok {
		t.Fatalf("expected 'workouts' array in response, got %T", body["workouts"])
	}
	if len(workoutsRaw) != 20 {
		t.Errorf("expected 20 workouts, got %d", len(workoutsRaw))
	}
	if body["nextSk"] != nil {
		t.Errorf("expected nextSk to be null when no more pages, got %v", body["nextSk"])
	}
}

func TestNoSKParamMeansNoExclusiveStartKey(t *testing.T) {
	mock := &mockDynamo{
		queryFunc: func(ctx context.Context, params *dynamodb.QueryInput, optFns ...func(*dynamodb.Options)) (*dynamodb.QueryOutput, error) {
			return &dynamodb.QueryOutput{Items: []map[string]types.AttributeValue{}}, nil
		},
	}
	h := NewHandler(mock, "TestTable")

	h.HandleRequest(context.Background(), makeAuthedRequest("user-123", nil))

	if mock.lastInput.ExclusiveStartKey != nil {
		t.Error("expected ExclusiveStartKey to be nil when no sk query param")
	}
}

// TEST: Pagination Read
func TestUsesSKParamAsExclusiveStartKey(t *testing.T) {
	cursor := "WORKOUT#2026-03-01T10:00:00.000Z#wkt-001"
	mock := &mockDynamo{
		queryFunc: func(ctx context.Context, params *dynamodb.QueryInput, optFns ...func(*dynamodb.Options)) (*dynamodb.QueryOutput, error) {
			return &dynamodb.QueryOutput{Items: []map[string]types.AttributeValue{}}, nil
		},
	}
	h := NewHandler(mock, "TestTable")

	h.HandleRequest(context.Background(), makeAuthedRequest("user-123", map[string]string{"sk": cursor}))

	esk := mock.lastInput.ExclusiveStartKey
	if esk == nil {
		t.Fatal("expected ExclusiveStartKey to be set when sk param is provided")
	}
	skVal, ok := esk["sk"].(*types.AttributeValueMemberS)
	if !ok {
		t.Fatal("expected sk in ExclusiveStartKey to be a string attribute")
	}
	if skVal.Value != cursor {
		t.Errorf("expected ExclusiveStartKey sk = %q, got %q", cursor, skVal.Value)
	}
	userIDVal, ok := esk["userId"].(*types.AttributeValueMemberS)
	if !ok {
		t.Fatal("expected userId in ExclusiveStartKey to be a string attribute")
	}
	if userIDVal.Value != "user-123" {
		t.Errorf("expected ExclusiveStartKey userId = %q, got %q", "user-123", userIDVal.Value)
	}
}

func TestReturnsNextSKWhenMorePagesExist(t *testing.T) {
	lastSK := "WORKOUT#2026-02-28T10:00:00.000Z#wkt-019"
	items := buildWorkoutItems("user-123", 20)
	mock := &mockDynamo{
		queryFunc: func(ctx context.Context, params *dynamodb.QueryInput, optFns ...func(*dynamodb.Options)) (*dynamodb.QueryOutput, error) {
			return &dynamodb.QueryOutput{
				Items: items,
				LastEvaluatedKey: map[string]types.AttributeValue{
					"userId": &types.AttributeValueMemberS{Value: "user-123"},
					"sk":     &types.AttributeValueMemberS{Value: lastSK},
				},
			}, nil
		},
	}
	h := NewHandler(mock, "TestTable")

	resp, err := h.HandleRequest(context.Background(), makeAuthedRequest("user-123", nil))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	body := parseResponseBody(t, resp.Body)
	if body["nextSk"] == nil {
		t.Fatal("expected nextSk to be set when more pages exist")
	}
	if body["nextSk"] != lastSK {
		t.Errorf("expected nextSk = %q, got %q", lastSK, body["nextSk"])
	}
}

func TestNextSKIsNullWhenNoMorePages(t *testing.T) {
	mock := &mockDynamo{
		queryFunc: func(ctx context.Context, params *dynamodb.QueryInput, optFns ...func(*dynamodb.Options)) (*dynamodb.QueryOutput, error) {
			return &dynamodb.QueryOutput{
				Items:            buildWorkoutItems("user-123", 5),
				LastEvaluatedKey: nil, // no more pages
			}, nil
		},
	}
	h := NewHandler(mock, "TestTable")

	resp, err := h.HandleRequest(context.Background(), makeAuthedRequest("user-123", nil))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	body := parseResponseBody(t, resp.Body)
	if body["nextSk"] != nil {
		t.Errorf("expected nextSk null when no LastEvaluatedKey, got %v", body["nextSk"])
	}
}

// TEST: DynamoDB query parameters
func TestQueryUsesCorrectTableName(t *testing.T) {
	mock := &mockDynamo{}
	h := NewHandler(mock, "MyWorkoutsTable")

	h.HandleRequest(context.Background(), makeAuthedRequest("user-123", nil))

	if *mock.lastInput.TableName != "MyWorkoutsTable" {
		t.Errorf("expected TableName 'MyWorkoutsTable', got %q", *mock.lastInput.TableName)
	}
}

func TestQueryUsesDescendingOrder(t *testing.T) {
	mock := &mockDynamo{}
	h := NewHandler(mock, "TestTable")

	h.HandleRequest(context.Background(), makeAuthedRequest("user-123", nil))

	if mock.lastInput.ScanIndexForward == nil || *mock.lastInput.ScanIndexForward {
		t.Error("expected ScanIndexForward=false for descending order")
	}
}

func TestQueryLimitIsPageSize(t *testing.T) {
	mock := &mockDynamo{}
	h := NewHandler(mock, "TestTable")

	h.HandleRequest(context.Background(), makeAuthedRequest("user-123", nil))

	if mock.lastInput.Limit == nil || *mock.lastInput.Limit != pageSize {
		t.Errorf("expected Limit=%d, got %v", pageSize, mock.lastInput.Limit)
	}
}

func TestQueryFiltersToWorkoutPrefix(t *testing.T) {
	mock := &mockDynamo{}
	h := NewHandler(mock, "TestTable")

	h.HandleRequest(context.Background(), makeAuthedRequest("user-123", nil))

	prefixAttr, ok := mock.lastInput.ExpressionAttributeValues[":prefix"]
	if !ok {
		t.Fatal("expected :prefix expression attribute value")
	}
	prefixVal, ok := prefixAttr.(*types.AttributeValueMemberS)
	if !ok {
		t.Fatal("expected :prefix to be a string attribute")
	}
	if prefixVal.Value != "WORKOUT#" {
		t.Errorf("expected :prefix = 'WORKOUT#', got %q", prefixVal.Value)
	}
}

func TestQueryScopedToCurrentUser(t *testing.T) {
	mock := &mockDynamo{}
	h := NewHandler(mock, "TestTable")

	h.HandleRequest(context.Background(), makeAuthedRequest("user-xyz", nil))

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

// Tests: Empty result
func TestReturnsEmptyArrayWhenNoWorkouts(t *testing.T) {
	mock := &mockDynamo{
		queryFunc: func(ctx context.Context, params *dynamodb.QueryInput, optFns ...func(*dynamodb.Options)) (*dynamodb.QueryOutput, error) {
			return &dynamodb.QueryOutput{Items: []map[string]types.AttributeValue{}}, nil
		},
	}
	h := NewHandler(mock, "TestTable")

	resp, err := h.HandleRequest(context.Background(), makeAuthedRequest("user-123", nil))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}

	body := parseResponseBody(t, resp.Body)
	workoutsRaw, ok := body["workouts"].([]any)
	if !ok {
		t.Fatalf("expected 'workouts' array in response, got %T", body["workouts"])
	}
	if len(workoutsRaw) != 0 {
		t.Errorf("expected 0 workouts, got %d", len(workoutsRaw))
	}
}

// TEST: DynamoDB error
func TestDynamoDBQueryError(t *testing.T) {
	mock := &mockDynamo{
		queryFunc: func(ctx context.Context, params *dynamodb.QueryInput, optFns ...func(*dynamodb.Options)) (*dynamodb.QueryOutput, error) {
			return nil, errors.New("service unavailable")
		},
	}
	h := NewHandler(mock, "TestTable")

	resp, err := h.HandleRequest(context.Background(), makeAuthedRequest("user-123", nil))
	if err != nil {
		t.Errorf("unexpected error returned from HandleRequest: %q", err.Error())
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, resp.StatusCode)
	}
	body := parseResponseBody(t, resp.Body)
	if body["error"] != "failed to read workouts" {
		t.Errorf("expected 'failed to read workouts', got %q", body["error"])
	}
}

func TestDynamoDBContentError(t *testing.T) {
	mock := &mockDynamo{
		queryFunc: func(ctx context.Context, params *dynamodb.QueryInput, optFns ...func(*dynamodb.Options)) (*dynamodb.QueryOutput, error) {
			return &dynamodb.QueryOutput{Items: []map[string]types.AttributeValue{
				{
					"PK": &types.AttributeValueMemberN{ Value: "10" },
					"userId":          &types.AttributeValueMemberN{Value: "10"},
					"sk":              &types.AttributeValueMemberS{Value: "EXERCISE#"},
					"savedExerciseId": &types.AttributeValueMemberS{Value: ""},
					"name":            &types.AttributeValueMemberS{Value: "Exercise 10"},
					"note":            &types.AttributeValueMemberS{Value: ""},
					"tags":            &types.AttributeValueMemberN{Value: "10"},
					"createdAt":       &types.AttributeValueMemberS{Value: fmt.Sprintf("2026-03-%02dT10:00:00.000Z", 0+1)},
					"updatedAt":       &types.AttributeValueMemberS{Value: fmt.Sprintf("2026-03-%02dT10:00:00.000Z", 0+1)},
				},
			}}, nil
		},
	}
	h := NewHandler(mock, "TestTable")

	resp, err := h.HandleRequest(context.Background(), makeAuthedRequest("user-123", nil))
	if err != nil {
		t.Errorf("unexpected error returned from HandleRequest: %q", err.Error())
	}

	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, resp.StatusCode)
	}

	body := parseResponseBody(t, resp.Body)
	if body["error"] != "failed to read workouts" {
		t.Errorf("expected 'failed to read workouts', got %q", body["error"])
	}
}

// TEST: Response format
func TestResponseHasJSONContentType(t *testing.T) {
	mock := &mockDynamo{}
	h := NewHandler(mock, "TestTable")

	resp, _ := h.HandleRequest(context.Background(), makeAuthedRequest("user-123", nil))
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

// TEST: Workout fields are returned correctly
func TestWorkoutFieldsAreUnmarshalled(t *testing.T) {
	reps := 10
	weight := 80.0
	items := []map[string]types.AttributeValue{
		{
			"userId":    &types.AttributeValueMemberS{Value: "user-123"},
			"sk":        &types.AttributeValueMemberS{Value: "WORKOUT#2026-03-14T10:00:00.000Z#wkt-abc"},
			"workoutId": &types.AttributeValueMemberS{Value: "wkt-abc"},
			"startedAt": &types.AttributeValueMemberS{Value: "2026-03-14T10:00:00.000Z"},
			"notes":     &types.AttributeValueMemberS{Value: "Great session"},
			"createdAt": &types.AttributeValueMemberS{Value: "2026-03-14T10:00:00.000Z"},
			"updatedAt": &types.AttributeValueMemberS{Value: "2026-03-14T10:30:00.000Z"},
			"exercises": &types.AttributeValueMemberL{
				Value: []types.AttributeValue{
					&types.AttributeValueMemberM{
						Value: map[string]types.AttributeValue{
							"exerciseId": &types.AttributeValueMemberS{Value: "ex-1"},
							"name":       &types.AttributeValueMemberS{Value: "Bench Press"},
							"order":      &types.AttributeValueMemberN{Value: "0"},
							"sets": &types.AttributeValueMemberL{
								Value: []types.AttributeValue{
									&types.AttributeValueMemberM{
										Value: map[string]types.AttributeValue{
											"setId":      &types.AttributeValueMemberS{Value: "s-1"},
											"order":      &types.AttributeValueMemberN{Value: "0"},
											"reps":       &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", reps)},
											"weight":     &types.AttributeValueMemberN{Value: fmt.Sprintf("%g", weight)},
											"weightUnit": &types.AttributeValueMemberS{Value: "kg"},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	mock := &mockDynamo{
		queryFunc: func(ctx context.Context, params *dynamodb.QueryInput, optFns ...func(*dynamodb.Options)) (*dynamodb.QueryOutput, error) {
			return &dynamodb.QueryOutput{Items: items}, nil
		},
	}
	h := NewHandler(mock, "TestTable")

	resp, err := h.HandleRequest(context.Background(), makeAuthedRequest("user-123", nil))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}

	body := parseResponseBody(t, resp.Body)
	workoutsRaw, ok := body["workouts"].([]any)
	if !ok || len(workoutsRaw) != 1 {
		t.Fatalf("expected 1 workout in response")
	}

	w := workoutsRaw[0].(map[string]any)
	if w["workoutId"] != "wkt-abc" {
		t.Errorf("expected workoutId 'wkt-abc', got %q", w["workoutId"])
	}
	if w["notes"] != "Great session" {
		t.Errorf("expected notes 'Great session', got %q", w["notes"])
	}
	exercises, ok := w["exercises"].([]any)
	if !ok || len(exercises) != 1 {
		t.Fatalf("expected 1 exercise in workout")
	}
	ex := exercises[0].(map[string]any)
	if ex["name"] != "Bench Press" {
		t.Errorf("expected exercise name 'Bench Press', got %q", ex["name"])
	}
}
