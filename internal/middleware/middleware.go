package middleware

import (
	"encoding/base64"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/cors"
	"github.com/gorilla/csrf"
)

// GetCSRFMiddleware returns the CSRF protection middleware
func GetCSRFMiddleware() func(http.Handler) http.Handler {
	// Get CSRF key from environment variable
	csrfKeyStr := os.Getenv("CSRF_KEY")
	if csrfKeyStr == "" {
		// Generate a key for development - in production, use environment variable
		log.Println("Warning: CSRF_KEY not set, generating a temporary one")
		tempKey := make([]byte, 32)
		csrfKeyStr = base64.StdEncoding.EncodeToString(tempKey)
	}

	csrfKey, err := base64.StdEncoding.DecodeString(csrfKeyStr)
	if err != nil {
		log.Fatalf("Error decoding CSRF key: %v", err)
	}

	// Use secure cookies in production only
	env := os.Getenv("APP_ENV")
	secure := env == "PROD"

	return csrf.Protect(
		csrfKey,
		csrf.Path("/"),      // Use the same path for all requests
		csrf.Secure(secure), // True in PROD, false in LOCAL
		csrf.HttpOnly(true),
		csrf.SameSite(csrf.SameSiteStrictMode),
	)
}

// Cors returns the CORS middleware based on environment
func Cors() func(http.Handler) http.Handler {
	// Check environment variable for mode
	env := os.Getenv("APP_ENV")

	// Use production CORS settings if in PROD mode
	if env == "PROD" {
		log.Println("Using production CORS settings")
		return productionCors()
	}

	// Default to development CORS settings
	log.Println("Using development CORS settings")
	return developmentCors()
}

// developmentCors returns CORS settings for local development
func developmentCors() func(http.Handler) http.Handler {
	return cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300, // Maximum value not ignored by any major browsers
	})
}

// productionCors returns CORS settings for production
func productionCors() func(http.Handler) http.Handler {
	return cors.Handler(cors.Options{
		// Only allow specific origins
		AllowedOrigins: []string{
			"https://kanji-go-pdjzxrqjaq-uc.a.run.app",
		},
		// Restrict methods as needed
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		// Only allow necessary headers
		AllowedHeaders: []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders: []string{"Link"},
		// Control credentials in production
		AllowCredentials: true,
		MaxAge:           300,
	})
}
