package main

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// DynamoGetter is the subset of the DynamoDB client that this handler needs.
// Using an interface makes the handler easy to test with a mock.
type DynamoGetter interface {
	GetItem(ctx context.Context, params *dynamodb.GetItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error)
}

// Handler holds the dependencies injected at startup.
type Handler struct {
	db        DynamoGetter
	tableName string
}

// NewHandler creates a Handler with the given DynamoDB client and table name.
func NewHandler(db DynamoGetter, tableName string) *Handler {
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

	profile, err := h.getUserProfile(ctx, userID)
	if err != nil {
		return response(http.StatusInternalServerError, errorBody("failed to read subscription status")), err
	}

	if profile == nil {
		return response(http.StatusOK, StatusResponse{Status: "none"}), nil
	}

	return response(http.StatusOK, StatusResponse{
		Status:             profile.SubscriptionStatus,
		SubscriptionSource: profile.SubscriptionSource,
		CurrentPeriodEnd:   profile.CurrentPeriodEnd,
	}), nil
}

// getUserProfile fetches the user profile from DynamoDB. Returns nil if no item exists.
func (h *Handler) getUserProfile(ctx context.Context, userID string) (*UserProfile, error) {
	result, err := h.db.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(h.tableName),
		Key: map[string]types.AttributeValue{
			"userId": &types.AttributeValueMemberS{Value: userID},
		},
	})
	if err != nil {
		return nil, err
	}

	if result.Item == nil {
		return nil, nil
	}

	var profile UserProfile
	if err := attributevalue.UnmarshalMap(result.Item, &profile); err != nil {
		return nil, err
	}

	return &profile, nil
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
