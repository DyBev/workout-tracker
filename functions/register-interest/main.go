package main

import (
	"context"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
	"dybev.uk/workout-tracker/functions/register-interest/handler"
)

func main() {
	userPoolID := os.Getenv("USER_POOL_ID")

	cfg, err := config.LoadDefaultConfig(context.Background(), config.WithRegion("eu-west-2"))
	if err != nil {
		log.Fatalf("unable to load AWS config: %v", err)
	}

	client := cognitoidentityprovider.NewFromConfig(cfg)

	sendInvites := true
	if v := os.Getenv("SEND_INVITES"); v == "false" || v == "0" {
		sendInvites = false
	}

	h := handler.NewHandler(client, userPoolID, sendInvites)

	lambda.Start(h.HandleRequest)
}
