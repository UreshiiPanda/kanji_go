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
)

//go:embed templates
var templatesFS embed.FS

//go:embed static
var staticFS embed.FS

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
	templatesSubFS, err := fs.Sub(templatesFS, "templates")
	if err != nil {
		log.Fatalf("Error with templates subfolder: %v", err)
	}

	// Parse templates from both pages and fragments folders
	tmpl := template.New("")
	
	// First, parse the base page template
	pageTemplates, err := fs.ReadDir(templatesSubFS, "pages")
	if err != nil {
		log.Fatalf("Error reading pages template directory: %v", err)
	}
	
	for _, pageEntry := range pageTemplates {
		if !pageEntry.IsDir() {
			pageContent, err := fs.ReadFile(templatesSubFS, "pages/"+pageEntry.Name())
			if err != nil {
				log.Fatalf("Error reading page template %s: %v", pageEntry.Name(), err)
			}
			_, err = tmpl.New(pageEntry.Name()).Parse(string(pageContent))
			if err != nil {
				log.Fatalf("Error parsing page template %s: %v", pageEntry.Name(), err)
			}
		}
	}
	
	// Next, parse fragment templates
	fragmentTemplates, err := fs.ReadDir(templatesSubFS, "fragments")
	if err != nil {
		log.Fatalf("Error reading fragments template directory: %v", err)
	}
	
	for _, fragmentEntry := range fragmentTemplates {
		if !fragmentEntry.IsDir() {
			fragmentContent, err := fs.ReadFile(templatesSubFS, "fragments/"+fragmentEntry.Name())
			if err != nil {
				log.Fatalf("Error reading fragment template %s: %v", fragmentEntry.Name(), err)
			}
			_, err = tmpl.New(fragmentEntry.Name()).Parse(string(fragmentContent))
			if err != nil {
				log.Fatalf("Error parsing fragment template %s: %v", fragmentEntry.Name(), err)
			}
		}
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
	r.Get("/", handlers.HomeHandler(tmpl))
	r.Get("/api/kanji", handlers.GetKanjiHandler(dbConn, tmpl))
	r.Get("/dialog", handlers.GetDialogHandler())
	r.Get("/empty", handlers.EmptyHandler())
	r.Get("/list-files", handlers.ListFilesHandler(tmpl))
	r.Post("/upload", handlers.UploadHandler())
	r.Post("/delete-file", handlers.DeleteFileHandler())

	// Start server
	log.Printf("Server starting on port %s", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, r); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}
