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

const pageSize = 20

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
// Optional query parameter "sk" is used as an exclusive start key for pagination.
// Results are returned in descending SK order (most recent first).
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

	skCursor := req.QueryStringParameters["sk"]

	workouts, nextSK, err := h.queryWorkouts(ctx, userID, skCursor)
	if err != nil {
		return response(http.StatusInternalServerError, errorBody("failed to read workouts")), err
	}

	return response(http.StatusOK, map[string]any{
		"workouts": workouts,
		"nextSk":   nextSK,
	}), nil
}

// queryWorkouts runs a DynamoDB Query for the given user, returning up to pageSize
// items in descending SK order. When skCursor is non-empty it is used as an
// exclusive start key so the next page begins just before that SK value.
func (h *Handler) queryWorkouts(ctx context.Context, userID, skCursor string) ([]Workout, *string, error) {
	input := &dynamodb.QueryInput{
		TableName:              aws.String(h.tableName),
		KeyConditionExpression: aws.String("userId = :uid AND begins_with(sk, :prefix)"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":uid":    &types.AttributeValueMemberS{Value: userID},
			":prefix": &types.AttributeValueMemberS{Value: "WORKOUT#"},
		},
		ScanIndexForward: aws.Bool(false), // descending
		Limit:            aws.Int32(pageSize),
	}

	// If a cursor SK was supplied, set it as the exclusive start key so DynamoDB
	// begins scanning from the item just after (before in descending order) that SK.
	if skCursor != "" {
		input.ExclusiveStartKey = map[string]types.AttributeValue{
			"userId": &types.AttributeValueMemberS{Value: userID},
			"sk":     &types.AttributeValueMemberS{Value: skCursor},
		}
	}

	out, err := h.db.Query(ctx, input)
	if err != nil {
		return nil, nil, fmt.Errorf("dynamodb query: %w", err)
	}

	workouts := make([]Workout, 0, len(out.Items))
	if err := attributevalue.UnmarshalListOfMaps(out.Items, &workouts); err != nil {
		return nil, nil, fmt.Errorf("unmarshal items: %w", err)
	}

	// If DynamoDB returned a LastEvaluatedKey there are more pages; extract its SK
	// value so the caller can use it as the next cursor.
	var nextSK *string
	if lek := out.LastEvaluatedKey; len(lek) > 0 {
		if v, ok := lek["sk"].(*types.AttributeValueMemberS); ok {
			nextSK = &v.Value
		}
	}

	return workouts, nextSK, nil
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
