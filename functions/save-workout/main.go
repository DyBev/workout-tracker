package main

import (
	"context"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

func main() {
	tableName := os.Getenv("TABLE_NAME")
	if tableName == "" {
		tableName = "WorkoutTable"
	}

	cfg, err := config.LoadDefaultConfig(context.Background(), config.WithRegion("eu-west-2"))
	if err != nil {
		log.Fatalf("unable to load AWS config: %v", err)
	}

	client := dynamodb.NewFromConfig(cfg)
	h := NewHandler(client, tableName)

	lambda.Start(h.HandleRequest)
}
