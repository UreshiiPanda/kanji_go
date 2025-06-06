// In main.go
package main

import (
	"embed"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"os"

	"github.com/UreshiiPanda/kanji_go/internal/db"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"
)

//go:embed static
var staticFS embed.FS

//go:embed templates
var templatesFS embed.FS

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

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

	// Serve static files
	staticSubFS, err := fs.Sub(staticFS, "static")
	if err != nil {
		log.Fatalf("Error with static subfolder: %v", err)
	}

	// Initialize Chi router
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Static files
	r.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(http.FS(staticSubFS))))

	// Home handler
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		data := map[string]any{
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

	// API endpoint to fetch kanji
	r.Get("/api/kanji", func(w http.ResponseWriter, r *http.Request) {
		// Query the database
		rows, err := dbConn.Query(`
			SELECT kanji_char_id, kanji_char, romaji_onyomi, romaji_kunyomi, 
			       hiragana_onyomi, hiragana_kunyomi, jlpt_level
			FROM kanji_go.kanji
			ORDER BY kanji_char_id
		`)
		if err != nil {
			log.Printf("Error querying kanji: %v", err)
			http.Error(w, "Failed to retrieve kanji", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		// Build HTML response
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">`))

		// Track if we found any kanji
		count := 0

		// Process each row
		for rows.Next() {
			count++
			var id int
			var kanjiChar, romajiOnyomi, romajiKunyomi, hiraganaOnyomi, hiraganaKunyomi, jlptLevel string

			if err := rows.Scan(&id, &kanjiChar, &romajiOnyomi, &romajiKunyomi,
				&hiraganaOnyomi, &hiraganaKunyomi, &jlptLevel); err != nil {
				log.Printf("Error scanning row: %v", err)
				continue
			}

			// Generate HTML for this kanji
			kanjiHTML := `
			<div class="border border-gray-200 rounded-lg p-4 bg-white shadow-sm hover:shadow-md transition-shadow">
				<div class="text-center mb-2">
					<span class="text-4xl font-bold">` + kanjiChar + `</span>
				</div>
				<div class="text-sm text-gray-700">
					<p><span class="font-semibold">On'yomi:</span> ` + hiraganaOnyomi + ` (` + romajiOnyomi + `)</p>
					<p><span class="font-semibold">Kun'yomi:</span> ` + hiraganaKunyomi + ` (` + romajiKunyomi + `)</p>
					<p><span class="font-semibold">JLPT Level:</span> ` + jlptLevel + `</p>
				</div>
			</div>`

			w.Write([]byte(kanjiHTML))
		}

		// Check for errors from iterating over rows
		if err := rows.Err(); err != nil {
			log.Printf("Error iterating rows: %v", err)
		}

		// If no kanji found, show a message
		if count == 0 {
			w.Write([]byte(`
			<div class="col-span-3 text-center py-4 text-gray-500">
				No kanji found in the database.
			</div>`))
		}

		w.Write([]byte(`</div>`))
	})

	r.Get("/dialog", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		dialogHTML := `
		<div class="overlay blur active"></div>
			<dialog class="active">
			  <h5>Custom overlay</h5>
			  <div>Some text here</div>
			  <nav class="right-align no-space">
				<button class="transparent link" hx-get="/empty" hx-target="#dialog-container" hx-swap="innerHTML">Cancel</button>
				<button class="transparent link">Confirm</button>
			  </nav>
			</dialog>

    	`
		w.Write([]byte(dialogHTML))
	})

	// Empty endpoint remains the same
	r.Get("/empty", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(""))
	})

	// Get port from environment or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Start server
	log.Printf("Server starting on port %s", port)
	if err := http.ListenAndServe(":"+port, r); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}
