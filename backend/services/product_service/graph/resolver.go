package graph

import (
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require
// here.

type Resolver struct {
	DynamoClient      *dynamodb.Client
	EventBridgeClient *eventbridge.Client
	ProductsTable     string
	ReviewsTable      string
}
