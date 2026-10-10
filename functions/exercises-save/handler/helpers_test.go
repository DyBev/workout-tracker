package handler

import (
	"encoding/json"
	"testing"

	"github.com/aws/aws-lambda-go/events"
)

func findExercise(array []SaveResult, val string) bool {
	for _, value := range array {
		if value.SavedExerciseID == val {
			return true;
		}
	}

	return false;
}

func makeAuthedRequest(userID string, body string) events.APIGatewayProxyRequest {
	return events.APIGatewayProxyRequest{
		Body: body,
		RequestContext: events.APIGatewayProxyRequestContext{
			Authorizer: map[string]interface{}{
				"jwt": map[string]interface{}{
					"claims": map[string]interface{}{
						"sub": userID,
					},
				},
			},
		},
	}
}

func makeUnauthedRequest(body string) events.APIGatewayProxyRequest {
	return events.APIGatewayProxyRequest{
		Body: body,
	}
}

type ResponseBody struct {
	Message string `json:"message,omitempty"`
	Error string `json:"error,omitempty"`
	Resutlts []SaveResult `json:"results,omitempty"`
}

func parseResponseBody(t *testing.T, body string) ResponseBody {
	t.Helper()
	var m ResponseBody
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

func validExerciseJSON() string {
	ex := SavedExercise{
		SavedExerciseID: "abc123",
		Name:            "Bench Press",
		Note:            "Keep elbows tucked",
		Tags:            []string{"chest", "push"},
		CreatedAt:       "2026-03-01T10:00:00.000Z",
		UpdatedAt:       "2026-03-15T14:30:00.000Z",
	}
	b, _ := json.Marshal(ex)
	return string(b)
}

func validExerciseArrayJSON() string {
	exercises := []SavedExercise{
		{
			SavedExerciseID: "abc123",
			Name:            "Bench Press",
			Note:            "Keep elbows tucked",
			Tags:            []string{"chest", "push"},
			CreatedAt:       "2026-03-01T10:00:00.000Z",
			UpdatedAt:       "2026-03-15T14:30:00.000Z",
		},
	}
	b, _ := json.Marshal(exercises)
	return string(b)
}

func multipleExercisesJSON() string {
	exercises := []SavedExercise{
		{
			SavedExerciseID: "abc123",
			Name:            "Bench Press",
			Note:            "Keep elbows tucked",
			Tags:            []string{"chest", "push"},
			CreatedAt:       "2026-03-01T10:00:00.000Z",
			UpdatedAt:       "2026-03-15T14:30:00.000Z",
		},
		{
			SavedExerciseID: "def456",
			Name:            "Squat",
			Note:            "Full depth",
			Tags:            []string{"legs"},
			CreatedAt:       "2026-03-02T10:00:00.000Z",
			UpdatedAt:       "2026-03-16T14:30:00.000Z",
		},
	}
	b, _ := json.Marshal(exercises)
	return string(b)
}

func makeRequestWithContext(
	method string,
	body []SavedExercise,
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

func parseErrorResponseBody(t *testing.T, body string) map[string]string {
	t.Helper()
	var m map[string]string
	if err := json.Unmarshal([]byte(body), &m); err != nil {
		t.Fatalf("failed to parse response body: %v", err)
	}
	return m
}
