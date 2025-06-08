package handlers

import (
	"database/sql"
	"html/template"
	"log"
	"net/http"

	"github.com/gorilla/csrf"
)

// HomeHandler handles the root route
func HomeHandler(tmpl *template.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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
	}
}

// GetKanjiHandler returns all kanji from the database
func GetKanjiHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Query the database
		rows, err := db.Query(`
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
	}
}

// GetDialogHandler returns a BeerCSS dialog
func GetDialogHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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
	}
}

// EmptyHandler returns an empty response (used to clear content)
func EmptyHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(""))
	}
}
