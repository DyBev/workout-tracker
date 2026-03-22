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
	maxBatchSize   = 25
	maxRetries     = 3
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
// The request body may be a single SavedExercise object or an array of SavedExercises.
func (h *Handler) HandleRequest(
	ctx context.Context,
	req events.APIGatewayProxyRequest,
) (events.APIGatewayProxyResponse, error) {
	if h.tableName == "" {
		return response(http.StatusInternalServerError, errorBody("table name not configured")), nil
	}

	var userID string
	if req.RequestContext.Authorizer != nil {
		jwt, ok := req.RequestContext.Authorizer["jwt"]
		if !ok {
			return response(http.StatusUnauthorized, errorBody("not authorised")), nil
		}
		claims, ok := jwt.(map[string]any)["claims"]
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

	exercises, err := parseAndValidateMany(req.Body, userID)
	if err != nil {
		return response(http.StatusBadRequest, errorBody(err.Error())), nil
	}

	failed, err := h.saveExercises(ctx, exercises)
	if err != nil {
		return response(http.StatusInternalServerError, errorBody("failed to save exercises")), err
	}

	results := make([]SaveResult, len(exercises))
	for i, ex := range exercises {
		if failed[ex.SavedExerciseID] {
			results[i] = SaveResult{
				SavedExerciseID: ex.SavedExerciseID,
				Status:          "error",
				Error:           "failed to save exercise",
			}
		} else {
			results[i] = SaveResult{
				SavedExerciseID: ex.SavedExerciseID,
				Status:          "saved",
			}
		}
	}

	return response(http.StatusCreated, map[string]any{
		"message": "batch complete",
		"results": results,
	}), nil
}

// saveExercises writes all exercises using DynamoDB BatchWriteItem, chunking into
// groups of 25 and retrying UnprocessedItems with exponential backoff.
// It returns a set of savedExerciseIDs that could not be saved.
func (h *Handler) saveExercises(ctx context.Context, exercises []SavedExercise) (map[string]bool, error) {
	requests, idByItem, err := h.buildWriteRequests(exercises)
	if err != nil {
		return nil, err
	}

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

// buildWriteRequests marshals each SavedExercise into a DynamoDB WriteRequest.
func (h *Handler) buildWriteRequests(exercises []SavedExercise) ([]types.WriteRequest, map[string]string, error) {
	requests := make([]types.WriteRequest, 0, len(exercises))
	idByItem := make(map[string]string, len(exercises))

	for i := range exercises {
		item, err := attributevalue.MarshalMap(&exercises[i])
		if err != nil {
			return nil, nil, fmt.Errorf("marshal exercise %d: %w", i, err)
		}

		requests = append(requests, types.WriteRequest{
			PutRequest: &types.PutRequest{Item: item},
		})
		idByItem[itemKey(item)] = exercises[i].SavedExerciseID
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

	return pending, nil
}

// itemKey produces a string key from the userId and sk attributes of a DynamoDB item,
// used to correlate unprocessed items back to their savedExerciseIDs.
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

func parseAndValidateMany(body string, userID string) ([]SavedExercise, error) {
	trimmed := strings.TrimSpace(body)
	if trimmed == "" {
		return nil, errors.New("request body is empty")
	}

	if trimmed[0] == '[' {
		var exercises []SavedExercise
		if err := json.Unmarshal([]byte(trimmed), &exercises); err != nil {
			return nil, fmt.Errorf("invalid JSON: %w", err)
		}
		if len(exercises) == 0 {
			return nil, errors.New("exercise array is empty")
		}
		for i := range exercises {
			if err := validate(&exercises[i], userID); err != nil {
				return nil, fmt.Errorf("exercises[%d]: %w", i, err)
			}
		}
		return exercises, nil
	}

	var ex SavedExercise
	if err := json.Unmarshal([]byte(trimmed), &ex); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}
	if err := validate(&ex, userID); err != nil {
		return nil, err
	}
	return []SavedExercise{ex}, nil
}

func validate(ex *SavedExercise, userID string) error {
	ex.UserID = userID
	ex.SK = fmt.Sprintf("EXERCISE#%s", ex.SavedExerciseID)

	var missing []string
	if ex.SavedExerciseID == "" {
		missing = append(missing, "savedExerciseId")
	}
	if ex.Name == "" {
		missing = append(missing, "name")
	}
	if ex.CreatedAt == "" {
		missing = append(missing, "createdAt")
	}
	if ex.UpdatedAt == "" {
		missing = append(missing, "updatedAt")
	}
	if len(missing) > 0 {
		return fmt.Errorf("missing required fields: %s", strings.Join(missing, ", "))
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
