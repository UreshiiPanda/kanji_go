package main

import (
	"embed"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"os"
)

//go:embed static
var staticFS embed.FS

//go:embed templates
var templatesFS embed.FS

func main() {
	// Create template
	templates, err := fs.Sub(templatesFS, "templates")
	if err != nil {
		log.Fatalf("Error with templates subfolder: %v", err)
	}
	
	tmpl, err := template.ParseFS(templates, "base.html")
	if err != nil {
		log.Fatalf("Error parsing template: %v", err)
	}

	// Serve static files
	staticSubFS, err := fs.Sub(staticFS, "static")
	if err != nil {
		log.Fatalf("Error with static subfolder: %v", err)
	}
	
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(staticSubFS))))

	// Home handler
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		data := map[string]interface{}{
			"Title":   "Kanji Go",
			"Message": "Welcome to Kanji Go!",
		}
		
		err := tmpl.Execute(w, data)
		if err != nil {
			log.Printf("Error executing template: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	})

	// Get port from environment or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Start server
	log.Printf("Server starting on port %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}
