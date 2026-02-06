package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

// =============================================================================
// Test Setup
// =============================================================================

func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)

	r := gin.Default()

	// Public routes
	r.GET("/health", HandleHealth)

	// For testing protected routes, we skip JWT middleware and set context manually
	return r
}

// setupProtectedTestRouter creates a router with a mock JWT middleware
// that injects sellerId into the context.
func setupProtectedTestRouter(sellerID string) *gin.Engine {
	gin.SetMode(gin.TestMode)

	r := gin.Default()

	// Mock JWT middleware: injects sellerId without actual token validation
	r.Use(func(c *gin.Context) {
		c.Set("sellerId", sellerID)
		c.Set("sellerEmail", "seller@test.com")
		c.Next()
	})

	r.POST("/addProduct", HandleAddProduct)
	r.PUT("/editProduct/:productId", HandleEditProduct)
	r.GET("/orders", HandleGetOrders)
	r.PUT("/updateOrderStatus/:orderId", HandleUpdateOrderStatus)

	return r
}

// =============================================================================
// Health Endpoint Tests
// =============================================================================

func TestHealthEndpoint(t *testing.T) {
	router := setupTestRouter()

	req, _ := http.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response HealthResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response.Status != "healthy" {
		t.Errorf("Expected status 'healthy', got '%s'", response.Status)
	}
}

// =============================================================================
// Seller Login Tests
// =============================================================================

func TestSellerLoginValidation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	r.POST("/sellerLogin", HandleSellerLogin)

	tests := []struct {
		name           string
		body           map[string]interface{}
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "missing email",
			body:           map[string]interface{}{"password": "password123"},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Email and password are required",
		},
		{
			name:           "missing password",
			body:           map[string]interface{}{"email": "seller@test.com"},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Email and password are required",
		},
		{
			name:           "empty body",
			body:           map[string]interface{}{},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Email and password are required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.body)
			req, _ := http.NewRequest(http.MethodPost, "/sellerLogin", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			var errResp ErrorResponse
			json.Unmarshal(w.Body.Bytes(), &errResp)
			if !containsSubstring(errResp.Error, tt.expectedError) {
				t.Errorf("Expected error containing '%s', got '%s'", tt.expectedError, errResp.Error)
			}
		})
	}
}

// =============================================================================
// Seller Register Tests
// =============================================================================

func TestSellerRegisterValidation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	r.POST("/sellerRegister", HandleSellerRegister)

	tests := []struct {
		name           string
		body           map[string]interface{}
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "missing email",
			body:           map[string]interface{}{"password": "password123", "name": "Test"},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Email, password, and name are required",
		},
		{
			name:           "missing password",
			body:           map[string]interface{}{"email": "s@t.com", "name": "Test"},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Email, password, and name are required",
		},
		{
			name:           "missing name",
			body:           map[string]interface{}{"email": "s@t.com", "password": "pass1234"},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Email, password, and name are required",
		},
		{
			name:           "empty body",
			body:           map[string]interface{}{},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Email, password, and name are required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.body)
			req, _ := http.NewRequest(http.MethodPost, "/sellerRegister", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			var errResp ErrorResponse
			json.Unmarshal(w.Body.Bytes(), &errResp)
			if !containsSubstring(errResp.Error, tt.expectedError) {
				t.Errorf("Expected error containing '%s', got '%s'", tt.expectedError, errResp.Error)
			}
		})
	}
}

// =============================================================================
// Add Product Input Validation Tests
// =============================================================================

func TestAddProductValidation(t *testing.T) {
	router := setupProtectedTestRouter("seller-123")

	tests := []struct {
		name           string
		body           map[string]interface{}
		expectedStatus int
	}{
		{
			name:           "missing name",
			body:           map[string]interface{}{"price": 9.99, "description": "Test", "stock": 10},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "missing price",
			body:           map[string]interface{}{"name": "Widget", "description": "Test", "stock": 10},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "missing description",
			body:           map[string]interface{}{"name": "Widget", "price": 9.99, "stock": 10},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "missing stock",
			body:           map[string]interface{}{"name": "Widget", "price": 9.99, "description": "Test"},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "empty body",
			body:           map[string]interface{}{},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.body)
			req, _ := http.NewRequest(http.MethodPost, "/addProduct", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

// =============================================================================
// Update Order Status Validation Tests
// =============================================================================

func TestUpdateOrderStatusValidation(t *testing.T) {
	router := setupProtectedTestRouter("seller-123")

	tests := []struct {
		name           string
		orderID        string
		body           map[string]interface{}
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "invalid status value",
			orderID:        "order-1",
			body:           map[string]interface{}{"status": "invalid"},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Invalid status",
		},
		{
			name:           "missing status",
			orderID:        "order-1",
			body:           map[string]interface{}{},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Status is required",
		},
		{
			name:           "valid shipped status",
			orderID:        "order-1",
			body:           map[string]interface{}{"status": "shipped"},
			expectedStatus: http.StatusInternalServerError, // Will fail because OrderService is not running
		},
		{
			name:           "valid delivered status",
			orderID:        "order-1",
			body:           map[string]interface{}{"status": "delivered"},
			expectedStatus: http.StatusInternalServerError, // Will fail because OrderService is not running
		},
		{
			name:           "valid cancelled status",
			orderID:        "order-1",
			body:           map[string]interface{}{"status": "cancelled"},
			expectedStatus: http.StatusInternalServerError, // Will fail because OrderService is not running
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.body)
			req, _ := http.NewRequest(http.MethodPut, "/updateOrderStatus/"+tt.orderID, bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d. Body: %s", tt.expectedStatus, w.Code, w.Body.String())
			}

			if tt.expectedError != "" && w.Code == http.StatusBadRequest {
				var errResp ErrorResponse
				json.Unmarshal(w.Body.Bytes(), &errResp)
				if !containsSubstring(errResp.Error, tt.expectedError) {
					t.Errorf("Expected error containing '%s', got '%s'", tt.expectedError, errResp.Error)
				}
			}
		})
	}
}

// =============================================================================
// JWT Middleware Tests
// =============================================================================

func TestJWTMiddlewareRejectsNoToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	// Need config + JWKS cache for middleware to work
	config = Config{
		CognitoRegion:     "us-east-1",
		CognitoUserPoolID: "us-east-1_test",
	}
	initJWKSCache()

	r := gin.Default()
	r.Use(JWTAuthMiddleware())
	r.GET("/protected", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	tests := []struct {
		name           string
		authHeader     string
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "no auth header",
			authHeader:     "",
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "Authorization header is required",
		},
		{
			name:           "invalid format - no Bearer",
			authHeader:     "Token abc123",
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "Invalid authorization format",
		},
		{
			name:           "invalid format - missing token",
			authHeader:     "Bearer",
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "Invalid authorization format",
		},
		{
			name:           "invalid JWT token",
			authHeader:     "Bearer invalid.token.here",
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "Invalid or expired token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest(http.MethodGet, "/protected", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			var errResp ErrorResponse
			json.Unmarshal(w.Body.Bytes(), &errResp)
			if !containsSubstring(errResp.Error, tt.expectedError) {
				t.Errorf("Expected error containing '%s', got '%s'", tt.expectedError, errResp.Error)
			}
		})
	}
}

// =============================================================================
// Helpers
// =============================================================================

func containsSubstring(s, substr string) bool {
	if substr == "" {
		return true
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
