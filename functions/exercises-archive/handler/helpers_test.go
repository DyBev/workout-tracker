package handler

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/aws/aws-lambda-go/events"
)

func makeAuthedRequest(userID string, body string) events.APIGatewayProxyRequest {
	return events.APIGatewayProxyRequest{
		Body: body,
		RequestContext: events.APIGatewayProxyRequestContext{
			Authorizer: map[string]any{
				"jwt": map[string]any{
					"claims": map[string]any{
						"sub": userID,
					},
				},
			},
		},
	}
}

func makeRequestWithContext(context *events.APIGatewayProxyRequestContext) events.APIGatewayProxyRequest {
	if context == nil {
		return events.APIGatewayProxyRequest{
			Body: `{"savedExerciseId":"abc123","archive":true}`,
		}
	}

	return events.APIGatewayProxyRequest{
		Body: `{"savedExerciseId":"abc123","archive":true}`,
		RequestContext: *context,
	}
}

func makeUnauthedRequest(body string) events.APIGatewayProxyRequest {
	return events.APIGatewayProxyRequest{
		Body: body,
	}
}

func parseResponseBody(t *testing.T, body string) map[string]any {
	t.Helper()
	var m map[string]any
	if err := json.Unmarshal([]byte(body), &m); err != nil {
		t.Fatalf("failed to parse response body: %v", err)
	}
	return m
}

func fixedTime() time.Time {
	return time.Date(2026, 3, 20, 12, 0, 0, 0, time.UTC)
}

