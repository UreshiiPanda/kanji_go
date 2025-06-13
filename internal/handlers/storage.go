package handlers

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"cloud.google.com/go/storage"
	"github.com/google/uuid"
	"google.golang.org/api/iterator"
	"github.com/gorilla/csrf"
)

// Maximum file size (5MB)
const maxUploadSize = 5 << 20

// Storage client instance
var storageClient *storage.Client

// InitStorage initializes the Cloud Storage client
func InitStorage(ctx context.Context) error {
	var err error
	
	// For production, this will use the service account credentials
	// For local development, this uses local gcloud credentials
	storageClient, err = storage.NewClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to create storage client: %w", err)
	}
	
	return nil
}

// CloseStorage closes the storage client
func CloseStorage() {
	if storageClient != nil {
		storageClient.Close()
	}
}

// GetBucketName returns the configured bucket name
func GetBucketName() string {
	bucketName := os.Getenv("BUCKET_NAME")
	if bucketName == "" {
		log.Println("Warning: BUCKET_NAME not set, using default bucket name")
		bucketName = "default-bucket-name"
	}
	return bucketName
}

// UploadHandler handles file uploads to Google Cloud Storage
func UploadHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("Upload handler started")
		
		// Set a reasonable timeout for the upload
		ctx, cancel := context.WithTimeout(r.Context(), 60*time.Second)
		defer cancel()

		// Limit file size
		r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)
		if err := r.ParseMultipartForm(maxUploadSize); err != nil {
			log.Printf("Error parsing multipart form: %v", err)
			http.Error(w, "File too large or invalid form", http.StatusBadRequest)
			return
		}
		
		// Get the file from the form
		file, header, err := r.FormFile("image")
		if err != nil {
			log.Printf("Error getting file from form: %v", err)
			http.Error(w, "Error retrieving file", http.StatusBadRequest)
			return
		}
		defer file.Close()
		
		log.Printf("Received file: %s (size: %d bytes, type: %s)", 
			header.Filename, header.Size, header.Header.Get("Content-Type"))

		// Validate file type
		if !isAllowedFileType(header.Filename) {
			log.Printf("Invalid file type: %s", filepath.Ext(header.Filename))
			http.Error(w, "Invalid file type. Only jpg, jpeg, png, and gif are allowed", http.StatusBadRequest)
			return
		}

		// Generate a unique filename
		filename := generateUniqueFilename(header.Filename)
		log.Printf("Generated unique filename: %s", filename)
		
		// Get bucket name from environment
		bucketName := GetBucketName()
		
		// Check if storageClient is initialized
		if storageClient == nil {
			log.Println("Storage client not initialized, initializing now")
			if err := InitStorage(ctx); err != nil {
				log.Printf("Failed to initialize storage client: %v", err)
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}
			defer CloseStorage()
		}
		
		// Upload the file to Google Cloud Storage
		objectName := fmt.Sprintf("uploads/%s", filename)
		log.Printf("Uploading to object: %s in bucket: %s", objectName, bucketName)
		
		object := storageClient.Bucket(bucketName).Object(objectName)
		wc := object.NewWriter(ctx)
		
		// Set Content-Type based on file extension
		contentType := getContentType(filename)
		wc.ContentType = contentType
		log.Printf("Set content type to: %s", contentType)
		
		// Copy the file to GCS
		bytesWritten, err := io.Copy(wc, file)
		if err != nil {
			log.Printf("Error copying file to GCS: %v", err)
			http.Error(w, "Error uploading file", http.StatusInternalServerError)
			return
		}
		log.Printf("Copied %d bytes to GCS", bytesWritten)
		
		// Close the writer to finalize the upload
		if err := wc.Close(); err != nil {
			log.Printf("Error closing GCS writer: %v", err)
			http.Error(w, "Error finalizing upload", http.StatusInternalServerError)
			return
		}
		log.Println("Successfully closed GCS writer")
		
		// Generate a public URL for the file
		publicURL := fmt.Sprintf("https://storage.googleapis.com/%s/%s", bucketName, objectName)
		log.Printf("Generated public URL: %s", publicURL)
		
		// Get the kanji_char_id from the form (if it exists)
		kanjiID := r.FormValue("kanji_char_id")
		kanjiIDText := ""
		if kanjiID != "" {
			kanjiIDText = fmt.Sprintf("<p>Associated with Kanji ID: %s</p>", kanjiID)
		}
		
		// Return the URL in the response for HTMX
		w.Header().Set("Content-Type", "text/html")
		successHTML := fmt.Sprintf(`
			<div class="upload-success bg-green-100 border border-green-400 text-green-700 px-4 py-3 rounded mb-4">
				<p>File uploaded successfully!</p>
				%s
				<div class="mt-2">
					<img src="%s" alt="Uploaded image" class="max-w-full h-auto rounded shadow" style="max-height: 200px;">
				</div>
				<p class="mt-2 text-sm">
					<a href="%s" target="_blank" class="text-blue-600 hover:text-blue-800">View full image</a>
				</p>
				<input type="hidden" name="imageURL" value="%s">
			</div>
		`, kanjiIDText, publicURL, publicURL, publicURL)
		
		log.Println("Upload handler completed successfully")
		w.Write([]byte(successHTML))
	}
}

// ServeFileHandler serves a file from Google Cloud Storage
func ServeFileHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Extract the object name from the request
		// Assumes path format like /files/{objectName}
		objectName := strings.TrimPrefix(r.URL.Path, "/files/")
		if objectName == "" {
			http.Error(w, "File not specified", http.StatusBadRequest)
			return
		}
		
		// Set a reasonable timeout
		ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
		defer cancel()
		
		// Get bucket name from environment
		bucketName := GetBucketName()
		
		// Get the object from GCS
		object := storageClient.Bucket(bucketName).Object(objectName)
		reader, err := object.NewReader(ctx)
		if err != nil {
			log.Printf("Error creating reader for object %s: %v", objectName, err)
			http.Error(w, "File not found", http.StatusNotFound)
			return
		}
		defer reader.Close()
		
		// Set content type based on the object's metadata
		w.Header().Set("Content-Type", reader.Attrs.ContentType)
		
		// Copy the file contents to the response
		if _, err := io.Copy(w, reader); err != nil {
			log.Printf("Error serving file: %v", err)
			http.Error(w, "Error serving file", http.StatusInternalServerError)
			return
		}
	}
}

// isAllowedFileType checks if the file has an allowed extension
func isAllowedFileType(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".jpg", ".jpeg", ".png", ".gif":
		return true
	}
	return false
}

// generateUniqueFilename creates a unique filename to prevent collisions
func generateUniqueFilename(originalFilename string) string {
	ext := filepath.Ext(originalFilename)
	return uuid.New().String() + ext
}

// getContentType determines the content type based on file extension
func getContentType(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".gif":
		return "image/gif"
	default:
		return "application/octet-stream" // Default content type
	}
}

// DeleteFileHandler deletes a file from Google Cloud Storage
func DeleteFileHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Only allow POST requests for deletion
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		
		// Extract the object name from the request
		if err := r.ParseForm(); err != nil {
			log.Printf("Error parsing form: %v", err)
			http.Error(w, "Error parsing form", http.StatusBadRequest)
			return
		}
		
		objectName := r.FormValue("objectName")
		if objectName == "" {
			http.Error(w, "Object name not provided", http.StatusBadRequest)
			return
		}
		
		log.Printf("Request to delete object: %s", objectName)
		
		// Set a reasonable timeout
		ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
		defer cancel()
		
		// Get bucket name from environment
		bucketName := GetBucketName()
		
		// Check if storageClient is initialized
		if storageClient == nil {
			log.Println("Storage client not initialized, initializing now")
			if err := InitStorage(ctx); err != nil {
				log.Printf("Failed to initialize storage client: %v", err)
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}
			defer CloseStorage()
		}
		
		// Delete the object
		object := storageClient.Bucket(bucketName).Object(objectName)
		if err := object.Delete(ctx); err != nil {
			log.Printf("Error deleting object %s: %v", objectName, err)
			http.Error(w, "Error deleting file", http.StatusInternalServerError)
			return
		}
		
		log.Printf("Successfully deleted object: %s", objectName)
		
		// Get CSRF token for the return HTML
		csrfField := csrf.TemplateField(r)
		
		// Return success response for HTMX
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`
			<div class="delete-success bg-blue-100 border border-blue-400 text-blue-700 px-4 py-3 rounded mb-4">
				<p>File deleted successfully!</p>
				<form hx-get="/list-files" hx-target="#files-list" class="mt-2">
					` + string(csrfField) + `
					<button type="submit" class="bg-blue-500 hover:bg-blue-700 text-white text-xs py-1 px-2 rounded">
						Refresh File List
					</button>
				</form>
			</div>
		`))
	}
}

// ListFilesHandler lists files in the Cloud Storage bucket
func ListFilesHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
		defer cancel()
		
		// Check if storageClient is initialized - ADDED THIS CHECK
		if storageClient == nil {
			log.Println("Storage client not initialized, initializing now")
			if err := InitStorage(ctx); err != nil {
				log.Printf("Failed to initialize storage client: %v", err)
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}
			// Don't defer CloseStorage() here as it will close immediately
			// We'll keep the client for future requests
		}
		
		// Get bucket name from environment
		bucketName := GetBucketName()
		
		// Get CSRF token for the forms
		csrfField := csrf.TemplateField(r)
		
		// List objects in the bucket
		it := storageClient.Bucket(bucketName).Objects(ctx, &storage.Query{
			Prefix: "uploads/", // Optional: filter by prefix
		})
		
		// Start the HTML response
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`
			<div class="bg-white p-4 rounded shadow">
				<h3 class="text-lg font-bold mb-2">Files in Bucket</h3>
				<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
		`))
		
		// Track if we found any files
		count := 0
		
		// Process each object
		for {
			attrs, err := it.Next()
			if err == iterator.Done {
				break
			}
			if err != nil {
				log.Printf("Error iterating bucket objects: %v", err)
				http.Error(w, "Error listing files", http.StatusInternalServerError)
				return
			}
			
			count++
			
			// Generate a public URL for the file
			publicURL := fmt.Sprintf("https://storage.googleapis.com/%s/%s", bucketName, attrs.Name)
			
			// Only display images
			if isAllowedFileType(attrs.Name) {
				fileHTML := fmt.Sprintf(`
					<div class="border border-gray-200 rounded-lg p-3">
						<div class="mb-2">
							<img src="%s" alt="%s" class="max-w-full h-auto rounded max-h-32 mx-auto">
						</div>
						<div class="text-sm text-gray-700 truncate">
							<p>Name: %s</p>
							<p>Size: %d KB</p>
							<p>Created: %s</p>
							<form hx-post="/delete-file" hx-target="#files-list" class="mt-2">
								%s
								<input type="hidden" name="objectName" value="%s">
								<button type="submit" class="bg-red-500 hover:bg-red-700 text-white text-xs py-1 px-2 rounded">
									Delete
								</button>
							</form>
						</div>
					</div>
				`, publicURL, attrs.Name, attrs.Name, attrs.Size/1024, attrs.Created.Format("2006-01-02"), csrfField, attrs.Name)
				
				w.Write([]byte(fileHTML))
			}
		}
		
		// If no files found, show a message
		if count == 0 {
			w.Write([]byte(`
				<div class="col-span-3 text-center py-4 text-gray-500">
					No files found in the bucket.
				</div>
			`))
		}
		
		// Close the HTML
		w.Write([]byte(`
				</div>
			</div>
		`))
	}
}
