// Package main provides the entry point for the order microservice.
// This service handles order creation, payment simulation, and order management
// for the CloudRetail e-commerce platform. It integrates with ProductService
// (stock checks via GraphQL), fires EventBridge events, and uses RDS PostgreSQL
// with GORM for persistence. JWT validation uses Cognito JWKS.
//
// Suggested folder structure for scaling:
//
//	order_service/
//	├── main.go
//	├── handlers/
//	│   ├── order.go         # Order CRUD handlers
//	│   └── payment.go       # Payment simulation handlers
//	├── middleware/
//	│   └── jwt.go           # JWT auth middleware (Cognito JWKS)
//	├── models/
//	│   └── models.go        # Order model and request/response structs
//	├── clients/
//	│   ├── graphql.go       # ProductService GraphQL client
//	│   └── eventbridge.go   # EventBridge client
//	└── config/
//	    └── config.go        # Environment configuration
package main

import (
	"context"
	"crypto/rsa"
	"database/sql/driver"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge/types"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	graphql "github.com/hasura/go-graphql-client"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// =============================================================================
// Global Variables
// =============================================================================

var (
	db                *gorm.DB
	eventBridgeClient *eventbridge.Client
	graphqlClient     *graphql.Client
	eventBusArn       string
	productGraphQLURL string
	cognitoRegion     string
	userPoolID        string
	jwksCache         map[string]*rsa.PublicKey
	jwksCacheTime     time.Time
	jwksCacheTTL      = 1 * time.Hour
)

// =============================================================================
// Models
// =============================================================================

// OrderItem represents a single item in an order.
type OrderItem struct {
	ProductID string `json:"productId"`
	Quantity  int    `json:"quantity"`
}

// OrderItemsJSON is a JSONB column type for storing order items in PostgreSQL.
type OrderItemsJSON []OrderItem

// Value implements driver.Valuer interface for GORM.
func (o OrderItemsJSON) Value() (driver.Value, error) {
	return json.Marshal(o)
}

// Scan implements sql.Scanner interface for GORM.
func (o *OrderItemsJSON) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to unmarshal JSONB value: %v", value)
	}
	return json.Unmarshal(bytes, o)
}

// OrderModel represents the Orders table in PostgreSQL.
type OrderModel struct {
	OrderID    string         `gorm:"primaryKey;type:uuid;column:order_id" json:"orderId"`
	BuyerID    string         `gorm:"not null;column:buyer_id" json:"buyerId"`
	SellerID   string         `gorm:"not null;column:seller_id" json:"sellerId"`
	Items      OrderItemsJSON `gorm:"type:jsonb;not null;column:items" json:"items"`
	Status     string         `gorm:"not null;default:pending;column:status" json:"status"` // pending, paid, shipped, delivered
	TotalPrice float64        `gorm:"not null;column:total_price" json:"totalPrice"`
	CreatedAt  time.Time      `gorm:"autoCreateTime;column:created_at" json:"createdAt"`
	UpdatedAt  time.Time      `gorm:"autoUpdateTime;column:updated_at" json:"updatedAt"`
}

// TableName specifies the table name for GORM.
func (OrderModel) TableName() string {
	return "orders"
}

// =============================================================================
// Request/Response Structs
// =============================================================================

// CreateOrderInput represents the expected JSON body for creating an order.
type CreateOrderInput struct {
	Items []OrderItem `json:"items" binding:"required,min=1"`
}

// CreateOrderResponse represents the response after creating an order.
type CreateOrderResponse struct {
	OrderID    string `json:"orderId"`
	PaymentURL string `json:"paymentUrl"`
}

// MarkPaymentDoneInput represents the expected JSON body for marking payment as done.
type MarkPaymentDoneInput struct {
	Paid bool `json:"paid" binding:"required"`
}

// MarkPaymentDoneResponse represents the response after marking payment.
type MarkPaymentDoneResponse struct {
	Redirect string `json:"redirect"`
}

// UpdateStatusInput represents the expected JSON body for updating order status.
type UpdateStatusInput struct {
	Status string `json:"status" binding:"required"`
}

// ErrorResponse represents a standard error response.
type ErrorResponse struct {
	Error string `json:"error"`
}

// =============================================================================
// JWT Structures
// =============================================================================

// JWTClaims represents the JWT claims from Cognito.
type JWTClaims struct {
	Email      string `json:"email"`
	Sub        string `json:"sub"`
	CustomRole string `json:"custom:role"`
	jwt.RegisteredClaims
}

// JWK represents a JSON Web Key from Cognito JWKS.
type JWK struct {
	Kid string `json:"kid"`
	Kty string `json:"kty"`
	Alg string `json:"alg"`
	Use string `json:"use"`
	N   string `json:"n"`
	E   string `json:"e"`
}

// JWKS represents the JSON Web Key Set from Cognito.
type JWKS struct {
	Keys []JWK `json:"keys"`
}

// =============================================================================
// GraphQL Query Structures
// =============================================================================

// ProductQuery represents the GraphQL query for getting product details.
type ProductQuery struct {
	GetProductById struct {
		ProductID string  `graphql:"productId"`
		Name      string  `graphql:"name"`
		Price     float64 `graphql:"price"`
		Stock     int     `graphql:"stock"`
		SellerID  string  `graphql:"sellerId"`
	} `graphql:"getProductById(id: $id)"`
}

// =============================================================================
// Initialization
// =============================================================================

func init() {
	_ = godotenv.Load()

	// AWS Configuration
	awsRegion := os.Getenv("AWS_REGION")
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

	// Service Configuration
	eventBusArn = os.Getenv("EVENTBRIDGE_BUS_ARN")
	if eventBusArn == "" {
		log.Fatal("EVENTBRIDGE_BUS_ARN is required")
	}

	productGraphQLURL = os.Getenv("PRODUCT_GRAPHQL_URL")
	if productGraphQLURL == "" {
		productGraphQLURL = "http://product-service:8082/graphql"
	}

	// Initialize AWS clients
	cfg, err := awsconfig.LoadDefaultConfig(context.Background(), awsconfig.WithRegion(awsRegion))
	if err != nil {
		log.Fatalf("Failed to load AWS config: %v", err)
	}

	eventBridgeClient = eventbridge.NewFromConfig(cfg)
	log.Println("✅ EventBridge client initialized")

	// Initialize GraphQL client
	graphqlClient = graphql.NewClient(productGraphQLURL, nil)
	log.Printf("✅ GraphQL client initialized: %s", productGraphQLURL)

	// Initialize JWKS cache
	jwksCache = make(map[string]*rsa.PublicKey)
}

func main() {
	// Initialize database
	dsn := os.Getenv("RDS_DSN")
	if dsn == "" {
		log.Fatal("RDS_DSN is required")
	}

	var err error
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Auto-migrate the schema
	if err := db.AutoMigrate(&OrderModel{}); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	log.Println("✅ Database connected and migrated")

	// Set up Gin router
	r := gin.Default()

	// Health check endpoint
	r.GET("/health", HandleHealth)

	// Public payment simulation endpoints (no JWT required for simplicity)
	r.GET("/simulatePayment/:orderId", HandleSimulatePayment)
	r.POST("/markPaymentDone/:orderId", HandleMarkPaymentDone)
	r.GET("/orderConfirmed/:orderId", HandleOrderConfirmed)

	// Protected endpoints (require JWT)
	protected := r.Group("/")
	protected.Use(JWTMiddleware())
	{
		protected.POST("/createOrder", HandleCreateOrder)
		protected.GET("/getOrders", HandleGetOrders)
		protected.PUT("/updateStatus/:orderId", HandleUpdateStatus)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8083"
	}

	log.Printf("🚀 Order Service running on http://localhost:%s", port)

	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// =============================================================================
// JWT Middleware
// =============================================================================

// JWTMiddleware validates the JWT token from Cognito and extracts claims.
func JWTMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Authorization header required"})
			c.Abort()
			return
		}

		if !strings.HasPrefix(authHeader, "Bearer ") {
			c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Invalid authorization format. Expected 'Bearer <token>'"})
			c.Abort()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		claims, err := ValidateJWT(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Invalid token: " + err.Error()})
			c.Abort()
			return
		}

		// Store claims in context
		c.Set("userId", claims.Sub)
		c.Set("userEmail", claims.Email)
		c.Set("customRole", claims.CustomRole)

		c.Next()
	}
}

// ValidateJWT validates the JWT token using Cognito JWKS.
func ValidateJWT(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		kid, ok := token.Header["kid"].(string)
		if !ok {
			return nil, fmt.Errorf("kid not found in token header")
		}

		return GetRSAPublicKey(kid)
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}

// GetRSAPublicKey fetches the RSA public key from Cognito JWKS (with caching).
func GetRSAPublicKey(kid string) (*rsa.PublicKey, error) {
	// Check cache validity
	if time.Since(jwksCacheTime) > jwksCacheTTL {
		jwksCache = make(map[string]*rsa.PublicKey)
	}

	// Return from cache if available
	if pubKey, ok := jwksCache[kid]; ok {
		return pubKey, nil
	}

	// Fetch JWKS
	jwks, err := FetchJWKS()
	if err != nil {
		return nil, err
	}

	// Cache all keys
	for _, jwk := range jwks.Keys {
		pubKey, err := JWKToPublicKey(jwk)
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

// FetchJWKS fetches the JWKS from Cognito.
func FetchJWKS() (*JWKS, error) {
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

// JWKToPublicKey converts a JWK to an RSA public key.
func JWKToPublicKey(jwk JWK) (*rsa.PublicKey, error) {
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

// =============================================================================
// Handlers
// =============================================================================

// HandleHealth godoc
// @Summary Health check
// @Description Returns the health status of the service
// @Tags health
// @Produce json
// @Success 200 {object} map[string]string
// @Router /health [get]
func HandleHealth(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "healthy"})
}

// HandleCreateOrder godoc
// @Summary Create a new order
// @Description Creates a new order, checks stock via ProductService GraphQL, fires EventBridge event
// @Tags orders
// @Accept json
// @Produce json
// @Param request body CreateOrderInput true "Order items"
// @Success 201 {object} CreateOrderResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /createOrder [post]
func HandleCreateOrder(c *gin.Context) {
	var input CreateOrderInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid request. Items are required."})
		return
	}

	// Get buyer ID from JWT
	buyerID, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "User ID not found in token"})
		return
	}

	// Validate items and check stock via GraphQL
	var totalPrice float64
	var sellerID string
	productDetails := make(map[string]struct {
		Price    float64
		SellerID string
	})

	for _, item := range input.Items {
		// Query ProductService for product details
		var query ProductQuery
		variables := map[string]interface{}{
			"id": graphql.ID(item.ProductID),
		}

		err := graphqlClient.Query(context.Background(), &query, variables)
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: fmt.Sprintf("Product %s not found: %v", item.ProductID, err)})
			return
		}

		product := query.GetProductById

		// Check stock availability
		if product.Stock < item.Quantity {
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Error: fmt.Sprintf("Insufficient stock for product %s. Available: %d, Requested: %d",
					item.ProductID, product.Stock, item.Quantity),
			})
			return
		}

		// Calculate total price
		totalPrice += product.Price * float64(item.Quantity)

		// Store product details
		productDetails[item.ProductID] = struct {
			Price    float64
			SellerID string
		}{
			Price:    product.Price,
			SellerID: product.SellerID,
		}

		// Use first product's seller as order seller (simplified)
		if sellerID == "" {
			sellerID = product.SellerID
		}
	}

	// Create order in database (with transaction)
	orderID := uuid.New().String()

	order := OrderModel{
		OrderID:    orderID,
		BuyerID:    buyerID.(string),
		SellerID:   sellerID,
		Items:      input.Items,
		Status:     "pending",
		TotalPrice: totalPrice,
	}

	err := db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&order).Error; err != nil {
			return fmt.Errorf("failed to create order: %w", err)
		}

		// Fire EventBridge event
		if err := FireOrderPlacedEvent(orderID, input.Items); err != nil {
			return fmt.Errorf("failed to fire EventBridge event: %w", err)
		}

		return nil
	})

	if err != nil {
		log.Printf("Transaction failed: %v", err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to create order: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, CreateOrderResponse{
		OrderID:    orderID,
		PaymentURL: fmt.Sprintf("/simulatePayment/%s", orderID),
	})
}

// HandleSimulatePayment godoc
// @Summary Simulate payment page
// @Description Returns a message for payment simulation (frontend shows checkbox)
// @Tags payment
// @Produce json
// @Param orderId path string true "Order ID"
// @Success 200 {object} map[string]string
// @Failure 404 {object} ErrorResponse
// @Router /simulatePayment/{orderId} [get]
func HandleSimulatePayment(c *gin.Context) {
	orderID := c.Param("orderId")

	// Verify order exists
	var order OrderModel
	if err := db.First(&order, "order_id = ?", orderID).Error; err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "Order not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "Check box to mark paid",
		"orderId":    orderID,
		"totalPrice": order.TotalPrice,
		"status":     order.Status,
	})
}

// HandleMarkPaymentDone godoc
// @Summary Mark payment as done
// @Description Marks the order as paid if checkbox is checked
// @Tags payment
// @Accept json
// @Produce json
// @Param orderId path string true "Order ID"
// @Param request body MarkPaymentDoneInput true "Payment status"
// @Success 200 {object} MarkPaymentDoneResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /markPaymentDone/{orderId} [post]
func HandleMarkPaymentDone(c *gin.Context) {
	orderID := c.Param("orderId")

	var input MarkPaymentDoneInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid request. 'paid' field is required."})
		return
	}

	if !input.Paid {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Payment not confirmed"})
		return
	}

	// Update order status to "paid"
	result := db.Model(&OrderModel{}).Where("order_id = ?", orderID).Update("status", "paid")
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to update order status"})
		return
	}

	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "Order not found"})
		return
	}

	c.JSON(http.StatusOK, MarkPaymentDoneResponse{
		Redirect: fmt.Sprintf("/orderConfirmed/%s", orderID),
	})
}

// HandleOrderConfirmed godoc
// @Summary Get confirmed order details
// @Description Returns the confirmed order details
// @Tags orders
// @Produce json
// @Param orderId path string true "Order ID"
// @Success 200 {object} OrderModel
// @Failure 404 {object} ErrorResponse
// @Router /orderConfirmed/{orderId} [get]
func HandleOrderConfirmed(c *gin.Context) {
	orderID := c.Param("orderId")

	var order OrderModel
	if err := db.First(&order, "order_id = ?", orderID).Error; err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "Order not found"})
		return
	}

	c.JSON(http.StatusOK, order)
}

// HandleGetOrders godoc
// @Summary Get orders
// @Description Returns orders filtered by buyer or seller ID from JWT claims
// @Tags orders
// @Produce json
// @Param sellerId query string false "Seller ID (for seller role)"
// @Success 200 {object} []OrderModel
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /getOrders [get]
func HandleGetOrders(c *gin.Context) {
	userID, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "User ID not found in token"})
		return
	}

	customRole, _ := c.Get("customRole")

	var orders []OrderModel

	// If seller, filter by sellerId
	if customRole == "seller" {
		sellerIDParam := c.Query("sellerId")
		if sellerIDParam != "" {
			// Verify seller owns the orders
			if sellerIDParam != userID.(string) {
				c.JSON(http.StatusForbidden, ErrorResponse{Error: "Cannot view other seller's orders"})
				return
			}
			if err := db.Where("seller_id = ?", sellerIDParam).Find(&orders).Error; err != nil {
				c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to fetch orders"})
				return
			}
		} else {
			// Return all orders for this seller
			if err := db.Where("seller_id = ?", userID.(string)).Find(&orders).Error; err != nil {
				c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to fetch orders"})
				return
			}
		}
	} else {
		// Buyer: filter by buyerId
		if err := db.Where("buyer_id = ?", userID.(string)).Find(&orders).Error; err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to fetch orders"})
			return
		}
	}

	c.JSON(http.StatusOK, orders)
}

// HandleUpdateStatus godoc
// @Summary Update order status
// @Description Updates order status (seller only, ownership verified)
// @Tags orders
// @Accept json
// @Produce json
// @Param orderId path string true "Order ID"
// @Param request body UpdateStatusInput true "New status"
// @Success 200 {object} map[string]string
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /updateStatus/{orderId} [put]
func HandleUpdateStatus(c *gin.Context) {
	orderID := c.Param("orderId")

	var input UpdateStatusInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid request. Status is required."})
		return
	}

	// Validate status
	validStatuses := map[string]bool{"shipped": true, "delivered": true, "cancelled": true}
	if !validStatuses[input.Status] {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid status. Must be: shipped, delivered, or cancelled."})
		return
	}

	// Verify seller role
	customRole, _ := c.Get("customRole")
	if customRole != "seller" {
		c.JSON(http.StatusForbidden, ErrorResponse{Error: "Only sellers can update order status"})
		return
	}

	userID, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "User ID not found in token"})
		return
	}

	// Get order to verify ownership
	var order OrderModel
	if err := db.First(&order, "order_id = ?", orderID).Error; err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "Order not found"})
		return
	}

	// Verify seller owns this order
	if order.SellerID != userID.(string) {
		c.JSON(http.StatusForbidden, ErrorResponse{Error: "You can only update your own orders"})
		return
	}

	// Update status
	if err := db.Model(&order).Update("status", input.Status).Error; err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to update order status"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Order status updated successfully",
		"orderId": orderID,
		"status":  input.Status,
	})
}

// =============================================================================
// EventBridge Integration
// =============================================================================

// FireOrderPlacedEvent fires an "order-placed" event to EventBridge.
func FireOrderPlacedEvent(orderID string, items []OrderItem) error {
	detail := map[string]interface{}{
		"orderId": orderID,
		"items":   items,
	}

	detailBytes, err := json.Marshal(detail)
	if err != nil {
		return fmt.Errorf("failed to marshal event detail: %w", err)
	}

	entry := types.PutEventsRequestEntry{
		Source:       aws.String("order-service"),
		DetailType:   aws.String("order-placed"),
		Detail:       aws.String(string(detailBytes)),
		EventBusName: aws.String(eventBusArn),
	}

	_, err = eventBridgeClient.PutEvents(context.Background(), &eventbridge.PutEventsInput{
		Entries: []types.PutEventsRequestEntry{entry},
	})

	if err != nil {
		return fmt.Errorf("failed to put event: %w", err)
	}

	log.Printf("✅ EventBridge event fired: order-placed for order %s", orderID)
	return nil
}
