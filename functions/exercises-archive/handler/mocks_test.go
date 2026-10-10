package handler

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

// mockDynamo implements DynamoUpdater for testing.
type mockDynamo struct {
	updateFunc func(ctx context.Context, params *dynamodb.UpdateItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.UpdateItemOutput, error)
	calls      int
	lastInput  *dynamodb.UpdateItemInput
}

func (m *mockDynamo) UpdateItem(ctx context.Context, params *dynamodb.UpdateItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.UpdateItemOutput, error) {
	m.calls++
	m.lastInput = params
	if m.updateFunc != nil {
		return m.updateFunc(ctx, params, optFns...)
	}
	return &dynamodb.UpdateItemOutput{}, nil
}

