package handler

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider/types"
)

type CognitoUserCreator interface {
	AdminCreateUser(ctx context.Context, params *cognitoidentityprovider.AdminCreateUserInput, optFns ...func(*cognitoidentityprovider.Options)) (*cognitoidentityprovider.AdminCreateUserOutput, error)
}

type Handler struct {
	cognito    CognitoUserCreator
	userPoolID string
	sendInvites bool
}

func NewHandler(cognito CognitoUserCreator, userPoolID string, sendInvites bool) *Handler {
	return &Handler{cognito: cognito, userPoolID: userPoolID, sendInvites: sendInvites}
}

func (h *Handler) HandleRequest(
	ctx context.Context,
	req events.APIGatewayProxyRequest,
) (events.APIGatewayProxyResponse, error) {
	htmx := isHtmxRequest(req)

	if h.userPoolID == "" {
		return errorResponse(http.StatusInternalServerError, "user pool ID not configured", htmx), nil
	}

	registerReq, err := parseRequest(req.Body)
	if err != nil {
		return errorResponse(http.StatusBadRequest, err.Error(), htmx), nil
	}

	err = h.createUser(ctx, registerReq.Email)
	if err != nil {
		var usernameExists *types.UsernameExistsException
		if errors.As(err, &usernameExists) {
			return successResponse(htmx), nil
		}
		return errorResponse(http.StatusInternalServerError, "failed to register interest", htmx), err
	}

	return successResponse(htmx), nil
}

func (h *Handler) createUser(ctx context.Context, email string) error {
	input := &cognitoidentityprovider.AdminCreateUserInput{
		UserPoolId: aws.String(h.userPoolID),
		Username:   aws.String(email),
		UserAttributes: []types.AttributeType{
			{
				Name:  aws.String("email"),
				Value: aws.String(email),
			},
			{
				Name:  aws.String("email_verified"),
				Value: aws.String("true"),
			},
		},
		DesiredDeliveryMediums: []types.DeliveryMediumType{
			types.DeliveryMediumTypeEmail,
		},
	}

	if !h.sendInvites {
		input.MessageAction = types.MessageActionTypeSuppress
	}

	_, err := h.cognito.AdminCreateUser(ctx, input)
	return err
}

func parseRequest(body string) (RegisterRequest, error) {

	decodedBytes, err := base64.StdEncoding.DecodeString(body)
	if err != nil {
		fmt.Printf("Error decoding base64 string: %v\n", err)
		return RegisterRequest{}, errors.New("decoding base64 string")
	}
	decodedBody := string(decodedBytes)

	trimmed := strings.TrimSpace(decodedBody)
	if trimmed == "" {
		return RegisterRequest{}, errors.New("request body is empty")
	}

	vals, err := url.ParseQuery(trimmed)
	if err != nil {
		return RegisterRequest{}, fmt.Errorf("invalid form body: %w", err)
	}

	req := RegisterRequest{Email: vals.Get("email")}
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	if req.Email == "" {
		return RegisterRequest{}, errors.New("missing required field: email")
	}
	if !isValidEmail(req.Email) {
		return RegisterRequest{}, errors.New("invalid email address")
	}

	return req, nil
}

func isValidEmail(email string) bool {
	at := strings.Index(email, "@")
	if at < 1 {
		return false
	}
	domain := email[at+1:]
	dot := strings.LastIndex(domain, ".")
	return dot > 0 && dot < len(domain)-1
}

func isHtmxRequest(req events.APIGatewayProxyRequest) bool {
	return req.Headers["hx-request"] == "true" || req.Headers["HX-Request"] == "true"
}

func successResponse(htmx bool) events.APIGatewayProxyResponse {
	if htmx {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusOK,
			Headers:    map[string]string{"Content-Type": "text/html"},
			Body:       `<p class="register-success">Thanks for registering! An email will be sent when we are ready for you to join.</p>`,
		}
	}
	return jsonResponse(http.StatusOK, map[string]string{
		"message": "thanks for registering your interest",
	})
}

func errorResponse(statusCode int, msg string, htmx bool) events.APIGatewayProxyResponse {
	if htmx {
		return events.APIGatewayProxyResponse{
			StatusCode: statusCode,
			Headers:    map[string]string{"Content-Type": "text/html"},
			Body:       fmt.Sprintf(`<p class="register-error">%s</p>`, msg),
		}
	}
	return jsonResponse(statusCode, errorBody(msg))
}

func jsonResponse(statusCode int, body any) events.APIGatewayProxyResponse {
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
