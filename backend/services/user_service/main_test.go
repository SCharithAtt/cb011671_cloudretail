package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

// setupTestRouter creates a Gin router configured for testing.
func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)

	authService := &MockAuthService{}
	authHandler := NewAuthHandler(authService)

	r := gin.Default()
	r.GET("/health", HandleHealth)
	r.POST("/register", authHandler.HandleRegister)
	r.POST("/login", authHandler.HandleLogin)

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
// Register Endpoint Tests
// =============================================================================

func TestRegisterEndpoint(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    map[string]interface{}
		expectedStatus int
		expectedError  string
	}{
		{
			name: "successful registration",
			requestBody: map[string]interface{}{
				"email":    "test@example.com",
				"password": "securepassword123",
				"name":     "Test User",
			},
			expectedStatus: http.StatusCreated,
			expectedError:  "",
		},
		{
			name: "successful registration without name",
			requestBody: map[string]interface{}{
				"email":    "test2@example.com",
				"password": "securepassword123",
			},
			expectedStatus: http.StatusCreated,
			expectedError:  "",
		},
		{
			name: "missing email",
			requestBody: map[string]interface{}{
				"password": "securepassword123",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Invalid request body",
		},
		{
			name: "missing password",
			requestBody: map[string]interface{}{
				"email": "test@example.com",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Invalid request body",
		},
		{
			name: "invalid email format",
			requestBody: map[string]interface{}{
				"email":    "invalid-email",
				"password": "securepassword123",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Invalid email format",
		},
		{
			name: "password too short",
			requestBody: map[string]interface{}{
				"email":    "test@example.com",
				"password": "short",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Password must be at least 8 characters",
		},
		{
			name:           "empty request body",
			requestBody:    map[string]interface{}{},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Invalid request body",
		},
	}

	router := setupTestRouter()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.requestBody)
			req, _ := http.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectedError != "" {
				var errResponse ErrorResponse
				if err := json.Unmarshal(w.Body.Bytes(), &errResponse); err != nil {
					t.Fatalf("Failed to unmarshal error response: %v", err)
				}
				if !contains(errResponse.Error, tt.expectedError) {
					t.Errorf("Expected error containing '%s', got '%s'", tt.expectedError, errResponse.Error)
				}
			}

			if tt.expectedStatus == http.StatusCreated {
				var response RegisterResponse
				if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}
				if response.UserID == "" {
					t.Error("Expected non-empty userId")
				}
				if response.Message == "" {
					t.Error("Expected non-empty message")
				}
			}
		})
	}
}

// =============================================================================
// Login Endpoint Tests
// =============================================================================

func TestLoginEndpoint(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    map[string]interface{}
		expectedStatus int
		expectedError  string
	}{
		{
			name: "successful login",
			requestBody: map[string]interface{}{
				"email":    "test@example.com",
				"password": "securepassword123",
			},
			expectedStatus: http.StatusOK,
			expectedError:  "",
		},
		{
			name: "missing email",
			requestBody: map[string]interface{}{
				"password": "securepassword123",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Invalid request body",
		},
		{
			name: "missing password",
			requestBody: map[string]interface{}{
				"email": "test@example.com",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Invalid request body",
		},
		{
			name: "invalid email format",
			requestBody: map[string]interface{}{
				"email":    "not-an-email",
				"password": "securepassword123",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Invalid email format",
		},
		{
			name:           "empty request body",
			requestBody:    map[string]interface{}{},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Invalid request body",
		},
	}

	router := setupTestRouter()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.requestBody)
			req, _ := http.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectedError != "" {
				var errResponse ErrorResponse
				if err := json.Unmarshal(w.Body.Bytes(), &errResponse); err != nil {
					t.Fatalf("Failed to unmarshal error response: %v", err)
				}
				if !contains(errResponse.Error, tt.expectedError) {
					t.Errorf("Expected error containing '%s', got '%s'", tt.expectedError, errResponse.Error)
				}
			}

			if tt.expectedStatus == http.StatusOK {
				var response LoginResponse
				if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}
				if response.Token == "" {
					t.Error("Expected non-empty token")
				}
				if response.User.Email == "" {
					t.Error("Expected non-empty user email")
				}
				if response.User.Role == "" {
					t.Error("Expected non-empty user role")
				}
			}
		})
	}
}

// =============================================================================
// Email Validation Tests
// =============================================================================

func TestIsValidEmail(t *testing.T) {
	tests := []struct {
		email    string
		expected bool
	}{
		{"test@example.com", true},
		{"user.name@domain.co.uk", true},
		{"user+tag@example.org", true},
		{"invalid-email", false},
		{"@example.com", false},
		{"user@", false},
		{"", false},
		{"user@domain", false},
		{"user name@example.com", false},
	}

	for _, tt := range tests {
		t.Run(tt.email, func(t *testing.T) {
			result := isValidEmail(tt.email)
			if result != tt.expected {
				t.Errorf("isValidEmail(%q) = %v, expected %v", tt.email, result, tt.expected)
			}
		})
	}
}

// =============================================================================
// Helpers
// =============================================================================

// contains checks if substr is contained in s (case-insensitive partial match).
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && containsSubstring(s, substr)))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
