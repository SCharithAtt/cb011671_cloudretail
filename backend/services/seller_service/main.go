// Package main provides the entry point for the seller microservice.
// This service handles seller-specific operations: authentication,
// product management (via ProductService GraphQL), and order management
// (via OrderService REST). JWT validation uses Cognito JWKS.
//
// Suggested folder structure for scaling:
//
//	seller_service/
//	├── main.go
//	├── handlers/
//	│   ├── auth.go          # Seller login/register handlers
//	│   ├── product.go       # Product CRUD handlers
//	│   └── order.go         # Order viewing/update handlers
//	├── middleware/
//	│   └── jwt.go           # JWT auth middleware (Cognito JWKS)
//	├── models/
//	│   └── models.go        # Request/Response structs
//	├── clients/
//	│   ├── graphql.go       # ProductService GraphQL client
//	│   └── rest.go          # OrderService REST client
//	└── config/
//	    └── config.go        # Environment configuration
package main

import (
	"context"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/big"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider/types"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	graphql "github.com/hasura/go-graphql-client"
	"github.com/joho/godotenv"
)

// =============================================================================
// Models (Request/Response structs)
// =============================================================================

// SellerLoginInput represents the expected JSON body for seller login.
type SellerLoginInput struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// SellerLoginResponse represents the response after successful seller login.
type SellerLoginResponse struct {
	IDToken      string `json:"id_token"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// SellerRegisterInput represents the expected JSON body for seller registration.
type SellerRegisterInput struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
	Name     string `json:"name" binding:"required"`
}

// SellerRegisterResponse represents the response after successful registration.
type SellerRegisterResponse struct {
	Message string `json:"message"`
	UserSub string `json:"userSub"`
}

// AddProductInput represents the expected JSON body for adding a product.
type AddProductInput struct {
	Name        string  `json:"name" binding:"required"`
	Price       float64 `json:"price" binding:"required"`
	Description string  `json:"description" binding:"required"`
	Stock       int     `json:"stock" binding:"required"`
}

// AddProductResponse represents the response after adding a product.
type AddProductResponse struct {
	ProductID string `json:"productId"`
}

// EditProductInput represents the expected JSON body for editing a product.
type EditProductInput struct {
	Name        *string  `json:"name,omitempty"`
	Price       *float64 `json:"price,omitempty"`
	Description *string  `json:"description,omitempty"`
	Stock       *int     `json:"stock,omitempty"`
}

// OrderItem represents an item in an order.
type OrderItem struct {
	ProductID string  `json:"productId"`
	Name      string  `json:"name"`
	Quantity  int     `json:"quantity"`
	Price     float64 `json:"price"`
}

// Order represents an order from OrderService.
type Order struct {
	OrderID string      `json:"orderId"`
	Status  string      `json:"status"`
	Items   []OrderItem `json:"items"`
}

// UpdateOrderStatusInput represents the expected JSON body for updating order status.
type UpdateOrderStatusInput struct {
	Status string `json:"status" binding:"required"`
}

// HealthResponse represents the health check response.
type HealthResponse struct {
	Status string `json:"status"`
}

// ErrorResponse represents a standardized error response.
type ErrorResponse struct {
	Error string `json:"error"`
}

// =============================================================================
// Configuration
// =============================================================================

// Config holds all service configuration loaded from environment variables.
type Config struct {
	CognitoUserPoolID string
	CognitoClientID   string
	CognitoRegion     string
	ProductGraphQLURL string
	OrderRESTURL      string
	Port              string
}

var config Config

func loadConfig() {
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using system environment variables")
	}

	config = Config{
		CognitoUserPoolID: os.Getenv("COGNITO_USER_POOL_ID"),
		CognitoClientID:   os.Getenv("COGNITO_CLIENT_ID"),
		CognitoRegion:     os.Getenv("COGNITO_REGION"),
		ProductGraphQLURL: os.Getenv("PRODUCT_GRAPHQL_URL"),
		OrderRESTURL:      os.Getenv("ORDER_REST_URL"),
		Port:              os.Getenv("PORT"),
	}

	// Defaults
	if config.ProductGraphQLURL == "" {
		config.ProductGraphQLURL = "http://product-service:8082/graphql"
	}
	if config.OrderRESTURL == "" {
		config.OrderRESTURL = "http://order-service:8083"
	}
	if config.Port == "" {
		config.Port = "8081"
	}

	if config.CognitoUserPoolID == "" || config.CognitoClientID == "" || config.CognitoRegion == "" {
		log.Fatal("Missing required env vars: COGNITO_USER_POOL_ID, COGNITO_CLIENT_ID, COGNITO_REGION")
	}
}

// =============================================================================
// JWKS (JSON Web Key Set) for Cognito JWT Validation
// =============================================================================

// JWK represents a single JSON Web Key.
type JWK struct {
	Kty string `json:"kty"`
	Kid string `json:"kid"`
	Use string `json:"use"`
	N   string `json:"n"`
	E   string `json:"e"`
	Alg string `json:"alg"`
}

// JWKS represents a JSON Web Key Set.
type JWKS struct {
	Keys []JWK `json:"keys"`
}

// JWKSCache caches the public keys from Cognito.
type JWKSCache struct {
	mu      sync.RWMutex
	keys    map[string]*rsa.PublicKey
	fetched time.Time
	ttl     time.Duration
	url     string
}

var jwksCache *JWKSCache

func initJWKSCache() {
	jwksURL := fmt.Sprintf(
		"https://cognito-idp.%s.amazonaws.com/%s/.well-known/jwks.json",
		config.CognitoRegion, config.CognitoUserPoolID,
	)
	jwksCache = &JWKSCache{
		keys: make(map[string]*rsa.PublicKey),
		ttl:  1 * time.Hour,
		url:  jwksURL,
	}
}

// GetKey returns the RSA public key for the given key ID.
func (c *JWKSCache) GetKey(kid string) (*rsa.PublicKey, error) {
	c.mu.RLock()
	if key, ok := c.keys[kid]; ok && time.Since(c.fetched) < c.ttl {
		c.mu.RUnlock()
		return key, nil
	}
	c.mu.RUnlock()

	// Fetch and refresh keys
	return c.refresh(kid)
}

func (c *JWKSCache) refresh(kid string) (*rsa.PublicKey, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Double-check after acquiring write lock
	if key, ok := c.keys[kid]; ok && time.Since(c.fetched) < c.ttl {
		return key, nil
	}

	resp, err := http.Get(c.url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch JWKS: %w", err)
	}
	defer resp.Body.Close()

	var jwks JWKS
	if err := json.NewDecoder(resp.Body).Decode(&jwks); err != nil {
		return nil, fmt.Errorf("failed to decode JWKS: %w", err)
	}

	c.keys = make(map[string]*rsa.PublicKey)
	for _, key := range jwks.Keys {
		if key.Kty != "RSA" || key.Use != "sig" {
			continue
		}
		pubKey, err := parseRSAPublicKey(key)
		if err != nil {
			log.Printf("Warning: failed to parse key %s: %v", key.Kid, err)
			continue
		}
		c.keys[key.Kid] = pubKey
	}
	c.fetched = time.Now()

	if key, ok := c.keys[kid]; ok {
		return key, nil
	}
	return nil, fmt.Errorf("key %s not found in JWKS", kid)
}

func parseRSAPublicKey(key JWK) (*rsa.PublicKey, error) {
	nBytes, err := base64.RawURLEncoding.DecodeString(key.N)
	if err != nil {
		return nil, fmt.Errorf("failed to decode modulus: %w", err)
	}
	eBytes, err := base64.RawURLEncoding.DecodeString(key.E)
	if err != nil {
		return nil, fmt.Errorf("failed to decode exponent: %w", err)
	}

	n := new(big.Int).SetBytes(nBytes)
	e := new(big.Int).SetBytes(eBytes)

	return &rsa.PublicKey{
		N: n,
		E: int(e.Int64()),
	}, nil
}

// =============================================================================
// Cognito Client
// =============================================================================

var cognitoClient *cognitoidentityprovider.Client

func initCognitoClient() {
	cfg, err := awsconfig.LoadDefaultConfig(context.Background(),
		awsconfig.WithRegion(config.CognitoRegion),
	)
	if err != nil {
		log.Fatalf("Failed to load AWS config: %v", err)
	}
	cognitoClient = cognitoidentityprovider.NewFromConfig(cfg)
}

// =============================================================================
// GraphQL Client (ProductService)
// =============================================================================

var productGQLClient *graphql.Client

func initGraphQLClient() {
	productGQLClient = graphql.NewClient(config.ProductGraphQLURL, nil)
}

// =============================================================================
// Middleware
// =============================================================================

// JWTAuthMiddleware validates JWT tokens from Cognito and checks for seller role.
// Extracts claims and sets sellerId in the Gin context.
func JWTAuthMiddleware() gin.HandlerFunc {
	issuerURL := fmt.Sprintf(
		"https://cognito-idp.%s.amazonaws.com/%s",
		config.CognitoRegion, config.CognitoUserPoolID,
	)

	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Authorization header is required."})
			c.Abort()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Invalid authorization format. Use: Bearer <token>"})
			c.Abort()
			return
		}

		tokenString := parts[1]

		// Parse and validate JWT
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Verify signing method
			if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}

			kid, ok := token.Header["kid"].(string)
			if !ok {
				return nil, fmt.Errorf("missing kid in token header")
			}

			return jwksCache.GetKey(kid)
		},
			jwt.WithIssuer(issuerURL),
			jwt.WithExpirationRequired(),
		)

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Invalid or expired token."})
			c.Abort()
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Invalid claims."})
			c.Abort()
			return
		}

		// Check custom:role == "seller"
		role, _ := claims["custom:role"].(string)
		if role != "seller" {
			c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Access denied. Seller role required."})
			c.Abort()
			return
		}

		// Extract seller ID (Cognito sub)
		sellerID, _ := claims["sub"].(string)
		if sellerID == "" {
			c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Invalid token: missing sub claim."})
			c.Abort()
			return
		}

		// Set seller info in context for downstream handlers
		c.Set("sellerId", sellerID)
		c.Set("sellerEmail", claims["email"])
		c.Set("claims", claims)

		c.Next()
	}
}

// =============================================================================
// Handlers
// =============================================================================

// HandleSellerLogin godoc
// @Summary Authenticate a seller
// @Description Authenticates seller via Cognito USER_PASSWORD_AUTH flow, validates role == "seller"
// @Tags auth
// @Accept json
// @Produce json
// @Param request body SellerLoginInput true "Login credentials"
// @Success 200 {object} SellerLoginResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /sellerLogin [post]
func HandleSellerLogin(c *gin.Context) {
	var input SellerLoginInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid request. Email and password are required."})
		return
	}

	// Call Cognito InitiateAuth with USER_PASSWORD_AUTH
	authOutput, err := cognitoClient.InitiateAuth(context.Background(),
		&cognitoidentityprovider.InitiateAuthInput{
			AuthFlow: types.AuthFlowTypeUserPasswordAuth,
			ClientId: aws.String(config.CognitoClientID),
			AuthParameters: map[string]string{
				"USERNAME": input.Email,
				"PASSWORD": input.Password,
			},
		},
	)
	if err != nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Authentication failed: " + err.Error()})
		return
	}

	if authOutput.AuthenticationResult == nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Authentication failed: no result returned."})
		return
	}

	result := authOutput.AuthenticationResult

	// Parse id_token to verify role == "seller"
	idToken := aws.ToString(result.IdToken)
	token, _, err := new(jwt.Parser).ParseUnverified(idToken, jwt.MapClaims{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to parse ID token."})
		return
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Invalid token claims."})
		return
	}

	role, _ := claims["custom:role"].(string)
	if role != "seller" {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Access denied. This account is not a seller."})
		return
	}

	c.JSON(http.StatusOK, SellerLoginResponse{
		IDToken:      idToken,
		AccessToken:  aws.ToString(result.AccessToken),
		RefreshToken: aws.ToString(result.RefreshToken),
	})
}

// HandleSellerRegister godoc
// @Summary Register a new seller
// @Description Registers a seller in Cognito, auto-confirms, sets custom:role="seller"
// @Tags auth
// @Accept json
// @Produce json
// @Param request body SellerRegisterInput true "Registration details"
// @Success 201 {object} SellerRegisterResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /sellerRegister [post]
func HandleSellerRegister(c *gin.Context) {
	var input SellerRegisterInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid request. Email, password, and name are required."})
		return
	}

	ctx := context.Background()

	// Sign up user in Cognito
	signUpOutput, err := cognitoClient.SignUp(ctx,
		&cognitoidentityprovider.SignUpInput{
			ClientId: aws.String(config.CognitoClientID),
			Username: aws.String(input.Email),
			Password: aws.String(input.Password),
			UserAttributes: []types.AttributeType{
				{Name: aws.String("email"), Value: aws.String(input.Email)},
				{Name: aws.String("name"), Value: aws.String(input.Name)},
				{Name: aws.String("custom:role"), Value: aws.String("seller")},
			},
		},
	)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Registration failed: " + err.Error()})
		return
	}

	// Auto-confirm for development
	_, err = cognitoClient.AdminConfirmSignUp(ctx,
		&cognitoidentityprovider.AdminConfirmSignUpInput{
			UserPoolId: aws.String(config.CognitoUserPoolID),
			Username:   aws.String(input.Email),
		},
	)
	if err != nil {
		log.Printf("Warning: Auto-confirm failed (may need manual confirmation): %v", err)
	}

	c.JSON(http.StatusCreated, SellerRegisterResponse{
		Message: "Seller registered successfully",
		UserSub: aws.ToString(signUpOutput.UserSub),
	})
}

// HandleAddProduct godoc
// @Summary Add a new product
// @Description Creates a new product via ProductService GraphQL mutation
// @Tags products
// @Accept json
// @Produce json
// @Param request body AddProductInput true "Product details"
// @Success 201 {object} AddProductResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /addProduct [post]
func HandleAddProduct(c *gin.Context) {
	var input AddProductInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid request. Name, price, description, and stock are required."})
		return
	}

	sellerID := c.GetString("sellerId")

	// GraphQL mutation to add product
	var mutation struct {
		AddProduct struct {
			ProductID graphql.String `graphql:"productId"`
		} `graphql:"addProduct(input: $input)"`
	}

	variables := map[string]interface{}{
		"input": map[string]interface{}{
			"name":        input.Name,
			"price":       input.Price,
			"description": input.Description,
			"stock":       input.Stock,
			"sellerId":    sellerID,
		},
	}

	err := productGQLClient.Mutate(context.Background(), &mutation, variables)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to add product: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, AddProductResponse{
		ProductID: string(mutation.AddProduct.ProductID),
	})
}

// HandleEditProduct godoc
// @Summary Edit an existing product
// @Description Updates product fields via ProductService GraphQL, verifies seller ownership
// @Tags products
// @Accept json
// @Produce json
// @Param productId path string true "Product ID"
// @Param request body EditProductInput true "Fields to update"
// @Success 200 {object} map[string]string
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /editProduct/{productId} [put]
func HandleEditProduct(c *gin.Context) {
	productID := c.Param("productId")
	if productID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Product ID is required."})
		return
	}

	var input EditProductInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid request body."})
		return
	}

	sellerID := c.GetString("sellerId")

	// Step 1: Verify ownership via GraphQL query
	var query struct {
		GetProductByID struct {
			SellerID graphql.String `graphql:"sellerId"`
		} `graphql:"getProductById(id: $id)"`
	}

	queryVars := map[string]interface{}{
		"id": graphql.String(productID),
	}

	err := productGQLClient.Query(context.Background(), &query, queryVars)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to verify product ownership: " + err.Error()})
		return
	}

	if string(query.GetProductByID.SellerID) != sellerID {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Access denied. You do not own this product."})
		return
	}

	// Step 2: Mutate product
	var mutation struct {
		EditProduct struct {
			ProductID graphql.String `graphql:"productId"`
		} `graphql:"editProduct(input: $input)"`
	}

	updates := map[string]interface{}{}
	if input.Name != nil {
		updates["name"] = *input.Name
	}
	if input.Price != nil {
		updates["price"] = *input.Price
	}
	if input.Description != nil {
		updates["description"] = *input.Description
	}
	if input.Stock != nil {
		updates["stock"] = *input.Stock
	}

	mutateVars := map[string]interface{}{
		"input": map[string]interface{}{
			"productId": productID,
			"updates":   updates,
		},
	}

	err = productGQLClient.Mutate(context.Background(), &mutation, mutateVars)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to update product: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Product updated successfully", "productId": productID})
}

// HandleGetOrders godoc
// @Summary Get seller's orders
// @Description Fetches orders from OrderService REST API for the authenticated seller
// @Tags orders
// @Produce json
// @Success 200 {array} Order
// @Failure 500 {object} ErrorResponse
// @Router /orders [get]
func HandleGetOrders(c *gin.Context) {
	sellerID := c.GetString("sellerId")

	// Call OrderService REST API
	url := fmt.Sprintf("%s/getOrders?sellerId=%s", config.OrderRESTURL, sellerID)
	resp, err := http.Get(url)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to fetch orders: " + err.Error()})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		c.JSON(resp.StatusCode, ErrorResponse{Error: "OrderService error: " + string(body)})
		return
	}

	var orders []Order
	if err := json.NewDecoder(resp.Body).Decode(&orders); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to decode orders response."})
		return
	}

	c.JSON(http.StatusOK, orders)
}

// HandleUpdateOrderStatus godoc
// @Summary Update order status
// @Description Updates order status via OrderService REST API, verifies seller ownership
// @Tags orders
// @Accept json
// @Produce json
// @Param orderId path string true "Order ID"
// @Param request body UpdateOrderStatusInput true "New status"
// @Success 200 {object} map[string]string
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /updateOrderStatus/{orderId} [put]
func HandleUpdateOrderStatus(c *gin.Context) {
	orderID := c.Param("orderId")
	if orderID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Order ID is required."})
		return
	}

	var input UpdateOrderStatusInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid request. Status is required."})
		return
	}

	// Validate status values
	validStatuses := map[string]bool{"shipped": true, "delivered": true, "cancelled": true}
	if !validStatuses[input.Status] {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid status. Must be: shipped, delivered, or cancelled."})
		return
	}

	sellerID := c.GetString("sellerId")

	// Call OrderService REST API PUT /updateStatus/{orderId}
	url := fmt.Sprintf("%s/updateStatus/%s", config.OrderRESTURL, orderID)
	payload, _ := json.Marshal(map[string]string{
		"status":   input.Status,
		"sellerId": sellerID,
	})

	req, err := http.NewRequest(http.MethodPut, url, strings.NewReader(string(payload)))
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to create request."})
		return
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to update order status: " + err.Error()})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		c.JSON(resp.StatusCode, ErrorResponse{Error: "OrderService error: " + string(body)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Order status updated", "orderId": orderID, "status": input.Status})
}

// HandleHealth godoc
// @Summary Health check endpoint
// @Description Returns the health status of the seller service
// @Tags system
// @Produce json
// @Success 200 {object} HealthResponse
// @Router /health [get]
func HandleHealth(c *gin.Context) {
	c.JSON(http.StatusOK, HealthResponse{Status: "healthy"})
}

// =============================================================================
// Main
// =============================================================================

func main() {
	// Load configuration from environment
	loadConfig()

	// Initialize services
	initJWKSCache()
	initCognitoClient()
	initGraphQLClient()

	// Create Gin router with default middleware (logger, recovery)
	r := gin.Default()

	// Public routes (no authentication required)
	r.GET("/health", HandleHealth)
	r.POST("/sellerLogin", HandleSellerLogin)
	r.POST("/sellerRegister", HandleSellerRegister)

	// Protected routes (require JWT with seller role)
	protected := r.Group("/")
	protected.Use(JWTAuthMiddleware())
	{
		// Product management
		protected.POST("/addProduct", HandleAddProduct)
		protected.PUT("/editProduct/:productId", HandleEditProduct)

		// Order management
		protected.GET("/orders", HandleGetOrders)
		protected.PUT("/updateOrderStatus/:orderId", HandleUpdateOrderStatus)
	}

	// Start server
	log.Printf("SellerService starting on port %s", config.Port)
	r.Run(":" + config.Port)
}
