package handler

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// makeAuthedRequest builds an APIGatewayProxyRequest with a fake Cognito JWT
// authorizer context for the given user ID.
func makeAuthedRequest(userID string, queryParams map[string]string) events.APIGatewayProxyRequest {
	return events.APIGatewayProxyRequest{
		QueryStringParameters: queryParams,
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

func makeUnauthedRequest() events.APIGatewayProxyRequest {
	return events.APIGatewayProxyRequest{}
}

func parseResponseBody(t *testing.T, body string) map[string]any {
	t.Helper()
	var m map[string]any
	if err := json.Unmarshal([]byte(body), &m); err != nil {
		t.Fatalf("failed to parse response body: %v", err)
	}
	return m
}

// buildWorkoutItems constructs n DynamoDB attribute maps that look like Workout items.
func buildWorkoutItems(userID string, n int) []map[string]types.AttributeValue {
	items := make([]map[string]types.AttributeValue, n)
	for i := range n {
		sk := fmt.Sprintf("WORKOUT#2026-03-%02dT10:00:00.000Z#wkt-%03d", i+1, i)
		items[i] = map[string]types.AttributeValue{
			"userId":    &types.AttributeValueMemberS{Value: userID},
			"sk":        &types.AttributeValueMemberS{Value: sk},
			"workoutId": &types.AttributeValueMemberS{Value: fmt.Sprintf("wkt-%03d", i)},
			"startedAt": &types.AttributeValueMemberS{Value: fmt.Sprintf("2026-03-%02dT10:00:00.000Z", i+1)},
			"createdAt": &types.AttributeValueMemberS{Value: fmt.Sprintf("2026-03-%02dT10:00:00.000Z", i+1)},
			"updatedAt": &types.AttributeValueMemberS{Value: fmt.Sprintf("2026-03-%02dT10:00:00.000Z", i+1)},
			"notes":     &types.AttributeValueMemberS{Value: ""},
		}
	}
	return items
}

