package main

import (
	"context"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"os"
	"strings"
	"time"

	"product_service/graph"
	"product_service/graph/model"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
)

var (
	dynamoClient     *dynamodb.Client
	eventBridgeClient *eventbridge.Client
	productsTable    string
	reviewsTable     string
	eventBusName     string
	awsRegion        string
	cognitoRegion    string
	userPoolID       string
	jwksCache        map[string]*rsa.PublicKey
	jwksCacheTime    time.Time
	jwksCacheTTL     = 1 * time.Hour
)

type JWTClaims struct {
	Email      string `json:"email"`
	Sub        string `json:"sub"`
	CustomRole string `json:"custom:role"`
	jwt.RegisteredClaims
}

type JWK struct {
	Kid string `json:"kid"`
	Kty string `json:"kty"`
	Alg string `json:"alg"`
	Use string `json:"use"`
	N   string `json:"n"`
	E   string `json:"e"`
}

type JWKS struct {
	Keys []JWK `json:"keys"`
}

// DynamoDB Product struct
type DynamoProduct struct {
	ProductID   string   `dynamodbav:"productId"`
	Name        string   `dynamodbav:"name"`
	Price       float64  `dynamodbav:"price"`
	Description string   `dynamodbav:"description"`
	Stock       int      `dynamodbav:"stock"`
	SellerID    string   `dynamodbav:"sellerId"`
	CreatedAt   string   `dynamodbav:"createdAt"`
	UpdatedAt   string   `dynamodbav:"updatedAt"`
}

// DynamoDB Review struct
type DynamoReview struct {
	ReviewID  string `dynamodbav:"reviewId"`
	ProductID string `dynamodbav:"productId"`
	Text      string `dynamodbav:"text"`
	Rating    int    `dynamodbav:"rating"`
	UserID    string `dynamodbav:"userId"`
	CreatedAt string `dynamodbav:"createdAt"`
}

// OrderPlacedEvent from EventBridge
type OrderPlacedEvent struct {
	Detail OrderDetail `json:"detail"`
}

type OrderDetail struct {
	ProductID string `json:"productId"`
	Quantity  int    `json:"quantity"`
}

func init() {
	_ = godotenv.Load()

	awsRegion = os.Getenv("AWS_REGION")
	if awsRegion == "" {
		awsRegion = "us-east-1"
	}

	cognitoRegion = os.Getenv("COGNITO_REGION")
	if cognitoRegion == "" {
		cognitoRegion = "us-east-1"
	}

	userPoolID = os.Getenv("COGNITO_USER_POOL_ID")
	if userPoolID == "" {
		log.Fatal("COGNITO_USER_POOL_ID is required")
	}

	productsTable = os.Getenv("PRODUCTS_TABLE")
	if productsTable == "" {
		productsTable = "Products"
	}

	reviewsTable = os.Getenv("REVIEWS_TABLE")
	if reviewsTable == "" {
		reviewsTable = "Reviews"
	}

	eventBusName = os.Getenv("EVENT_BUS_NAME")
	if eventBusName == "" {
		eventBusName = "default"
	}

	jwksCache = make(map[string]*rsa.PublicKey)
}

func main() {
	ctx := context.Background()

	// Initialize AWS clients
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(awsRegion))
	if err != nil {
		log.Fatalf("Failed to load AWS config: %v", err)
	}

	dynamoClient = dynamodb.NewFromConfig(cfg)
	eventBridgeClient = eventbridge.NewFromConfig(cfg)

	log.Println("✅ DynamoDB and EventBridge clients initialized")

	// Start EventBridge listener in background
	go listenToEventBridge(ctx)

	// Set up GraphQL server with Gin
	r := gin.Default()

	// CORS middleware
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusOK)
			return
		}
		c.Next()
	})

	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "healthy"})
	})

	// GraphQL playground (for development)
	r.GET("/playground", gin.WrapH(playground.Handler("GraphQL Playground", "/graphql")))

	// GraphQL endpoint (with JWT middleware for mutations)
	r.POST("/graphql", GinContextToGraphQL(), gin.WrapH(handler.NewDefaultServer(
		graph.NewExecutableSchema(graph.Config{Resolvers: &Resolver{}}),
	)))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8082"
	}

	log.Printf("🚀 Product Service running on http://localhost:%s/graphql", port)
	log.Printf("🎮 GraphQL Playground: http://localhost:%s/playground", port)

	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// GinContextToGraphQL transfers Gin context to GraphQL context
func GinContextToGraphQL() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract JWT token if present
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
			tokenString := strings.TrimPrefix(authHeader, "Bearer ")

			// Validate JWT
			claims, err := validateJWT(tokenString)
			if err == nil {
				// Add seller ID to Gin context for GraphQL resolvers to use
				c.Set("sellerId", claims.Sub)
				c.Set("userEmail", claims.Email)
				c.Set("customRole", claims.CustomRole)
			}
		}

		// Store Gin context in request context for GraphQL resolvers
		ctx := context.WithValue(c.Request.Context(), "GinContextKey", c)
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}

// Fetch JWKS from Cognito
func fetchJWKS() (*JWKS, error) {
	jwksURL := fmt.Sprintf("https://cognito-idp.%s.amazonaws.com/%s/.well-known/jwks.json", cognitoRegion, userPoolID)

	resp, err := http.Get(jwksURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch JWKS: %w", err)
	}
	defer resp.Body.Close()

	var jwks JWKS
	if err := json.NewDecoder(resp.Body).Decode(&jwks); err != nil {
		return nil, fmt.Errorf("failed to decode JWKS: %w", err)
	}

	return &jwks, nil
}

// Convert JWK to RSA public key
func jwkToPublicKey(jwk JWK) (*rsa.PublicKey, error) {
	nBytes, err := base64.RawURLEncoding.DecodeString(jwk.N)
	if err != nil {
		return nil, err
	}

	eBytes, err := base64.RawURLEncoding.DecodeString(jwk.E)
	if err != nil {
		return nil, err
	}

	n := new(big.Int).SetBytes(nBytes)
	var e int
	for _, b := range eBytes {
		e = e<<8 + int(b)
	}

	return &rsa.PublicKey{N: n, E: e}, nil
}

// Get RSA public key from cache or fetch
func getRSAPublicKey(kid string) (*rsa.PublicKey, error) {
	// Check cache validity
	if time.Since(jwksCacheTime) > jwksCacheTTL {
		jwksCache = make(map[string]*rsa.PublicKey)
	}

	// Return from cache if available
	if pubKey, ok := jwksCache[kid]; ok {
		return pubKey, nil
	}

	// Fetch JWKS
	jwks, err := fetchJWKS()
	if err != nil {
		return nil, err
	}

	// Cache all keys
	for _, jwk := range jwks.Keys {
		pubKey, err := jwkToPublicKey(jwk)
		if err != nil {
			continue
		}
		jwksCache[jwk.Kid] = pubKey
	}

	jwksCacheTime = time.Now()

	// Return requested key
	if pubKey, ok := jwksCache[kid]; ok {
		return pubKey, nil
	}

	return nil, fmt.Errorf("kid not found in JWKS")
}

// Validate JWT token
func validateJWT(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		kid, ok := token.Header["kid"].(string)
		if !ok {
			return nil, fmt.Errorf("kid not found in token header")
		}

		return getRSAPublicKey(kid)
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}

// EventBridge listener to handle order-placed events
func listenToEventBridge(ctx context.Context) {
	log.Println("📡 EventBridge listener started (polling mode)")

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("EventBridge listener stopped")
			return
		case <-ticker.C:
			// In production, use EventBridge Rules + Lambda or SQS
			// For this implementation, we'll use a DynamoDB stream or direct API calls
			// This is a placeholder for the polling logic
			log.Println("🔍 Polling for order-placed events...")
			// TODO: Implement proper event polling from EventBridge/SQS
		}
	}
}

// Handle stock update from order-placed event
func handleOrderPlacedEvent(event OrderPlacedEvent) error {
	productID := event.Detail.ProductID
	quantity := event.Detail.Quantity

	log.Printf("📦 Order placed event: ProductID=%s, Quantity=%d", productID, quantity)

	// Get current product
	result, err := dynamoClient.GetItem(context.Background(), &dynamodb.GetItemInput{
		TableName: aws.String(productsTable),
		Key: map[string]types.AttributeValue{
			"productId": &types.AttributeValueMemberS{Value: productID},
		},
	})

	if err != nil {
		return fmt.Errorf("failed to get product: %w", err)
	}

	if result.Item == nil {
		return fmt.Errorf("product not found: %s", productID)
	}

	var product DynamoProduct
	if err := attributevalue.UnmarshalMap(result.Item, &product); err != nil {
		return fmt.Errorf("failed to unmarshal product: %w", err)
	}

	// Check stock availability
	if product.Stock < quantity {
		return fmt.Errorf("insufficient stock: available=%d, requested=%d", product.Stock, quantity)
	}

	// Update stock
	newStock := product.Stock - quantity
	_, err = dynamoClient.UpdateItem(context.Background(), &dynamodb.UpdateItemInput{
		TableName: aws.String(productsTable),
		Key: map[string]types.AttributeValue{
			"productId": &types.AttributeValueMemberS{Value: productID},
		},
		UpdateExpression: aws.String("SET stock = :newStock, updatedAt = :updatedAt"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":newStock": &types.AttributeValueMemberN{
				Value: fmt.Sprintf("%d", newStock),
			},
			":updatedAt": &types.AttributeValueMemberS{
				Value: time.Now().UTC().Format(time.RFC3339),
			},
		},
	})

	if err != nil {
		return fmt.Errorf("failed to update stock: %w", err)
	}

	log.Printf("✅ Stock updated: ProductID=%s, NewStock=%d", productID, newStock)
	return nil
}

// Resolver implementation
type Resolver struct{}

// Query resolver
func (r *Resolver) Query() graph.QueryResolver {
	return &queryResolver{r}
}

// Mutation resolver
func (r *Resolver) Mutation() graph.MutationResolver {
	return &mutationResolver{r}
}

type queryResolver struct{ *Resolver }
type mutationResolver struct{ *Resolver }

// GetProductByID resolver
func (r *queryResolver) GetProductByID(ctx context.Context, id string) (*model.Product, error) {
	result, err := dynamoClient.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(productsTable),
		Key: map[string]types.AttributeValue{
			"productId": &types.AttributeValueMemberS{Value: id},
		},
	})

	if err != nil {
		return nil, fmt.Errorf("failed to get product: %w", err)
	}

	if result.Item == nil {
		return nil, fmt.Errorf("product not found")
	}

	var product DynamoProduct
	if err := attributevalue.UnmarshalMap(result.Item, &product); err != nil {
		return nil, fmt.Errorf("failed to unmarshal product: %w", err)
	}

	// Fetch reviews for this product
	reviews, err := getReviewsForProduct(ctx, id)
	if err != nil {
		log.Printf("Warning: failed to fetch reviews for product %s: %v", id, err)
		reviews = []*model.Review{} // Return empty reviews on error
	}

	return &model.Product{
		ProductID:   product.ProductID,
		Name:        product.Name,
		Price:       product.Price,
		Description: &product.Description,
		Stock:       product.Stock,
		SellerID:    product.SellerID,
		Reviews:     reviews,
		CreatedAt:   &product.CreatedAt,
		UpdatedAt:   &product.UpdatedAt,
	}, nil
}

// GetAllProducts resolver
func (r *queryResolver) GetAllProducts(ctx context.Context, filter *model.ProductFilter) ([]*model.Product, error) {
	var input *dynamodb.ScanInput

	if filter != nil && filter.SellerID != nil {
		// Filter by seller ID
		input = &dynamodb.ScanInput{
			TableName:        aws.String(productsTable),
			FilterExpression: aws.String("sellerId = :sellerId"),
			ExpressionAttributeValues: map[string]types.AttributeValue{
				":sellerId": &types.AttributeValueMemberS{Value: *filter.SellerID},
			},
		}
	} else {
		// Get all products
		input = &dynamodb.ScanInput{
			TableName: aws.String(productsTable),
		}
	}

	result, err := dynamoClient.Scan(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to scan products: %w", err)
	}

	products := make([]*model.Product, 0, len(result.Items))
	for _, item := range result.Items {
		var product DynamoProduct
		if err := attributevalue.UnmarshalMap(item, &product); err != nil {
			log.Printf("Warning: failed to unmarshal product: %v", err)
			continue
		}

		// Fetch reviews for this product
		reviews, _ := getReviewsForProduct(ctx, product.ProductID)
		if reviews == nil {
			reviews = []*model.Review{}
		}

		products = append(products, &model.Product{
			ProductID:   product.ProductID,
			Name:        product.Name,
			Price:       product.Price,
			Description: &product.Description,
			Stock:       product.Stock,
			SellerID:    product.SellerID,
			Reviews:     reviews,
			CreatedAt:   &product.CreatedAt,
			UpdatedAt:   &product.UpdatedAt,
		})
	}

	return products, nil
}

// Health resolver
func (r *queryResolver) Health(ctx context.Context) (string, error) {
	return "Product Service is healthy", nil
}

// AddProduct resolver (requires JWT)
func (r *mutationResolver) AddProduct(ctx context.Context, input model.AddProductInput) (*model.Product, error) {
	// Get seller ID from Gin context
	ginCtx, ok := ctx.Value("GinContextKey").(*gin.Context)
	if !ok {
		return nil, fmt.Errorf("unauthorized: missing authentication")
	}

	sellerID, exists := ginCtx.Get("sellerId")
	if !exists {
		return nil, fmt.Errorf("unauthorized: missing seller ID in token")
	}

	// Verify seller ID matches
	if sellerID.(string) != input.SellerID {
		return nil, fmt.Errorf("forbidden: seller ID mismatch")
	}

	// Create product
	productID := uuid.New().String()
	now := time.Now().UTC().Format(time.RFC3339)

	description := ""
	if input.Description != nil {
		description = *input.Description
	}

	product := DynamoProduct{
		ProductID:   productID,
		Name:        input.Name,
		Price:       input.Price,
		Description: description,
		Stock:       input.Stock,
		SellerID:    input.SellerID,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	av, err := attributevalue.MarshalMap(product)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal product: %w", err)
	}

	_, err = dynamoClient.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(productsTable),
		Item:      av,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to create product: %w", err)
	}

	return &model.Product{
		ProductID:   product.ProductID,
		Name:        product.Name,
		Price:       product.Price,
		Description: &product.Description,
		Stock:       product.Stock,
		SellerID:    product.SellerID,
		Reviews:     []*model.Review{},
		CreatedAt:   &product.CreatedAt,
		UpdatedAt:   &product.UpdatedAt,
	}, nil
}

// EditProduct resolver (requires JWT and ownership check)
func (r *mutationResolver) EditProduct(ctx context.Context, input model.EditProductInput) (*model.Product, error) {
	// Get seller ID from Gin context
	ginCtx, ok := ctx.Value("GinContextKey").(*gin.Context)
	if !ok {
		return nil, fmt.Errorf("unauthorized: missing authentication")
	}

	sellerID, exists := ginCtx.Get("sellerId")
	if !exists {
		return nil, fmt.Errorf("unauthorized: missing seller ID in token")
	}

	// Get existing product
	result, err := dynamoClient.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(productsTable),
		Key: map[string]types.AttributeValue{
			"productId": &types.AttributeValueMemberS{Value: input.ProductID},
		},
	})

	if err != nil {
		return nil, fmt.Errorf("failed to get product: %w", err)
	}

	if result.Item == nil {
		return nil, fmt.Errorf("product not found")
	}

	var product DynamoProduct
	if err := attributevalue.UnmarshalMap(result.Item, &product); err != nil {
		return nil, fmt.Errorf("failed to unmarshal product: %w", err)
	}

	// Verify ownership
	if product.SellerID != sellerID.(string) {
		return nil, fmt.Errorf("forbidden: you can only edit your own products")
	}

	// Build update expression
	updateExpr := "SET updatedAt = :updatedAt"
	exprAttrValues := map[string]types.AttributeValue{
		":updatedAt": &types.AttributeValueMemberS{
			Value: time.Now().UTC().Format(time.RFC3339),
		},
	}

	if input.Name != nil {
		updateExpr += ", #name = :name"
		exprAttrValues[":name"] = &types.AttributeValueMemberS{Value: *input.Name}
	}

	if input.Price != nil {
		updateExpr += ", price = :price"
		exprAttrValues[":price"] = &types.AttributeValueMemberN{
			Value: fmt.Sprintf("%f", *input.Price),
		}
	}

	if input.Description != nil {
		updateExpr += ", description = :description"
		exprAttrValues[":description"] = &types.AttributeValueMemberS{Value: *input.Description}
	}

	if input.Stock != nil {
		updateExpr += ", stock = :stock"
		exprAttrValues[":stock"] = &types.AttributeValueMemberN{
			Value: fmt.Sprintf("%d", *input.Stock),
		}
	}

	exprAttrNames := map[string]string{
		"#name": "name", // 'name' is a reserved keyword in DynamoDB
	}

	_, err = dynamoClient.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String(productsTable),
		Key: map[string]types.AttributeValue{
			"productId": &types.AttributeValueMemberS{Value: input.ProductID},
		},
		UpdateExpression:          aws.String(updateExpr),
		ExpressionAttributeValues: exprAttrValues,
		ExpressionAttributeNames:  exprAttrNames,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to update product: %w", err)
	}

	// Fetch and return updated product
	result2, err := dynamoClient.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(productsTable),
		Key: map[string]types.AttributeValue{
			"productId": &types.AttributeValueMemberS{Value: input.ProductID},
		},
	})

	if err != nil {
		return nil, fmt.Errorf("failed to get updated product: %w", err)
	}

	var updatedProduct DynamoProduct
	if err := attributevalue.UnmarshalMap(result2.Item, &updatedProduct); err != nil {
		return nil, fmt.Errorf("failed to unmarshal updated product: %w", err)
	}

	reviews, _ := getReviewsForProduct(ctx, input.ProductID)
	if reviews == nil {
		reviews = []*model.Review{}
	}

	return &model.Product{
		ProductID:   updatedProduct.ProductID,
		Name:        updatedProduct.Name,
		Price:       updatedProduct.Price,
		Description: &updatedProduct.Description,
		Stock:       updatedProduct.Stock,
		SellerID:    updatedProduct.SellerID,
		Reviews:     reviews,
		CreatedAt:   &updatedProduct.CreatedAt,
		UpdatedAt:   &updatedProduct.UpdatedAt,
	}, nil
}

// AddReview resolver (requires JWT)
func (r *mutationResolver) AddReview(ctx context.Context, input model.AddReviewInput) (*model.Review, error) {
	// Get user ID from Gin context
	ginCtx, ok := ctx.Value("GinContextKey").(*gin.Context)
	if !ok {
		return nil, fmt.Errorf("unauthorized: missing authentication")
	}

	userID, exists := ginCtx.Get("sellerId") // Can be userId or sellerId depending on token
	if !exists {
		return nil, fmt.Errorf("unauthorized: missing user ID in token")
	}

	// Verify product exists
	productResult, err := dynamoClient.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(productsTable),
		Key: map[string]types.AttributeValue{
			"productId": &types.AttributeValueMemberS{Value: input.ProductID},
		},
	})

	if err != nil {
		return nil, fmt.Errorf("failed to get product: %w", err)
	}

	if productResult.Item == nil {
		return nil, fmt.Errorf("product not found")
	}

	// Create review
	reviewID := uuid.New().String()
	now := time.Now().UTC().Format(time.RFC3339)

	review := DynamoReview{
		ReviewID:  reviewID,
		ProductID: input.ProductID,
		Text:      input.Text,
		Rating:    input.Rating,
		UserID:    userID.(string),
		CreatedAt: now,
	}

	av, err := attributevalue.MarshalMap(review)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal review: %w", err)
	}

	_, err = dynamoClient.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(reviewsTable),
		Item:      av,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to create review: %w", err)
	}

	return &model.Review{
		ReviewID:  review.ReviewID,
		ProductID: review.ProductID,
		Text:      &review.Text,
		Rating:    &review.Rating,
		UserID:    &review.UserID,
		CreatedAt: &review.CreatedAt,
	}, nil
}

// Get reviews for a product
func getReviewsForProduct(ctx context.Context, productID string) ([]*model.Review, error) {
	result, err := dynamoClient.Scan(ctx, &dynamodb.ScanInput{
		TableName:        aws.String(reviewsTable),
		FilterExpression: aws.String("productId = :productId"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":productId": &types.AttributeValueMemberS{Value: productID},
		},
	})

	if err != nil {
		return nil, fmt.Errorf("failed to scan reviews: %w", err)
	}

	reviews := make([]*model.Review, 0, len(result.Items))
	for _, item := range result.Items {
		var review DynamoReview
		if err := attributevalue.UnmarshalMap(item, &review); err != nil {
			log.Printf("Warning: failed to unmarshal review: %v", err)
			continue
		}

		reviews = append(reviews, &model.Review{
			ReviewID:  review.ReviewID,
			ProductID: review.ProductID,
			Text:      &review.Text,
			Rating:    &review.Rating,
			UserID:    &review.UserID,
			CreatedAt: &review.CreatedAt,
		})
	}

	return reviews, nil
}
