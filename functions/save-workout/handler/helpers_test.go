package handler

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/aws/aws-lambda-go/events"
)

func findWorkout(array []SaveResult, val string) bool {
	for _, value := range array {
		if value.WorkoutID == val {
			return true;
		}
	}

	return false;
}

func validWorkout(id string, note string) Workout {
	reps := 10
	weight := 80.0
	bw := float32(75.5)
	w := Workout{
		WorkoutID:  id,
		StartedAt:  "2026-03-14T10:00:00.000Z",
		Notes:      note,
		Tags:       []string{"push", "chest"},
		BodyWeight: &bw,
		Exercises: []WorkoutExercise{
			{
				ExerciseID: "ex-1",
				Name:       "Bench Press",
				Order:      0,
				Sets: []WorkoutSet{
					{SetID: "s-1", Order: 0, Reps: &reps, Weight: &weight, WeightUnit: "kg"},
					{SetID: "s-2", Order: 1, Reps: &reps, Weight: nil, WeightUnit: "lbs"},
				},
			},
		},
		CreatedAt: "2026-03-14T10:00:00.000Z",
		UpdatedAt: "2026-03-14T10:30:00.000Z",
	}
	return w
}

func validWorkoutWithSet(id string, exercises []WorkoutExercise) Workout {
	bw := float32(75.5)
	w := Workout{
		WorkoutID:  id,
		StartedAt:  "2026-03-14T10:00:00.000Z",
		Notes:      "Great session",
		Tags:       []string{"push", "chest"},
		BodyWeight: &bw,
		Exercises: exercises,
		CreatedAt: "2026-03-14T10:00:00.000Z",
		UpdatedAt: "2026-03-14T10:30:00.000Z",
	}
	return w
}

func validExercise(id string, order int, sets []WorkoutSet) WorkoutExercise {
	return WorkoutExercise{
		ExerciseID: id,
		Sets: sets,
		Name: "mock Exercise Name",
		Order: order,
	}
}

func validExerciseWithNote(id string, order int, sets []WorkoutSet, note string) WorkoutExercise {
	return WorkoutExercise{
		ExerciseID: id,
		Sets: sets,
		Name: "mock Exercise Name",
		Order: order,
		Note: &note,
	}
}

func invalidExerciseMissingName(id string, order int, sets []WorkoutSet) WorkoutExercise {
	return WorkoutExercise{
		ExerciseID: id,
		Sets: sets,
		Name: "",
		Order: order,
	}
}

func longNote() *string {
	string := strings.Repeat("0", 1002)
	return &string
}

func invalidExerciseLongNote(id string, order int, sets []WorkoutSet) WorkoutExercise {
	return WorkoutExercise{
		ExerciseID: id,
		Sets: sets,
		Name: "mock exercise name",
		Order: order,
		Note: longNote(),
	}
}

func validSet(id string) WorkoutSet {
	var weight float64 = 10
	reps := 10
	return WorkoutSet{
		SetID: id,
		Weight: &weight,
		Reps: &reps,
		Order: 1,
		WeightUnit: "kg",
	}
}

func minimalWorkout(id string) Workout {
	w := Workout{
		WorkoutID: id,
		StartedAt: "2026-03-14T12:00:00.000Z",
		Notes:     "",
		Tags:      []string{},
		Exercises: []WorkoutExercise{},
		CreatedAt: "2026-03-14T12:00:00.000Z",
		UpdatedAt: "2026-03-14T12:00:00.000Z",
	}
	return w
}

func invalidWorkoutNoID() Workout {
	w := Workout{
		WorkoutID: "",
		StartedAt: "2026-03-14T12:00:00.000Z",
		Notes:     "",
		Tags:      []string{},
		Exercises: []WorkoutExercise{},
		CreatedAt: "2026-03-14T12:00:00.000Z",
		UpdatedAt: "2026-03-14T12:00:00.000Z",
	}
	return w
}

func invalidWorkoutNoStarted(id string) Workout {
	w := Workout{
		WorkoutID: id,
		StartedAt: "",
		Notes:     "",
		Tags:      []string{},
		Exercises: []WorkoutExercise{},
		CreatedAt: "2026-03-14T12:00:00.000Z",
		UpdatedAt: "2026-03-14T12:00:00.000Z",
	}
	return w
}

func invalidWorkoutNoCreatedAt(id string) Workout {
	w := Workout{
		WorkoutID: id,
		StartedAt: "2026-03-14T12:00:00.000Z",
		Notes:     "",
		Tags:      []string{},
		Exercises: []WorkoutExercise{},
		CreatedAt: "",
		UpdatedAt: "2026-03-14T12:00:00.000Z",
	}
	return w
}

func invalidWorkoutNoUpdatedAt(id string) Workout {
	w := Workout{
		WorkoutID: id,
		StartedAt: "2026-03-14T12:00:00.000Z",
		Notes:     "",
		Tags:      []string{},
		Exercises: []WorkoutExercise{},
		CreatedAt: "2026-03-14T12:00:00.000Z",
		UpdatedAt: "",
	}
	return w
}

func makeRequestWithContext(
	method string,
	body []Workout,
	requestContext *events.APIGatewayProxyRequestContext,
) events.APIGatewayProxyRequest {
	bodyString, _ := json.Marshal(body)
	if requestContext == nil {
		return events.APIGatewayProxyRequest{
			HTTPMethod: method,
			Body:       string(bodyString),
		}
	}

	return events.APIGatewayProxyRequest{
		HTTPMethod: method,
		Body:       string(bodyString),
		RequestContext: *requestContext,
	}
}

func makeRequestWithString(method string, body string) events.APIGatewayProxyRequest {
	return events.APIGatewayProxyRequest{
		HTTPMethod: method,
		Body: body,
		RequestContext: events.APIGatewayProxyRequestContext{
			Authorizer: map[string]any{
				"jwt": map[string]any{
					"claims": map[string]any{
						"sub": "mockSub",
					},
				},
			},
		},
	}
}

func makeRequest(method string, body []Workout) events.APIGatewayProxyRequest {
	bodyString, _ := json.Marshal(body)
	return events.APIGatewayProxyRequest{
		HTTPMethod: method,
		Body:       string(bodyString),
		RequestContext: events.APIGatewayProxyRequestContext{
			Authorizer: map[string]any{
				"jwt": map[string]any{
					"claims": map[string]any{
						"sub": "mockSub",
					},
				},
			},
		},
	}
}

func makeRequestSingleWorkout(method string, body Workout) events.APIGatewayProxyRequest {
	bodyString, _ := json.Marshal(body)
	return events.APIGatewayProxyRequest{
		HTTPMethod: method,
		Body:       string(bodyString),
		RequestContext: events.APIGatewayProxyRequestContext{
			Authorizer: map[string]any{
				"jwt": map[string]any{
					"claims": map[string]any{
						"sub": "mockSub",
					},
				},
			},
		},
	}
}

func parseErrorResponseBody(t *testing.T, body string) map[string]string {
	t.Helper()
	var m map[string]string
	if err := json.Unmarshal([]byte(body), &m); err != nil {
		t.Fatalf("failed to parse response body: %v", err)
	}
	return m
}

func parseBatchResponseBody(t *testing.T, body string) (string, []SaveResult) {
	t.Helper()
	var m struct {
		Message string       `json:"message"`
		Results []SaveResult `json:"results"`
	}
	if err := json.Unmarshal([]byte(body), &m); err != nil {
		t.Fatalf("failed to parse batch response body: %v", err)
	}
	return m.Message, m.Results
}

