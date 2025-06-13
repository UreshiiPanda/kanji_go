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
    env := os.Getenv("APP_ENV")
    
    // For local development, just return a pass-through middleware
    if env != "PROD" {
        log.Println("CSRF protection disabled for local development")
        return func(next http.Handler) http.Handler {
            return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
                next.ServeHTTP(w, r)
            })
        }
    }
    
    // Production environment - use full CSRF protection
    csrfKeyStr := os.Getenv("CSRF_KEY")
    if csrfKeyStr == "" {
        log.Fatalf("CSRF_KEY not set in production environment")
    }
    
    csrfKey, err := base64.StdEncoding.DecodeString(csrfKeyStr)
    if err != nil {
        log.Fatalf("Error decoding CSRF key: %v", err)
    }
    
    // Create CSRF handler WITHOUT specifying domain
    // This makes it use the domain from the request automatically
    return csrf.Protect(
        csrfKey,
        csrf.Path("/"),
        csrf.Secure(true),
        csrf.HttpOnly(true),
        csrf.SameSite(csrf.SameSiteStrictMode),
        // No domain specified = use the domain from the request
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
		AllowCredentials: true,
		MaxAge:           300, // Maximum value not ignored by any major browsers
	})
}

// productionCors returns CORS settings for production
func productionCors() func(http.Handler) http.Handler {
    return cors.Handler(cors.Options{
        // Allow both domains
        AllowedOrigins: []string{
            "https://kanji-go-pdjzxrqjaq-uc.a.run.app",
            "https://kanji-go-111333019928.us-central1.run.app",
        },
        AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
        AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
        ExposedHeaders:   []string{"Link"},
        AllowCredentials: true,
        MaxAge:           300,
    })
}
