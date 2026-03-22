package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// DynamoQuerier is the subset of the DynamoDB client that this handler needs.
// Using an interface makes the handler easy to test with a mock.
type DynamoQuerier interface {
	Query(ctx context.Context, params *dynamodb.QueryInput, optFns ...func(*dynamodb.Options)) (*dynamodb.QueryOutput, error)
}

// Handler holds the dependencies injected at startup.
type Handler struct {
	db        DynamoQuerier
	tableName string
}

// NewHandler creates a Handler with the given DynamoDB client and table name.
func NewHandler(db DynamoQuerier, tableName string) *Handler {
	return &Handler{db: db, tableName: tableName}
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

	exercises, err := h.queryExercises(ctx, userID)
	if err != nil {
		return response(http.StatusInternalServerError, errorBody("failed to read exercises")), err
	}

	return response(http.StatusOK, map[string]any{
		"exercises": exercises,
	}), nil
}

// queryExercises runs a DynamoDB Query for all saved exercises belonging to the given user.
func (h *Handler) queryExercises(ctx context.Context, userID string) ([]SavedExercise, error) {
	input := &dynamodb.QueryInput{
		TableName:              aws.String(h.tableName),
		KeyConditionExpression: aws.String("userId = :uid AND begins_with(sk, :prefix)"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":uid":    &types.AttributeValueMemberS{Value: userID},
			":prefix": &types.AttributeValueMemberS{Value: "EXERCISE#"},
		},
	}

	out, err := h.db.Query(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("dynamodb query: %w", err)
	}

	exercises := make([]SavedExercise, 0, len(out.Items))
	if err := attributevalue.UnmarshalListOfMaps(out.Items, &exercises); err != nil {
		return nil, fmt.Errorf("unmarshal items: %w", err)
	}

	return exercises, nil
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
