// Package main provides the entry point for the user microservice.
// This service handles user authentication using AWS Cognito with OIDC/OAuth2.
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/joho/godotenv"
	"golang.org/x/oauth2"
)

// =============================================================================
// Models
// =============================================================================

// frontendURL is the base URL of the frontend application (for redirects).
var frontendURL string

// =============================================================================
// Global Variables (initialized from environment)
// =============================================================================

var (
	clientID     string
	clientSecret string
	redirectURL  string
	issuerURL    string
	logoutURL    string
	provider     *oidc.Provider
	oauth2Config oauth2.Config
)

// =============================================================================
// Initialization
// =============================================================================

func init() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using system environment variables")
	}

	// Load configuration from environment variables
	clientID = os.Getenv("COGNITO_CLIENT_ID")
	clientSecret = os.Getenv("COGNITO_CLIENT_SECRET") // Optional for public clients
	region := os.Getenv("COGNITO_REGION")
	userPoolID := os.Getenv("COGNITO_USER_POOL_ID")
	redirectURL = os.Getenv("COGNITO_REDIRECT_URL")
	logoutURL = os.Getenv("COGNITO_LOGOUT_URL")

	// Set defaults if not provided
	if redirectURL == "" {
		redirectURL = "http://localhost:8080/callback"
	}
	if logoutURL == "" {
		logoutURL = "http://localhost:8080"
	}

	// Frontend URL defaults to logoutURL (Amplify app URL)
	frontendURL = os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = logoutURL
	}

	// Validate required environment variables
	if clientID == "" || region == "" || userPoolID == "" {
		log.Fatal("Missing required environment variables: COGNITO_CLIENT_ID, COGNITO_REGION, COGNITO_USER_POOL_ID")
	}

	// Construct issuer URL from region and user pool ID
	issuerURL = fmt.Sprintf("https://cognito-idp.%s.amazonaws.com/%s", region, userPoolID)

	// Initialize OIDC provider
	var err error
	provider, err = oidc.NewProvider(context.Background(), issuerURL)
	if err != nil {
		log.Fatalf("Failed to create OIDC provider: %v", err)
	}

	// Set up OAuth2 config
	oauth2Config = oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		Endpoint:     provider.Endpoint(),
		Scopes:       []string{oidc.ScopeOpenID, "email", "phone", "profile"},
	}

	log.Printf("OIDC Provider initialized with issuer: %s", issuerURL)
}

// =============================================================================
// Handlers
// =============================================================================

// handleHome renders the home page with a login link.
func handleHome(w http.ResponseWriter, r *http.Request) {
	html := `
		<!DOCTYPE html>
		<html>
		<head>
			<title>CloudRetail - User Service</title>
			<style>
				body { font-family: Arial, sans-serif; max-width: 800px; margin: 50px auto; padding: 20px; }
				h1 { color: #333; }
				a { display: inline-block; padding: 10px 20px; background: #007bff; color: white; text-decoration: none; border-radius: 5px; }
				a:hover { background: #0056b3; }
			</style>
		</head>
		<body>
			<h1>Welcome to CloudRetail User Service</h1>
			<p>This service handles authentication via AWS Cognito.</p>
			<a href="/login">Login with Cognito</a>
		</body>
		</html>`
	fmt.Fprint(w, html)
}

// handleLogin redirects users to the Cognito authorization endpoint.
func handleLogin(w http.ResponseWriter, r *http.Request) {
	// In production, use a secure random string and store in session
	state := "secure-random-state"
	url := oauth2Config.AuthCodeURL(state, oauth2.AccessTypeOffline)
	http.Redirect(w, r, url, http.StatusFound)
}

// handleCallback exchanges the authorization code for tokens and redirects to the frontend.
func handleCallback(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	// Check for errors from Cognito
	if errMsg := r.URL.Query().Get("error"); errMsg != "" {
		errDesc := r.URL.Query().Get("error_description")
		redirectURL := fmt.Sprintf("%s/callback?error=%s&error_description=%s",
			frontendURL, url.QueryEscape(errMsg), url.QueryEscape(errDesc))
		http.Redirect(w, r, redirectURL, http.StatusFound)
		return
	}

	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "Missing authorization code", http.StatusBadRequest)
		return
	}

	// Exchange the authorization code for tokens
	rawToken, err := oauth2Config.Exchange(ctx, code)
	if err != nil {
		log.Printf("Token exchange failed: %v", err)
		redirectURL := fmt.Sprintf("%s/callback?error=%s",
			frontendURL, url.QueryEscape("Token exchange failed: "+err.Error()))
		http.Redirect(w, r, redirectURL, http.StatusFound)
		return
	}

	// Extract tokens
	accessToken := rawToken.AccessToken
	refreshToken := rawToken.RefreshToken

	// Extract ID token from extras (Cognito returns it in the token response)
	idToken, _ := rawToken.Extra("id_token").(string)
	if idToken == "" {
		// Fallback: use access token as ID token
		idToken = accessToken
	}

	// Redirect to frontend with tokens as query parameters
	redirectURL := fmt.Sprintf("%s/callback?id_token=%s&access_token=%s&refresh_token=%s",
		frontendURL,
		url.QueryEscape(idToken),
		url.QueryEscape(accessToken),
		url.QueryEscape(refreshToken),
	)

	log.Printf("Redirecting to frontend callback: %s/callback?...", frontendURL)
	http.Redirect(w, r, redirectURL, http.StatusFound)
}

// handleLogout redirects to the frontend home page.
func handleLogout(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, frontendURL, http.StatusFound)
}

// handleHealth returns the health status of the service.
func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, `{"status":"healthy"}`)
}

// =============================================================================
// Main
// =============================================================================

// enableCORS adds CORS headers to allow frontend requests.
func enableCORS(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next(w, r)
	}
}

func main() {
	// Set up routes
	http.HandleFunc("/", handleHome)
	http.HandleFunc("/login", handleLogin)
	http.HandleFunc("/logout", handleLogout)
	http.HandleFunc("/callback", handleCallback)
	http.HandleFunc("/health", enableCORS(handleHealth))

	// Get port from environment or default to 8080
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("Server is running on http://localhost:%s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
