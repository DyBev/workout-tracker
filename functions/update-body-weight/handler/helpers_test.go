package handler

import (
	"encoding/json"
	"testing"

	"github.com/aws/aws-lambda-go/events"
)

var validRequestBody = RequestBody{
	WorkoutID: "mockWorkoutID",
	BodyWeight: 80,
}

func makeRequestWithContext(
	method string,
	body RequestBody,
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

func makeRequest(method string, body RequestBody) events.APIGatewayProxyRequest {
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

