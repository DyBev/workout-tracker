package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

const (
	maxBatchSize = 25
	maxRetries = 3
	baseRetryDelay = 100 * time.Millisecond
)

// DynamoBatchWriter is the subset of the DynamoDB client that this handler needs.
// Using an interface makes the handler easy to test with a mock.
type DynamoBatchWriter interface {
	BatchWriteItem(ctx context.Context, params *dynamodb.BatchWriteItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.BatchWriteItemOutput, error)
}

// Handler holds the dependencies injected at startup.
type Handler struct {
	db        DynamoBatchWriter
	tableName string
}

// NewHandler creates a Handler with the given DynamoDB client and table name.
func NewHandler(db DynamoBatchWriter, tableName string) *Handler {
	return &Handler{db: db, tableName: tableName}
}

// HandleRequest processes the API Gateway proxy request.
// The request body may be a single Workout object or an array of Workouts.
func (h *Handler) HandleRequest(
	ctx context.Context,
	req events.APIGatewayProxyRequest,
) (events.APIGatewayProxyResponse, error) {
	if h.tableName == "" {
		return response(http.StatusInternalServerError, errorBody("table name not configured")), nil
	}

	var userID string;
	if req.RequestContext.Authorizer != nil {
		jwt, ok := req.RequestContext.Authorizer["jwt"];
		if !ok {
			return response(http.StatusUnauthorized, errorBody("not authorised")), nil
		}
		claims, ok := jwt.(map[string]any)["claims"];
		if !ok {
			return response(http.StatusUnauthorized, errorBody("not authorised")), nil
		}
		userID, ok = claims.(map[string]any)["sub"].(string)
		if !ok {
			return response(http.StatusUnauthorized, errorBody("not authorised")), nil
		}
	} else {
		return response(http.StatusUnauthorized, errorBody("not authorised")), nil
	}

	workouts, err := parseAndValidateMany(req.Body, userID)
	if err != nil {
		return response(http.StatusBadRequest, errorBody(err.Error())), nil
	}

	failed, err := h.saveWorkouts(ctx, workouts)
	if err != nil {
		return response(http.StatusInternalServerError, errorBody("failed to save workouts")), err
	}

	results := make([]SaveResult, len(workouts))
	for i, w := range workouts {
		if failed[w.WorkoutID] {
			results[i] = SaveResult{
				WorkoutID: w.WorkoutID,
				Status:    "error",
				Error:     "failed to save workout",
			}
		} else {
			results[i] = SaveResult{
				WorkoutID: w.WorkoutID,
				Status:    "saved",
			}
		}
	}

	return response(http.StatusCreated, map[string]any{
		"message": "batch complete",
		"results": results,
	}), nil
}

// saveWorkouts writes all workouts using DynamoDB BatchWriteItem, chunking into
// groups of 25 and retrying UnprocessedItems with exponential backoff.
// It returns a set of workoutIDs that could not be saved.
func (h *Handler) saveWorkouts(ctx context.Context, workouts []Workout) (map[string]bool, error) {
	// Marshal all workouts into WriteRequests, indexed by workoutID for tracking.
	requests, idByItem, err := h.buildWriteRequests(workouts)
	if err != nil {
		return nil, err
	}

	// Process in chunks of maxBatchSize.
	var allUnprocessed []types.WriteRequest
	for i := 0; i < len(requests); i += maxBatchSize {
		end := i + maxBatchSize
		if end > len(requests) {
			end = len(requests)
		}
		chunk := requests[i:end]

		unprocessed, err := h.batchWriteWithRetry(ctx, chunk)
		if err != nil {
			return nil, err
		}
		allUnprocessed = append(allUnprocessed, unprocessed...)
	}

	// Build the set of failed workoutIDs from any remaining unprocessed items.
	failed := make(map[string]bool)
	for _, req := range allUnprocessed {
		if req.PutRequest != nil {
			if id, ok := idByItem[itemKey(req.PutRequest.Item)]; ok {
				failed[id] = true
			}
		}
	}

	return failed, nil
}

// buildWriteRequests marshals each Workout into a DynamoDB WriteRequest.
// It also returns a mapping from the item's partition+sort key to its workoutID
// so we can identify unprocessed items later.
func (h *Handler) buildWriteRequests(workouts []Workout) ([]types.WriteRequest, map[string]string, error) {
	requests := make([]types.WriteRequest, 0, len(workouts))
	idByItem := make(map[string]string, len(workouts))

	for i := range workouts {
		item, err := attributevalue.MarshalMap(&workouts[i])
		if err != nil {
			return nil, nil, fmt.Errorf("marshal workout %d: %w", i, err)
		}

		requests = append(requests, types.WriteRequest{
			PutRequest: &types.PutRequest{Item: item},
		})
		idByItem[itemKey(item)] = workouts[i].WorkoutID
	}

	return requests, idByItem, nil
}

// batchWriteWithRetry calls BatchWriteItem and retries any UnprocessedItems
// with exponential backoff up to maxRetries times.
func (h *Handler) batchWriteWithRetry(ctx context.Context, requests []types.WriteRequest) ([]types.WriteRequest, error) {
	pending := requests

	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			delay := baseRetryDelay * time.Duration(math.Pow(2, float64(attempt-1)))
			select {
			case <-ctx.Done():
				return pending, ctx.Err()
			case <-time.After(delay):
			}
		}

		out, err := h.db.BatchWriteItem(ctx, &dynamodb.BatchWriteItemInput{
			RequestItems: map[string][]types.WriteRequest{
				h.tableName: pending,
			},
		})
		if err != nil {
			return nil, fmt.Errorf("batch write item: %w", err)
		}

		unprocessed := out.UnprocessedItems[h.tableName]
		if len(unprocessed) == 0 {
			return nil, nil
		}

		pending = unprocessed
	}

	// Exhausted retries — return whatever is still unprocessed.
	return pending, nil
}

// itemKey produces a string key from the userId and sk attributes of a DynamoDB item,
// used to correlate unprocessed items back to their workoutIDs.
func itemKey(item map[string]types.AttributeValue) string {
	var pk, sk string
	if v, ok := item["userId"].(*types.AttributeValueMemberS); ok {
		pk = v.Value
	}
	if v, ok := item["sk"].(*types.AttributeValueMemberS); ok {
		sk = v.Value
	}
	return pk + "|" + sk
}

func parseAndValidateMany(body string, userID string) ([]Workout, error) {
	trimmed := strings.TrimSpace(body)
	if trimmed == "" {
		return nil, errors.New("request body is empty")
	}

	if trimmed[0] == '[' {
		var workouts []Workout
		if err := json.Unmarshal([]byte(trimmed), &workouts); err != nil {
			return nil, fmt.Errorf("invalid JSON: %w", err)
		}
		if len(workouts) == 0 {
			return nil, errors.New("workout array is empty")
		}
		for i := range workouts {
			if err := validate(&workouts[i], userID); err != nil {
				return nil, fmt.Errorf("workouts[%d]: %w", i, err)
			}
		}
		return workouts, nil
	}

	var w Workout
	if err := json.Unmarshal([]byte(trimmed), &w); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}
	if err := validate(&w, userID); err != nil {
		return nil, err
	}
	return []Workout{w}, nil
}

func generateSK(w *Workout) {
	w.SK = fmt.Sprintf("WORKOUT#%s#%s", w.StartedAt, w.WorkoutID)
}

func validate(w *Workout, userID string) error {
	var missing []string
	w.UserID = userID;

	if w.WorkoutID == "" {
		missing = append(missing, "workoutId")
	}
	if w.StartedAt == "" {
		missing = append(missing, "startedAt")
	}
	if w.CreatedAt == "" {
		missing = append(missing, "createdAt")
	}
	if w.UpdatedAt == "" {
		missing = append(missing, "updatedAt")
	}

	generateSK(w)

	if len(missing) > 0 {
		return fmt.Errorf("missing required fields: %s", strings.Join(missing, ", "))
	}

	for i, ex := range w.Exercises {
		if ex.ExerciseID == "" {
			return fmt.Errorf("exercises[%d]: missing exerciseId", i)
		}
		if ex.Name == "" {
			return fmt.Errorf("exercises[%d]: missing name", i)
		}
		for j, s := range ex.Sets {
			if s.SetID == "" {
				return fmt.Errorf("exercises[%d].sets[%d]: missing setId", i, j)
			}
		}
	}

	if w.BodyWeight != nil {
		if w.BodyWeight.Unit != "kg" && w.BodyWeight.Unit != "lbs" {
			return errors.New("bodyWeight.unit must be \"kg\" or \"lbs\"")
		}
	}

	return nil
}

// response builds an APIGatewayProxyResponse with a JSON body.
func response(statusCode int, body any) events.APIGatewayProxyResponse {
	b, _ := json.Marshal(body)
	return events.APIGatewayProxyResponse{
		StatusCode: statusCode,
		Headers:    map[string]string{"Content-Type": "application/json"},
		Body:       string(b),
	}
}

// errorBody produces a simple JSON error object.
func errorBody(msg string) map[string]string {
	return map[string]string{"error": msg}
}
