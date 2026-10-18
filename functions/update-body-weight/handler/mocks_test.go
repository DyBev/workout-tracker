package handler

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

// mockDynamo implements Query and Update for testing.
type mockDynamo struct {
	MockQuery func(
		ctx context.Context,
		params *dynamodb.QueryInput,
		optFns ...func(*dynamodb.Options),
	) (*dynamodb.QueryOutput, error)

	MockUpdateItem func(
		ctx context.Context,
		params *dynamodb.UpdateItemInput,
		optFns ...func(*dynamodb.Options),
	) (*dynamodb.UpdateItemOutput, error)

	QueryCalls int
	UpdateCalls int
}

func (m *mockDynamo) Query(
	ctx context.Context,
	params *dynamodb.QueryInput,
	optFns ...func(*dynamodb.Options),
) (*dynamodb.QueryOutput, error) {
	m.QueryCalls++
	if m.MockQuery != nil {
		return m.MockQuery(ctx, params, optFns...)
	}
	return &dynamodb.QueryOutput{}, nil
}

func (m *mockDynamo) UpdateItem(
	ctx context.Context,
	params *dynamodb.UpdateItemInput,
	optFns ...func(*dynamodb.Options),
) (*dynamodb.UpdateItemOutput, error) {
	m.UpdateCalls++
	if m.MockUpdateItem != nil {
		return m.MockUpdateItem(ctx, params, optFns...)
	}
	return &dynamodb.UpdateItemOutput{}, nil
}
