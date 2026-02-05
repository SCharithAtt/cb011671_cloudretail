// Package main provides the entry point for the user microservice.
// This service handles user authentication using AWS Cognito with OIDC/OAuth2.
package main

import (
	"context"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/golang-jwt/jwt/v4"
	"github.com/joho/godotenv"
	"golang.org/x/oauth2"
)

// =============================================================================
// Models
// =============================================================================

// ClaimsPage holds data for rendering the claims template.
type ClaimsPage struct {
	AccessToken string
	Claims      jwt.MapClaims
}

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

// handleCallback exchanges the authorization code for tokens and displays user claims.
func handleCallback(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	// Check for errors from Cognito
	if errMsg := r.URL.Query().Get("error"); errMsg != "" {
		errDesc := r.URL.Query().Get("error_description")
		http.Error(w, fmt.Sprintf("Authentication error: %s - %s", errMsg, errDesc), http.StatusBadRequest)
		return
	}

	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "Missing authorization code", http.StatusBadRequest)
		return
	}

	// TODO: Verify state parameter in production
	// state := r.URL.Query().Get("state")

	// Exchange the authorization code for tokens
	rawToken, err := oauth2Config.Exchange(ctx, code)
	if err != nil {
		http.Error(w, "Failed to exchange token: "+err.Error(), http.StatusInternalServerError)
		return
	}

	tokenString := rawToken.AccessToken

	// Parse the token (for production, verify the signature)
	token, _, err := new(jwt.Parser).ParseUnverified(tokenString, jwt.MapClaims{})
	if err != nil {
		http.Error(w, "Error parsing token: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Extract claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		http.Error(w, "Invalid claims", http.StatusBadRequest)
		return
	}

	// Prepare data for rendering
	pageData := ClaimsPage{
		AccessToken: tokenString,
		Claims:      claims,
	}

	// Define the HTML template
	tmpl := `
	<!DOCTYPE html>
	<html>
	<head>
		<title>CloudRetail - User Information</title>
		<style>
			body { font-family: Arial, sans-serif; max-width: 800px; margin: 50px auto; padding: 20px; }
			h1 { color: #333; }
			.token { word-break: break-all; background: #f4f4f4; padding: 10px; border-radius: 5px; font-size: 12px; }
			ul { list-style: none; padding: 0; }
			li { padding: 8px; border-bottom: 1px solid #eee; }
			li strong { display: inline-block; min-width: 150px; }
			a { display: inline-block; margin-top: 20px; padding: 10px 20px; background: #dc3545; color: white; text-decoration: none; border-radius: 5px; }
			a:hover { background: #c82333; }
		</style>
	</head>
	<body>
		<h1>User Information</h1>
		<h2>Access Token</h2>
		<p class="token">{{.AccessToken}}</p>
		<h2>JWT Claims</h2>
		<ul>
			{{range $key, $value := .Claims}}
				<li><strong>{{$key}}:</strong> {{$value}}</li>
			{{end}}
		</ul>
		<a href="/logout">Logout</a>
	</body>
	</html>`

	// Parse and execute the template
	t := template.Must(template.New("claims").Parse(tmpl))
	t.Execute(w, pageData)
}

// handleLogout clears user session and redirects to home.
func handleLogout(w http.ResponseWriter, r *http.Request) {
	// In production, you would:
	// 1. Clear the session/cookie
	// 2. Optionally redirect to Cognito logout endpoint to clear Cognito session:
	//    https://<domain>.auth.<region>.amazoncognito.com/logout?client_id=<client_id>&logout_uri=<logout_uri>

	// For now, just redirect to home
	http.Redirect(w, r, "/", http.StatusFound)
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

func main() {
	// Set up routes
	http.HandleFunc("/", handleHome)
	http.HandleFunc("/login", handleLogin)
	http.HandleFunc("/logout", handleLogout)
	http.HandleFunc("/callback", handleCallback)
	http.HandleFunc("/health", handleHealth)

	// Get port from environment or default to 8080
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("Server is running on http://localhost:%s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
