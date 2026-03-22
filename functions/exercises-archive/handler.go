package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// DynamoUpdater is the subset of the DynamoDB client that this handler needs.
// Using an interface makes the handler easy to test with a mock.
type DynamoUpdater interface {
	UpdateItem(ctx context.Context, params *dynamodb.UpdateItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.UpdateItemOutput, error)
}

// Handler holds the dependencies injected at startup.
type Handler struct {
	db        DynamoUpdater
	tableName string
	now       func() time.Time
}

// NewHandler creates a Handler with the given DynamoDB client and table name.
func NewHandler(db DynamoUpdater, tableName string) *Handler {
	return &Handler{db: db, tableName: tableName, now: time.Now}
}

// HandleRequest processes the API Gateway proxy request.
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

	archiveReq, err := parseRequest(req.Body)
	if err != nil {
		return response(http.StatusBadRequest, errorBody(err.Error())), nil
	}

	err = h.updateExercise(ctx, userID, archiveReq)
	if err != nil {
		// Check if this is a ConditionalCheckFailedException (item not found).
		var condErr *types.ConditionalCheckFailedException
		if errors.As(err, &condErr) {
			return response(http.StatusNotFound, errorBody("exercise not found")), nil
		}
		return response(http.StatusInternalServerError, errorBody("failed to update exercise")), err
	}

	msg := "exercise archived"
	if !archiveReq.Archive {
		msg = "exercise restored"
	}
	return response(http.StatusOK, map[string]string{"message": msg}), nil
}

// updateExercise performs the DynamoDB UpdateItem to archive or restore an exercise.
func (h *Handler) updateExercise(ctx context.Context, userID string, req ArchiveRequest) error {
	sk := fmt.Sprintf("EXERCISE#%s", req.SavedExerciseID)
	now := h.now().UTC().Format(time.RFC3339Nano)

	key := map[string]types.AttributeValue{
		"userId": &types.AttributeValueMemberS{Value: userID},
		"sk":     &types.AttributeValueMemberS{Value: sk},
	}

	var input *dynamodb.UpdateItemInput

	if req.Archive {
		input = &dynamodb.UpdateItemInput{
			TableName:           aws.String(h.tableName),
			Key:                 key,
			UpdateExpression:    aws.String("SET archivedAt = :archivedAt, updatedAt = :updatedAt"),
			ConditionExpression: aws.String("attribute_exists(sk)"),
			ExpressionAttributeValues: map[string]types.AttributeValue{
				":archivedAt": &types.AttributeValueMemberS{Value: now},
				":updatedAt":  &types.AttributeValueMemberS{Value: now},
			},
		}
	} else {
		input = &dynamodb.UpdateItemInput{
			TableName:           aws.String(h.tableName),
			Key:                 key,
			UpdateExpression:    aws.String("SET updatedAt = :updatedAt REMOVE archivedAt"),
			ConditionExpression: aws.String("attribute_exists(sk)"),
			ExpressionAttributeValues: map[string]types.AttributeValue{
				":updatedAt": &types.AttributeValueMemberS{Value: now},
			},
		}
	}

	_, err := h.db.UpdateItem(ctx, input)
	return err
}

func parseRequest(body string) (ArchiveRequest, error) {
	trimmed := strings.TrimSpace(body)
	if trimmed == "" {
		return ArchiveRequest{}, errors.New("request body is empty")
	}

	var req ArchiveRequest
	if err := json.Unmarshal([]byte(trimmed), &req); err != nil {
		return ArchiveRequest{}, fmt.Errorf("invalid JSON: %w", err)
	}
	if req.SavedExerciseID == "" {
		return ArchiveRequest{}, errors.New("missing required field: savedExerciseId")
	}

	return req, nil
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
