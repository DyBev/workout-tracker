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
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type DynamoClient interface {
	UpdateItem(
		ctx context.Context,
		params *dynamodb.UpdateItemInput,
		optFns ...func(*dynamodb.Options),
	) (*dynamodb.UpdateItemOutput, error)

	Query(
		ctx context.Context,
		params *dynamodb.QueryInput,
		optFns ...func(*dynamodb.Options),
	) (*dynamodb.QueryOutput, error)
}

type Handler struct {
	db        DynamoClient
	tableName string
}

func NewHandler(db DynamoClient, tableName string) *Handler {
	return &Handler{db: db, tableName: tableName}
}

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

	workoutID, bodyWeight, err := parseAndValidateBody(req.Body)
	if err != nil {
		return response(http.StatusBadRequest, errorBody(err.Error())), nil
	}

	err = h.updateBodyWeight(ctx, userID, workoutID, bodyWeight)
	if err != nil {
		return response(http.StatusInternalServerError, errorBody("failed to save workouts")), err
	}

	return response(204, map[string]any{}), nil
}

type WorkoutQueryOutput struct {
	UserID string `dynamodbav:"userId"`
	SK string `dynamodbav:"sk"`
}

func (h *Handler) updateBodyWeight(
	ctx context.Context,
	userID string,
	workoutID string,
	bodyWeight float32,
) (error) {
	output, err := h.db.Query(ctx, &dynamodb.QueryInput{
		TableName: &h.tableName,
		IndexName: aws.String("workoutId"),
		KeyConditionExpression: aws.String("userId = :userID AND workoutId = :workoutID"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":userID": &types.AttributeValueMemberS{ Value: userID },
			":workoutID": &types.AttributeValueMemberS{ Value: workoutID },
		},
	})
	if err != nil {
		return err
	}

	if output.Count != 1 {
		return errors.New("invalid output count")
	}

	var indexOutput WorkoutQueryOutput
	if err = attributevalue.UnmarshalMap(output.Items[0], &indexOutput); err != nil {
		return err
	}

	_, err = h.db.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: &h.tableName,
		Key: map[string]types.AttributeValue{
			"userId": &types.AttributeValueMemberS{ Value: indexOutput.UserID },
			"sk": &types.AttributeValueMemberS{ Value: indexOutput.SK },
		},
		UpdateExpression: aws.String(`
			SET bodyWeight = :bodyWeight
		`),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":bodyWeight": &types.AttributeValueMemberN{ Value: fmt.Sprintf("%v", bodyWeight) },
			":deadline": &types.AttributeValueMemberS{ Value: time.Now().Add(time.Duration(-6) * time.Hour).Format(time.RFC3339) },
		},
		ConditionExpression: aws.String(`
			completedAt > :deadline
		`),
	})

	return err
}

type RequestBody struct {
	WorkoutID string `json:"workoutId"`
	BodyWeight float32 `json:"bodyweight"`
}
func parseAndValidateBody(body string) (string, float32, error) {
	trimmed := strings.TrimSpace(body)
	if trimmed == "" {
		return "", 0, errors.New("request body is empty")
	}

	var value RequestBody
	if err := json.Unmarshal([]byte(trimmed), &value); err != nil {
		return "", 0, fmt.Errorf("invalid JSON: %w", err)
	}
	if err := validate(value); err != nil {
		return "", 0, errors.New("invalid request body")
	}

	return value.WorkoutID, value.BodyWeight, nil
}

func validate(value RequestBody) error {
	if value.WorkoutID == "" || value.BodyWeight < 0 {
		return errors.New("invalid body")
	}
	return nil
}

func response(statusCode int, body any) events.APIGatewayProxyResponse {
	if statusCode == 204 {
		return events.APIGatewayProxyResponse{
			StatusCode: statusCode,
		}
	}

	b, _ := json.Marshal(body)
	return events.APIGatewayProxyResponse{
		StatusCode: statusCode,
		Headers:    map[string]string{"Content-Type": "application/json"},
		Body:       string(b),
	}
}

func errorBody(msg string) map[string]string {
	return map[string]string{"error": msg}
}
