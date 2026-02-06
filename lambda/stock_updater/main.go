package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// OrderItem mirrors the order_service item structure
type OrderItem struct {
	ProductID string `json:"productId"`
	Quantity  int    `json:"quantity"`
	Name      string `json:"name"`
	Price     float64 `json:"price"`
}

// OrderPlacedDetail is the EventBridge detail payload
type OrderPlacedDetail struct {
	OrderID string      `json:"orderId"`
	UserID  string      `json:"userId"`
	Items   []OrderItem `json:"items"`
	Total   float64     `json:"total"`
}

var (
	ddbClient      *dynamodb.Client
	productsTable  string
)

func init() {
	productsTable = os.Getenv("PRODUCTS_TABLE")
	if productsTable == "" {
		productsTable = "Products"
	}

	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		log.Fatalf("unable to load AWS config: %v", err)
	}
	ddbClient = dynamodb.NewFromConfig(cfg)
}

func handler(ctx context.Context, event events.CloudWatchEvent) error {
	log.Printf("Received event: source=%s, detail-type=%s", event.Source, event.DetailType)

	var detail OrderPlacedDetail
	if err := json.Unmarshal(event.Detail, &detail); err != nil {
		return fmt.Errorf("failed to unmarshal detail: %w", err)
	}

	log.Printf("Processing order %s with %d items", detail.OrderID, len(detail.Items))

	for _, item := range detail.Items {
		if err := decrementStock(ctx, item.ProductID, item.Quantity); err != nil {
			log.Printf("ERROR: failed to update stock for product %s: %v", item.ProductID, err)
			// Continue processing other items even if one fails
			continue
		}
		log.Printf("Decremented stock for product %s by %d", item.ProductID, item.Quantity)
	}

	log.Printf("Successfully processed order %s", detail.OrderID)
	return nil
}

func decrementStock(ctx context.Context, productID string, quantity int) error {
	_, err := ddbClient.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String(productsTable),
		Key: map[string]types.AttributeValue{
			"productId": &types.AttributeValueMemberS{Value: productID},
		},
		UpdateExpression: aws.String("SET stock = stock - :qty"),
		ConditionExpression: aws.String("stock >= :qty"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":qty": &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", quantity)},
		},
	})
	if err != nil {
		return fmt.Errorf("DynamoDB UpdateItem failed: %w", err)
	}
	return nil
}

func main() {
	lambda.Start(handler)
}
