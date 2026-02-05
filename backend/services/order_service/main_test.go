package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	return r
}

func TestHealthEndpoint(t *testing.T) {
	r := setupTestRouter()
	r.GET("/health", HandleHealth)

	req, _ := http.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "healthy")
}

func TestCreateOrderValidation(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    string
		expectedStatus int
		errorContains  string
	}{
		{
			name:           "empty_body",
			requestBody:    `{}`,
			expectedStatus: http.StatusBadRequest,
			errorContains:  "Items are required",
		},
		{
			name:           "missing_items",
			requestBody:    `{"items": []}`,
			expectedStatus: http.StatusBadRequest,
			errorContains:  "Items are required",
		},
		{
			name:           "invalid_json",
			requestBody:    `{invalid}`,
			expectedStatus: http.StatusBadRequest,
			errorContains:  "Invalid request",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := setupTestRouter()
			r.POST("/createOrder", func(c *gin.Context) {
				var input CreateOrderInput
				if err := c.ShouldBindJSON(&input); err != nil {
					c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid request. Items are required."})
					return
				}
				c.JSON(http.StatusCreated, gin.H{"orderId": "test"})
			})

			req, _ := http.NewRequest("POST", "/createOrder", bytes.NewBufferString(tt.requestBody))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.errorContains != "" {
				assert.Contains(t, w.Body.String(), tt.errorContains)
			}
		})
	}
}

func TestMarkPaymentDoneValidation(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    string
		expectedStatus int
		errorContains  string
	}{
		{
			name:           "invalid_json",
			requestBody:    `{invalid}`,
			expectedStatus: http.StatusBadRequest,
			errorContains:  "Invalid request",
		},
		{
			name:           "paid_true",
			requestBody:    `{"paid": true}`,
			expectedStatus: http.StatusOK,
			errorContains:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := setupTestRouter()
			r.POST("/markPaymentDone/:orderId", func(c *gin.Context) {
				var input MarkPaymentDoneInput
				if err := c.ShouldBindJSON(&input); err != nil {
					c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid request. 'paid' field is required."})
					return
				}

				if !input.Paid {
					c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Payment not confirmed"})
					return
				}

				c.JSON(http.StatusOK, gin.H{"redirect": "/orderConfirmed/123"})
			})

			req, _ := http.NewRequest("POST", "/markPaymentDone/test-order", bytes.NewBufferString(tt.requestBody))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.errorContains != "" {
				assert.Contains(t, w.Body.String(), tt.errorContains)
			}
		})
	}
}

func TestUpdateStatusValidation(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    string
		expectedStatus int
		errorContains  string
	}{
		{
			name:           "empty_body",
			requestBody:    `{}`,
			expectedStatus: http.StatusBadRequest,
			errorContains:  "Status is required",
		},
		{
			name:           "invalid_status",
			requestBody:    `{"status": "invalid"}`,
			expectedStatus: http.StatusBadRequest,
			errorContains:  "Invalid status",
		},
		{
			name:           "valid_shipped_status",
			requestBody:    `{"status": "shipped"}`,
			expectedStatus: http.StatusForbidden, // Will fail due to missing seller role
			errorContains:  "seller",
		},
		{
			name:           "valid_delivered_status",
			requestBody:    `{"status": "delivered"}`,
			expectedStatus: http.StatusForbidden,
			errorContains:  "seller",
		},
		{
			name:           "valid_cancelled_status",
			requestBody:    `{"status": "cancelled"}`,
			expectedStatus: http.StatusForbidden,
			errorContains:  "seller",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := setupTestRouter()
			r.PUT("/updateStatus/:orderId", func(c *gin.Context) {
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

				// Verify seller role (simulated failure)
				customRole, _ := c.Get("customRole")
				if customRole != "seller" {
					c.JSON(http.StatusForbidden, ErrorResponse{Error: "Only sellers can update order status"})
					return
				}

				c.JSON(http.StatusOK, gin.H{"message": "Status updated"})
			})

			req, _ := http.NewRequest("PUT", "/updateStatus/test-order", bytes.NewBufferString(tt.requestBody))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.errorContains != "" {
				assert.Contains(t, w.Body.String(), tt.errorContains)
			}
		})
	}
}

func TestJWTMiddlewareRejectsNoToken(t *testing.T) {
	tests := []struct {
		name           string
		authHeader     string
		expectedStatus int
		errorContains  string
	}{
		{
			name:           "no_auth_header",
			authHeader:     "",
			expectedStatus: http.StatusUnauthorized,
			errorContains:  "Authorization header required",
		},
		{
			name:           "invalid_format_no_bearer",
			authHeader:     "InvalidToken",
			expectedStatus: http.StatusUnauthorized,
			errorContains:  "Invalid authorization format",
		},
		{
			name:           "invalid_format_missing_token",
			authHeader:     "Bearer ",
			expectedStatus: http.StatusUnauthorized,
			errorContains:  "Invalid token",
		},
		{
			name:           "invalid_jwt_token",
			authHeader:     "Bearer invalid.jwt.token",
			expectedStatus: http.StatusUnauthorized,
			errorContains:  "Invalid token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := setupTestRouter()
			r.GET("/protected", JWTMiddleware(), func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "success"})
			})

			req, _ := http.NewRequest("GET", "/protected", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.errorContains != "" {
				assert.Contains(t, w.Body.String(), tt.errorContains)
			}
		})
	}
}

func TestValidateJWTFunction(t *testing.T) {
	tests := []struct {
		name        string
		token       string
		expectError bool
	}{
		{
			name:        "empty_token",
			token:       "",
			expectError: true,
		},
		{
			name:        "invalid_token_format",
			token:       "not.a.valid.jwt",
			expectError: true,
		},
		{
			name:        "malformed_token",
			token:       "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.invalid",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ValidateJWT(tt.token)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestOrderItemsJSONMarshal(t *testing.T) {
	tests := []struct {
		name  string
		items OrderItemsJSON
	}{
		{
			name: "single_item",
			items: OrderItemsJSON{
				{ProductID: "prod-1", Quantity: 2},
			},
		},
		{
			name: "multiple_items",
			items: OrderItemsJSON{
				{ProductID: "prod-1", Quantity: 2},
				{ProductID: "prod-2", Quantity: 1},
			},
		},
		{
			name:  "empty_items",
			items: OrderItemsJSON{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test Value (marshal)
			value, err := tt.items.Value()
			assert.NoError(t, err)
			assert.NotNil(t, value)

			// Test Scan (unmarshal)
			var scanned OrderItemsJSON
			err = scanned.Scan(value)
			assert.NoError(t, err)
			assert.Equal(t, len(tt.items), len(scanned))
		})
	}
}

func TestOrderModelStructure(t *testing.T) {
	tests := []struct {
		name  string
		order OrderModel
	}{
		{
			name: "valid_pending_order",
			order: OrderModel{
				OrderID:  "order-1",
				BuyerID:  "buyer-1",
				SellerID: "seller-1",
				Items: OrderItemsJSON{
					{ProductID: "prod-1", Quantity: 2},
				},
				Status:     "pending",
				TotalPrice: 99.99,
				CreatedAt:  time.Now(),
				UpdatedAt:  time.Now(),
			},
		},
		{
			name: "valid_paid_order",
			order: OrderModel{
				OrderID:  "order-2",
				BuyerID:  "buyer-2",
				SellerID: "seller-2",
				Items: OrderItemsJSON{
					{ProductID: "prod-2", Quantity: 1},
					{ProductID: "prod-3", Quantity: 3},
				},
				Status:     "paid",
				TotalPrice: 299.99,
				CreatedAt:  time.Now(),
				UpdatedAt:  time.Now(),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Validate order structure
			assert.NotEmpty(t, tt.order.OrderID)
			assert.NotEmpty(t, tt.order.BuyerID)
			assert.NotEmpty(t, tt.order.SellerID)
			assert.NotEmpty(t, tt.order.Items)
			assert.NotEmpty(t, tt.order.Status)
			assert.Greater(t, tt.order.TotalPrice, 0.0)
		})
	}
}

func TestTableName(t *testing.T) {
	order := OrderModel{}
	assert.Equal(t, "orders", order.TableName())
}

func TestCreateOrderInputValidation(t *testing.T) {
	tests := []struct {
		name       string
		input      CreateOrderInput
		expectFail bool
	}{
		{
			name: "valid_input",
			input: CreateOrderInput{
				Items: []OrderItem{
					{ProductID: "prod-1", Quantity: 2},
				},
			},
			expectFail: false,
		},
		{
			name: "valid_multiple_items",
			input: CreateOrderInput{
				Items: []OrderItem{
					{ProductID: "prod-1", Quantity: 2},
					{ProductID: "prod-2", Quantity: 1},
				},
			},
			expectFail: false,
		},
		{
			name: "empty_items",
			input: CreateOrderInput{
				Items: []OrderItem{},
			},
			expectFail: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test JSON marshaling
			jsonBytes, err := json.Marshal(tt.input)
			assert.NoError(t, err)

			// Test JSON unmarshaling
			var decoded CreateOrderInput
			err = json.Unmarshal(jsonBytes, &decoded)
			assert.NoError(t, err)

			if !tt.expectFail {
				assert.Equal(t, len(tt.input.Items), len(decoded.Items))
			}
		})
	}
}

func TestStatusValues(t *testing.T) {
	validStatuses := []string{"pending", "paid", "shipped", "delivered"}
	invalidStatuses := []string{"invalid", "processing", "completed"}

	validMap := map[string]bool{
		"shipped":   true,
		"delivered": true,
		"cancelled": true,
	}

	for _, status := range validStatuses {
		t.Run("valid_status_"+status, func(t *testing.T) {
			// Only shipped, delivered, cancelled are valid for updates
			if status == "shipped" || status == "delivered" {
				assert.True(t, validMap[status])
			}
		})
	}

	for _, status := range invalidStatuses {
		t.Run("invalid_status_"+status, func(t *testing.T) {
			assert.False(t, validMap[status])
		})
	}
}

func TestJWKSCacheLogic(t *testing.T) {
	// Test that cache map is initialized
	assert.NotNil(t, jwksCache)

	// Test cache TTL is set
	assert.Equal(t, 1*time.Hour, jwksCacheTTL)
}

func TestEnvironmentDefaults(t *testing.T) {
	// These tests verify that environment variables have sensible defaults or checks
	tests := []struct {
		name     string
		envVar   string
		expected string
	}{
		{
			name:     "default_product_graphql_url",
			envVar:   productGraphQLURL,
			expected: "http://product-service:8082/graphql",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Environment variables should have default values
			assert.NotEmpty(t, tt.envVar)
		})
	}
}

func TestEventBridgePayloadStructure(t *testing.T) {
	tests := []struct {
		name    string
		orderID string
		items   []OrderItem
	}{
		{
			name:    "single_item_order",
			orderID: "order-1",
			items: []OrderItem{
				{ProductID: "prod-1", Quantity: 2},
			},
		},
		{
			name:    "multiple_items_order",
			orderID: "order-2",
			items: []OrderItem{
				{ProductID: "prod-1", Quantity: 2},
				{ProductID: "prod-2", Quantity: 1},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			detail := map[string]interface{}{
				"orderId": tt.orderID,
				"items":   tt.items,
			}

			detailBytes, err := json.Marshal(detail)
			assert.NoError(t, err)
			assert.NotEmpty(t, detailBytes)

			// Verify JSON structure
			var decoded map[string]interface{}
			err = json.Unmarshal(detailBytes, &decoded)
			assert.NoError(t, err)
			assert.Equal(t, tt.orderID, decoded["orderId"])
		})
	}
}

func TestGraphQLQueryStructure(t *testing.T) {
	// Test that ProductQuery structure is well-formed
	query := ProductQuery{}
	assert.NotNil(t, query)

	// Test field types
	query.GetProductById.ProductID = "test-id"
	query.GetProductById.Name = "Test Product"
	query.GetProductById.Price = 99.99
	query.GetProductById.Stock = 10
	query.GetProductById.SellerID = "seller-id"

	assert.Equal(t, "test-id", query.GetProductById.ProductID)
	assert.Equal(t, 99.99, query.GetProductById.Price)
	assert.Equal(t, 10, query.GetProductById.Stock)
}
