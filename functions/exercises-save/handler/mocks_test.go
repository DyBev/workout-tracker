package handler

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

// mockDynamo implements DynamoBatchWriter for testing.
type mockDynamo struct {
	batchWriteFunc func(ctx context.Context, params *dynamodb.BatchWriteItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.BatchWriteItemOutput, error)
	calls          int
	lastInput      *dynamodb.BatchWriteItemInput
}

func (m *mockDynamo) BatchWriteItem(ctx context.Context, params *dynamodb.BatchWriteItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.BatchWriteItemOutput, error) {
	m.calls++
	m.lastInput = params
	if m.batchWriteFunc != nil {
		return m.batchWriteFunc(ctx, params, optFns...)
	}
	return &dynamodb.BatchWriteItemOutput{}, nil
}

var mockAttributeValueMapper = attributevalue.MarshalMap
