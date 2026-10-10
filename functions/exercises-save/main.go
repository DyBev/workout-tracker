package main

import (
	"context"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"

	"dybev.uk/workout-tracker/functions/exercises-save/handler"
)

func main() {
	tableName := os.Getenv("TABLE_NAME")

	cfg, err := config.LoadDefaultConfig(context.Background(), config.WithRegion("eu-west-2"))
	if err != nil {
		log.Fatalf("unable to load AWS config: %v", err)
	}

	client := dynamodb.NewFromConfig(cfg)
	attributeMarshalMap := attributevalue.MarshalMap
	h := handler.NewHandler(client, attributeMarshalMap, tableName)

	lambda.Start(h.HandleRequest)
}
