package handler

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// TEST: happy path
func TestSaveArray(t *testing.T) {
	mock := &mockDynamo{}
	h := NewHandler(mock, mockAttributeValueMapper, "mockTable")
	resp, _ := h.HandleRequest(context.Background(), makeRequest(http.MethodPost, []Workout{ 
		validWorkout("W1", ""),
		validWorkout("W2", ""),
	}))

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("Expected status %q, got %q", http.StatusCreated, resp.StatusCode)
	}

	message, parsedBody := parseBatchResponseBody(t, resp.Body)
	if message != "batch complete" {
		t.Fatalf("Unexpeted completion message")
	}

	if !findWorkout(parsedBody, "W1") ||
		!findWorkout(parsedBody, "W2") {
		t.Fatalf("missing workout ID in parsed Body: %q", parsedBody)
	}
}

func TestSaveSingle(t *testing.T) {
	mock := &mockDynamo{}
	h := NewHandler(mock, mockAttributeValueMapper, "mockTable")
	resp, _ := h.HandleRequest(context.Background(), makeRequestSingleWorkout(http.MethodPost,
		validWorkout("W1", ""),
	))

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("Expected status %q, got %q", http.StatusCreated, resp.StatusCode)
	}

	message, parsedBody := parseBatchResponseBody(t, resp.Body)
	if message != "batch complete" {
		t.Fatalf("Unexpeted completion message")
	}

	if !findWorkout(parsedBody, "W1") {
		t.Fatalf("missing workout ID in parsed Body: %q", parsedBody)
	}
}

func TestSaveSingleWithSpaceNoteFormatting(t *testing.T) {
	mock := &mockDynamo{}
	h := NewHandler(mock, mockAttributeValueMapper, "mockTable")
	resp, _ := h.HandleRequest(context.Background(), makeRequest(http.MethodPost, []Workout{
		validWorkoutWithSet("W1", []WorkoutExercise{
			validExerciseWithNote("Ex1", 1, []WorkoutSet{}, "   "),
			validExerciseWithNote("Ex2", 2, []WorkoutSet{}, "testNote"),
		}),
	}))

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("Expected status %q, got %q", http.StatusCreated, resp.StatusCode)
	}

	message, parsedBody := parseBatchResponseBody(t, resp.Body)
	if message != "batch complete" {
		t.Fatalf("Unexpeted completion message")
	}

	if !findWorkout(parsedBody, "W1") {
		t.Fatalf("missing workout ID in parsed Body: %q", parsedBody)
	}
}

// TEST: happy path batching and retries
func TestSaveBatching(t *testing.T) {
	callCount := 0
	var requestSizes []int
	mock := &mockDynamo{
		batchWriteFunc: func(
			ctx context.Context,
			params *dynamodb.BatchWriteItemInput,
			optFns ...func(*dynamodb.Options),
		) (*dynamodb.BatchWriteItemOutput, error) {
			callCount++
			requestSizes = append(requestSizes, len(params.RequestItems["TestTable"]))
			return &dynamodb.BatchWriteItemOutput{}, nil
		},
	}
	h := NewHandler(mock, mockAttributeValueMapper, "TestTable")

	var WorkoutArray []Workout;
	for i := range 30 {
		WorkoutArray = append(WorkoutArray, validWorkout(fmt.Sprintf("W%d", i), ""))
	}

	resp, _ := h.HandleRequest(
		context.Background(),
		makeRequest(http.MethodPost, WorkoutArray),
	)
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

func TestReportsFailedExercises(t *testing.T) {
	callCount := 0
	mock := &mockDynamo{
		batchWriteFunc: func(ctx context.Context, params *dynamodb.BatchWriteItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.BatchWriteItemOutput, error) {
			callCount++
			requests := params.RequestItems["TestTable"]
			var unprocessed []types.WriteRequest
			for _, req := range requests {
				if v, ok := req.PutRequest.Item["workoutId"].(*types.AttributeValueMemberS); ok && v.Value == "W1" {
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
	h := NewHandler(mock, mockAttributeValueMapper, "TestTable")

	var WorkoutArray []Workout;
	for i := range 2 {
		WorkoutArray = append(WorkoutArray, validWorkout(fmt.Sprintf("W%d", i), ""))
	}

	resp, _ := h.HandleRequest(
		context.Background(),
		makeRequest(http.MethodPost, WorkoutArray),
	)
	if resp.StatusCode != http.StatusCreated {
		t.Errorf("expected status %d, got %d", http.StatusCreated, resp.StatusCode)
	}

	_, results := parseBatchResponseBody(t, resp.Body)
	if len(results) != 2 {
		t.Fatalf("expected 30 results, got %d", len(results))
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

func TestSaveRetries(t *testing.T) {
	callCount := 0
	var requestSizes []int
	mock := &mockDynamo{
		batchWriteFunc: func(
			ctx context.Context,
			params *dynamodb.BatchWriteItemInput,
			optFns ...func(*dynamodb.Options),
		) (*dynamodb.BatchWriteItemOutput, error) {
			callCount++
			requestSizes = append(requestSizes, len(params.RequestItems["TestTable"]))
			return &dynamodb.BatchWriteItemOutput{}, nil
		},
	}
	h := NewHandler(mock, mockAttributeValueMapper, "TestTable")

	var WorkoutArray []Workout;
	for i := range 30 {
		WorkoutArray = append(WorkoutArray, validWorkout(fmt.Sprintf("W%d", i), ""))
	}

	resp, _ := h.HandleRequest(
		context.Background(),
		makeRequest(http.MethodPost, WorkoutArray),
	)
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

// TEST: unhappy path tableName undefined
func TestTableNameUndefined(t *testing.T) {
	mock := &mockDynamo{}
	h := NewHandler(mock, mockAttributeValueMapper, "")
	resp, err := h.HandleRequest(context.Background(), makeRequest(http.MethodPost, []Workout{ validWorkout("W1", "") }))

	if err != nil {
		t.Fatalf("Unexpected Error: %q", err.Error())
	}

	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("Expected status %q, got %q", http.StatusInternalServerError, resp.StatusCode)
	}

	message := parseErrorResponseBody(t, resp.Body)
	expectedErrorMessage := "table name not configured"
	if message["error"] != expectedErrorMessage {
		t.Fatalf("Expected error message: %q, got: %q", expectedErrorMessage, message["error"])
	}
}

func TestSaveSingleWithEmptyNote(t *testing.T) {
	mock := &mockDynamo{}
	h := NewHandler(mock, mockAttributeValueMapper, "mockTable")
	resp, _ := h.HandleRequest(context.Background(), makeRequestSingleWorkout(http.MethodPost,
		validWorkout("W1", "  "),
	))

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("Expected status %q, got %q", http.StatusCreated, resp.StatusCode)
	}

	message, parsedBody := parseBatchResponseBody(t, resp.Body)
	if message != "batch complete" {
		t.Fatalf("Unexpeted completion message")
	}

	if !findWorkout(parsedBody, "W1") {
		t.Fatalf("missing workout ID in parsed Body: %q", parsedBody)
	}
}

// TEST: unhappy path unauthorized user
func TestRejectsUnauthenticated(t *testing.T) {
	mock := &mockDynamo{}
	h := NewHandler(mock, mockAttributeValueMapper, "testTable")

	resp, err := h.HandleRequest(context.Background(), makeRequestWithContext(
		http.MethodPost,
		[]Workout{},
		nil,
	))

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, resp.StatusCode)
	}

}

func TestRejectsUnauthenticatedMissingJWT(t *testing.T) {
	mock := &mockDynamo{}
	h := NewHandler(mock, mockAttributeValueMapper, "testTable")

	resp, err := h.HandleRequest(context.Background(), makeRequestWithContext(
		http.MethodPost,
		[]Workout{},
		&events.APIGatewayProxyRequestContext{
			Authorizer: map[string]any{},
		},
	))

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, resp.StatusCode)
	}
}

func TestRejectsUnauthenticatedMissingClaims(t *testing.T) {
	mock := &mockDynamo{}
	h := NewHandler(mock, mockAttributeValueMapper, "testTable")

	resp, err := h.HandleRequest(context.Background(), makeRequestWithContext(
		http.MethodPost,
		[]Workout{},
		&events.APIGatewayProxyRequestContext{
			Authorizer: map[string]any{
				"jwt": map[string]any{ },
			},
		},
	))

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, resp.StatusCode)
	}
}

func TestRejectsUnauthenticatedMissingSub(t *testing.T) {
	mock := &mockDynamo{}
	h := NewHandler(mock, mockAttributeValueMapper, "testTable")
	resp, err := h.HandleRequest(context.Background(), makeRequestWithContext(
		http.MethodPost,
		[]Workout{},
		&events.APIGatewayProxyRequestContext{
			Authorizer: map[string]any{
				"jwt": map[string]any{
					"claims": map[string]any{ },
				},
			},
		},
	))

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, resp.StatusCode)
	}
}

func TestRejectsUnauthenticatedMalformedSub(t *testing.T) {
	mock := &mockDynamo{}
	h := NewHandler(mock, mockAttributeValueMapper, "testTable")
	resp, err := h.HandleRequest(context.Background(), makeRequestWithContext(
		http.MethodPost,
		[]Workout{},
		&events.APIGatewayProxyRequestContext{
			Authorizer: map[string]any{
				"jwt": map[string]any{
					"claims": map[string]any{
						"sub": "",
					},
				},
			},
		},
	))

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, resp.StatusCode)
	}
}

// TEST: unhappy path parseAndValidateBody
func TestRejectsInvalidEmptyBody(t *testing.T) {
	mock := &mockDynamo{}
	h := NewHandler(mock, mockAttributeValueMapper, "testTable")
	resp, err := h.HandleRequest(context.Background(), makeRequestWithString(
		http.MethodPost,
		"",
	))

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, resp.StatusCode)
	}
}

func TestRejectsInvalidJSONBody(t *testing.T) {
	mock := &mockDynamo{}
	h := NewHandler(mock, mockAttributeValueMapper, "testTable")
	resp, err := h.HandleRequest(context.Background(), makeRequestWithString(
		http.MethodPost,
		"SomeRandoNonJSONString",
	))

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, resp.StatusCode)
	}
}

func TestRejectsMalformedJSONBody(t *testing.T) {
	mock := &mockDynamo{}
	h := NewHandler(mock, mockAttributeValueMapper, "testTable")
	resp, err := h.HandleRequest(context.Background(), makeRequestWithString(
		http.MethodPost,
		"[",
	))

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, resp.StatusCode)
	}
}

func TestRejectsEmptyArrayJSONBody(t *testing.T) {
	mock := &mockDynamo{}
	h := NewHandler(mock, mockAttributeValueMapper, "testTable")
	resp, err := h.HandleRequest(context.Background(), makeRequestWithString(
		http.MethodPost,
		"[]",
	))

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, resp.StatusCode)
	}
}

func TestRejectsBodyMissingWorkoutID(t *testing.T) {
	mock := &mockDynamo{}
	h := NewHandler(mock, mockAttributeValueMapper, "testTable")
	resp, err := h.HandleRequest(context.Background(), makeRequest(
		http.MethodPost,
		[]Workout{
			invalidWorkoutNoID(),
		},
	))

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, resp.StatusCode)
	}
}

func TestRejectsBodyMissingStartedAt(t *testing.T) {
	mock := &mockDynamo{}
	h := NewHandler(mock, mockAttributeValueMapper, "testTable")
	resp, err := h.HandleRequest(context.Background(), makeRequest(
		http.MethodPost,
		[]Workout{
			invalidWorkoutNoStarted("1"),
		},
	))

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, resp.StatusCode)
	}
}

func TestRejectsBodyMissingUpdatedAt(t *testing.T) {
	mock := &mockDynamo{}
	h := NewHandler(mock, mockAttributeValueMapper, "testTable")
	resp, err := h.HandleRequest(context.Background(), makeRequest(
		http.MethodPost,
		[]Workout{
			invalidWorkoutNoUpdatedAt("1"),
		},
	))

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, resp.StatusCode)
	}
}

func TestRejectsBodyMissingCreatedAt(t *testing.T) {
	mock := &mockDynamo{}
	h := NewHandler(mock, mockAttributeValueMapper, "testTable")
	resp, err := h.HandleRequest(context.Background(), makeRequest(
		http.MethodPost,
		[]Workout{
			invalidWorkoutNoCreatedAt("1"),
		},
	))

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, resp.StatusCode)
	}
}

func TestRejectsBodyMissingWorkoutIDSingleWorkout(t *testing.T) {
	mock := &mockDynamo{}
	h := NewHandler(mock, mockAttributeValueMapper, "testTable")
	resp, err := h.HandleRequest(context.Background(), makeRequestSingleWorkout(
		http.MethodPost,
		invalidWorkoutNoID(),
	))

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, resp.StatusCode)
	}
}

func TestRejectsBodyMissingWorkoutExerciseID(t *testing.T) {
	mock := &mockDynamo{}
	h := NewHandler(mock, mockAttributeValueMapper, "testTable")
	resp, err := h.HandleRequest(context.Background(), makeRequest(
		http.MethodPost,
		[]Workout{
			validWorkoutWithSet("1", []WorkoutExercise{ validExercise("", 1, []WorkoutSet{ validSet("1") }) }),
		},
	))

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, resp.StatusCode)
	}
}

func TestRejectsBodyMissingWorkoutName(t *testing.T) {
	mock := &mockDynamo{}
	h := NewHandler(mock, mockAttributeValueMapper, "testTable")
	resp, err := h.HandleRequest(context.Background(), makeRequest(
		http.MethodPost,
		[]Workout{
			validWorkoutWithSet("1", []WorkoutExercise{ 
				invalidExerciseMissingName("1", 1, []WorkoutSet{ validSet("1") }),
			}),
		},
	))

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, resp.StatusCode)
	}
}

func TestRejectsBodyLongNote(t *testing.T) {
	mock := &mockDynamo{}
	h := NewHandler(mock, mockAttributeValueMapper, "testTable")
	resp, err := h.HandleRequest(context.Background(), makeRequest(
		http.MethodPost,
		[]Workout{
			validWorkoutWithSet("1", []WorkoutExercise{ 
				invalidExerciseLongNote("1", 1, []WorkoutSet{ validSet("1") }),
			}),
		},
	))

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, resp.StatusCode)
	}
}

func TestRejectsBodyInvalidSet(t *testing.T) {
	mock := &mockDynamo{}
	h := NewHandler(mock, mockAttributeValueMapper, "testTable")
	resp, err := h.HandleRequest(context.Background(), makeRequest(
		http.MethodPost,
		[]Workout{
			validWorkoutWithSet("1", []WorkoutExercise{ 
				validExercise("1", 1, []WorkoutSet{ validSet("") }),
			}),
		},
	))

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, resp.StatusCode)
	}
}

// TEST: unhappy path SaveWorkouts
func TestSaveWorkoutsBuildWriteRequestsError(t *testing.T) {
	mock := &mockDynamo{}
	var mockAttributeValueMapperError = func(
		in any,
	) (map[string]types.AttributeValue, error) {
		return map[string]types.AttributeValue{}, errors.New("Mock random error!")
	}
	h := NewHandler(mock, mockAttributeValueMapperError, "table name")
	resp, _ := h.HandleRequest(context.Background(), makeRequest(http.MethodPost, []Workout{
		validWorkout("w1", ""),
	}))

	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("Expected status %q, got %q", http.StatusInternalServerError, resp.StatusCode);
	}
}

// TEST: unhappy path BatchWriteFailure
func TestSaveWorkoutBatchWriteFailure(t *testing.T) {
	mock := &mockDynamo{
		batchWriteFunc: func(ctx context.Context, params *dynamodb.BatchWriteItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.BatchWriteItemOutput, error) {
			return nil, errors.New("service unavailable")
		},
	}
	h := NewHandler(mock, mockAttributeValueMapper, "table name")
	resp, _ := h.HandleRequest(context.Background(), makeRequest(http.MethodPut, []Workout{
		validWorkout("w1", ""),
	}))
	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("Expected status %q, got %q", http.StatusInternalServerError, resp.StatusCode);
	}
}

// TEST: unhappy path BatchWriteFailure retries exhausted
func TestSaveWorkoutBatchWriteRetriesExhausted(t *testing.T) {
	callCount := 0
	mock := &mockDynamo{
		batchWriteFunc: func(
			ctx context.Context,
			params *dynamodb.BatchWriteItemInput,
			optFns ...func(*dynamodb.Options),
		) (*dynamodb.BatchWriteItemOutput, error) {
			callCount++
			return &dynamodb.BatchWriteItemOutput{
				UnprocessedItems: params.RequestItems,
			}, nil
		},
	}
	h := NewHandler(mock, mockAttributeValueMapper, "TestTable")

	resp, _ := h.HandleRequest(context.Background(), makeRequest(http.MethodPost, []Workout{
		validWorkout("w1", ""),
	}))

	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, resp.StatusCode)
	}

	if callCount != 4 {
		t.Errorf("expected 4 BatchWriteItem calls (1 + 3 retries), got %d", callCount)
	}
}

// TEST: unhappy path context expiry
func TestSaveWorkoutContexExpiry(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

    // Cancel the context after 100 milliseconds.
    go func() {
        time.Sleep(100 * time.Millisecond)
        cancel()
    }()

	mock := &mockDynamo{
		batchWriteFunc: func(
			ctx context.Context,
			params *dynamodb.BatchWriteItemInput,
			optFns ...func(*dynamodb.Options),
		) (*dynamodb.BatchWriteItemOutput, error) {
			return &dynamodb.BatchWriteItemOutput{
				UnprocessedItems: params.RequestItems,
			}, nil
		},
	}
	h := NewHandler(mock, mockAttributeValueMapper, "TestTable")

	resp, _ := h.HandleRequest(ctx, makeRequest(http.MethodPost, []Workout{
		validWorkout("w1", ""),
	}))

	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, resp.StatusCode)
	}
}
