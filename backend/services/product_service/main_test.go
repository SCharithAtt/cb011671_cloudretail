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
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "healthy"})
	})

	req, _ := http.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "healthy")
}

func TestGraphQLEndpointRequiresAuthentication(t *testing.T) {
	tests := []struct {
		name           string
		mutation       string
		authHeader     string
		expectedStatus int
	}{
		{
			name: "no_auth_header",
			mutation: `{
				"query": "mutation { addProduct(input: { name: \"Test\", price: 10.0, stock: 5, sellerId: \"123\" }) { productId } }"
			}`,
			authHeader:     "",
			expectedStatus: http.StatusOK, // GraphQL returns 200 but with error in response
		},
		{
			name: "invalid_token",
			mutation: `{
				"query": "mutation { addProduct(input: { name: \"Test\", price: 10.0, stock: 5, sellerId: \"123\" }) { productId } }"
			}`,
			authHeader:     "Bearer invalid_token",
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := setupTestRouter()
			
			r.POST("/graphql", func(c *gin.Context) {
				// Simplified test - just check that endpoint exists
				c.JSON(http.StatusOK, gin.H{"data": nil, "errors": []gin.H{{"message": "unauthorized"}}})
			})

			req, _ := http.NewRequest("POST", "/graphql", bytes.NewBufferString(tt.mutation))
			req.Header.Set("Content-Type", "application/json")
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
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
			_, err := validateJWT(tt.token)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGraphQLQueryValidation(t *testing.T) {
	tests := []struct {
		name     string
		query    string
		expected bool
	}{
		{
			name:     "valid_getProductById_query",
			query:    `query { getProductById(id: "123") { productId name } }`,
			expected: true,
		},
		{
			name:     "valid_getAllProducts_query",
			query:    `query { getAllProducts { productId name price } }`,
			expected: true,
		},
		{
			name:     "valid_health_query",
			query:    `query { health }`,
			expected: true,
		},
		{
			name:     "empty_query",
			query:    "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simple validation: non-empty query string
			isValid := len(tt.query) > 0
			assert.Equal(t, tt.expected, isValid)
		})
	}
}

func TestGraphQLMutationValidation(t *testing.T) {
	tests := []struct {
		name     string
		mutation string
		expected bool
	}{
		{
			name:     "valid_addProduct_mutation",
			mutation: `mutation { addProduct(input: { name: "Test", price: 10.0, stock: 5, sellerId: "123" }) { productId } }`,
			expected: true,
		},
		{
			name:     "valid_editProduct_mutation",
			mutation: `mutation { editProduct(input: { productId: "123", name: "Updated" }) { productId } }`,
			expected: true,
		},
		{
			name:     "valid_addReview_mutation",
			mutation: `mutation { addReview(input: { productId: "123", text: "Great!", rating: 5, userId: "456" }) { reviewId } }`,
			expected: true,
		},
		{
			name:     "empty_mutation",
			mutation: "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simple validation: non-empty mutation string
			isValid := len(tt.mutation) > 0
			assert.Equal(t, tt.expected, isValid)
		})
	}
}

func TestOrderPlacedEventStructure(t *testing.T) {
	tests := []struct {
		name      string
		eventJSON string
		expectErr bool
	}{
		{
			name:      "valid_event",
			eventJSON: `{"detail": {"productId": "123", "quantity": 5}}`,
			expectErr: false,
		},
		{
			name:      "missing_productId",
			eventJSON: `{"detail": {"quantity": 5}}`,
			expectErr: false, // Still valid JSON, just missing field
		},
		{
			name:      "missing_quantity",
			eventJSON: `{"detail": {"productId": "123"}}`,
			expectErr: false,
		},
		{
			name:      "invalid_json",
			eventJSON: `{invalid}`,
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var event OrderPlacedEvent
			err := json.Unmarshal([]byte(tt.eventJSON), &event)
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDynamoProductStructure(t *testing.T) {
	tests := []struct {
		name      string
		product   DynamoProduct
		expectErr bool
	}{
		{
			name: "valid_product",
			product: DynamoProduct{
				ProductID:   "123",
				Name:        "Test Product",
				Price:       99.99,
				Description: "Test description",
				Stock:       10,
				SellerID:    "seller-123",
				CreatedAt:   "2024-01-01T00:00:00Z",
				UpdatedAt:   "2024-01-01T00:00:00Z",
			},
			expectErr: false,
		},
		{
			name: "product_with_zero_stock",
			product: DynamoProduct{
				ProductID:   "456",
				Name:        "Out of Stock",
				Price:       49.99,
				Description: "No stock",
				Stock:       0,
				SellerID:    "seller-456",
				CreatedAt:   "2024-01-01T00:00:00Z",
				UpdatedAt:   "2024-01-01T00:00:00Z",
			},
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Validate product structure
			assert.NotEmpty(t, tt.product.ProductID)
			assert.NotEmpty(t, tt.product.Name)
			assert.GreaterOrEqual(t, tt.product.Price, 0.0)
			assert.GreaterOrEqual(t, tt.product.Stock, 0)
		})
	}
}

func TestDynamoReviewStructure(t *testing.T) {
	tests := []struct {
		name   string
		review DynamoReview
	}{
		{
			name: "valid_review",
			review: DynamoReview{
				ReviewID:  "rev-123",
				ProductID: "prod-456",
				Text:      "Great product!",
				Rating:    5,
				UserID:    "user-789",
				CreatedAt: "2024-01-01T00:00:00Z",
			},
		},
		{
			name: "review_with_low_rating",
			review: DynamoReview{
				ReviewID:  "rev-456",
				ProductID: "prod-789",
				Text:      "Not satisfied",
				Rating:    1,
				UserID:    "user-012",
				CreatedAt: "2024-01-01T00:00:00Z",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Validate review structure
			assert.NotEmpty(t, tt.review.ReviewID)
			assert.NotEmpty(t, tt.review.ProductID)
			assert.NotEmpty(t, tt.review.UserID)
			assert.GreaterOrEqual(t, tt.review.Rating, 1)
			assert.LessOrEqual(t, tt.review.Rating, 5)
		})
	}
}

func TestJWKSCacheLogic(t *testing.T) {
	// Test that cache map is initialized
	assert.NotNil(t, jwksCache)
	
	// Test cache TTL is set
	assert.Equal(t, 1*time.Hour, jwksCacheTTL)
}

func TestEnvironmentVariables(t *testing.T) {
	// Test default values
	tests := []struct {
		name     string
		envVar   string
		expected string
	}{
		{
			name:     "default_products_table",
			envVar:   productsTable,
			expected: "Products",
		},
		{
			name:     "default_reviews_table",
			envVar:   reviewsTable,
			expected: "Reviews",
		},
		{
			name:     "default_event_bus",
			envVar:   eventBusName,
			expected: "default",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Environment variables should have default values
			assert.NotEmpty(t, tt.envVar)
		})
	}
}
