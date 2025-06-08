package middleware

import (
	"encoding/base64"
	"log"
	"net/http"
	"os"
	"path/filepath"

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
	
	return csrf.Protect(
		csrfKey,
		csrf.Path("/"),             // Use the same path for all requests
		csrf.Secure(false),         // Set to true in production (HTTPS)
		csrf.HttpOnly(true),
		csrf.SameSite(csrf.SameSiteStrictMode),
	)
}

// Cors returns the CORS middleware
func Cors() func(http.Handler) http.Handler {
	return cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300, // Maximum value not ignored by any major browsers
	})
}

// SetMimeTypes returns middleware that sets correct MIME types for static files
// In middleware.go
func SetMimeTypes(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Only process static files
        if len(r.URL.Path) > 8 && r.URL.Path[:8] == "/static/" {
            // Force correct MIME type for JavaScript modules
            if filepath.Ext(r.URL.Path) == ".js" {
                w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
            } else if filepath.Ext(r.URL.Path) == ".css" {
                w.Header().Set("Content-Type", "text/css; charset=utf-8")
            }
        }
        next.ServeHTTP(w, r)
    })
}


// PROD CORS
// Cors returns the CORS middleware configured for production
//func Cors() func(http.Handler) http.Handler {
//    return cors.Handler(cors.Options{
//        // Only allow specific origins
//        AllowedOrigins:   []string{"https://yourdomain.com", "https://app.yourdomain.com"},
//        // Restrict methods as needed
//        AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
//        // Only allow necessary headers
//        AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
//        ExposedHeaders:   []string{"Link"},
//        // Be careful with credentials in production
//        AllowCredentials: true,
//        MaxAge:           300,
//    })
//}
