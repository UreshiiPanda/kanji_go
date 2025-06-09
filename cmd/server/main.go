package main

import (
	"embed"
	"html/template"
	"io/fs"
	"log"
	"net/http"

	"github.com/UreshiiPanda/kanji_go/internal/config"
	"github.com/UreshiiPanda/kanji_go/internal/db"
	"github.com/UreshiiPanda/kanji_go/internal/handlers"
	"github.com/UreshiiPanda/kanji_go/internal/middleware"
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/gorilla/csrf"
)

//go:embed static
var staticFS embed.FS

//go:embed templates
var templatesFS embed.FS

func main() {
	// Load configuration
	cfg := config.Load()

	// Get database connection
	dbConn, err := db.GetDBConnection()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer dbConn.Close()

	// Create template
	templates, err := fs.Sub(templatesFS, "templates")
	if err != nil {
		log.Fatalf("Error with templates subfolder: %v", err)
	}

	tmpl, err := template.ParseFS(templates, "base.html")
	if err != nil {
		log.Fatalf("Error parsing template: %v", err)
	}

	// Prepare static files
	staticSubFS, err := fs.Sub(staticFS, "static")
	if err != nil {
		log.Fatalf("Error with static subfolder: %v", err)
	}

	// Initialize Chi router
	r := chi.NewRouter()

	// Basic middleware
	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)
	r.Use(middleware.Cors())
	r.Use(middleware.GetCSRFMiddleware())

	// Static files - using standard file server
	r.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(http.FS(staticSubFS))))

	// Routes
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		data := map[string]any{
			"Title":     "Kanji Go",
			"Message":   "Welcome to Kanji Go!",
			"csrfToken": csrf.Token(r), // Add CSRF token for JavaScript
		}

		err := tmpl.Execute(w, data)
		if err != nil {
			log.Printf("Error executing template: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	})
	
	r.Get("/api/kanji", handlers.GetKanjiHandler(dbConn))
	r.Get("/dialog", handlers.GetDialogHandler())
	r.Get("/empty", handlers.EmptyHandler())

	// Start server
	log.Printf("Server starting on port %s", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, r); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}
