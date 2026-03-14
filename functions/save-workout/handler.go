package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

// DynamoPutter is the subset of the DynamoDB client that this handler needs.
// Using an interface makes the handler easy to test with a mock.
type DynamoPutter interface {
	PutItem(ctx context.Context, params *dynamodb.PutItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error)
}

// Handler holds the dependencies injected at startup.
type Handler struct {
	db        DynamoPutter
	tableName string
}

// NewHandler creates a Handler with the given DynamoDB client and table name.
func NewHandler(db DynamoPutter, tableName string) *Handler {
	return &Handler{db: db, tableName: tableName}
}

// HandleRequest processes the API Gateway proxy request.
func (h *Handler) HandleRequest(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	if req.HTTPMethod != http.MethodPost {
		return response(http.StatusMethodNotAllowed, errorBody("method not allowed")), nil
	}

	workout, err := parseAndValidate(req.Body)
	if err != nil {
		return response(http.StatusBadRequest, errorBody(err.Error())), nil
	}

	if err := h.saveWorkout(ctx, workout); err != nil {
		return response(http.StatusInternalServerError, errorBody("failed to save workout")), err
	}

	return response(http.StatusCreated, map[string]string{
		"message":   "workout saved",
		"workoutId": workout.WorkoutID,
	}), nil
}

// parseAndValidate unmarshals the JSON body into a Workout and validates required fields.
func parseAndValidate(body string) (*Workout, error) {
	if strings.TrimSpace(body) == "" {
		return nil, errors.New("request body is empty")
	}

	var w Workout
	if err := json.Unmarshal([]byte(body), &w); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}

	if err := validate(&w); err != nil {
		return nil, err
	}

	return &w, nil
}

// validate checks that all required fields are present on the Workout.
func validate(w *Workout) error {
	var missing []string

	if w.UserID == "" {
		missing = append(missing, "userId")
	}
	if w.SK == "" {
		missing = append(missing, "sk")
	}
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
			if s.WeightUnit != "kg" && s.WeightUnit != "lbs" {
				return fmt.Errorf("exercises[%d].sets[%d]: weightUnit must be \"kg\" or \"lbs\"", i, j)
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

// saveWorkout marshals the Workout to a DynamoDB item and puts it in the table.
func (h *Handler) saveWorkout(ctx context.Context, w *Workout) error {
	item, err := attributevalue.MarshalMap(w)
	if err != nil {
		return fmt.Errorf("marshal workout: %w", err)
	}

	_, err = h.db.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: &h.tableName,
		Item:      item,
	})
	if err != nil {
		return fmt.Errorf("put item: %w", err)
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
