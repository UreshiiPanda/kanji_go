package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"cloud.google.com/go/storage"
	"github.com/joho/godotenv"
	"google.golang.org/api/iterator"
)

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Get bucket name
	bucketName := os.Getenv("BUCKET_NAME")
	if bucketName == "" {
		log.Fatal("BUCKET_NAME environment variable not set")
	}

	// Create context
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// Create storage client
	client, err := storage.NewClient(ctx)
	if err != nil {
		log.Fatalf("Failed to create storage client: %v", err)
	}
	defer client.Close()

	// Test if we can connect to the bucket
	bucketHandle := client.Bucket(bucketName)
	_, err = bucketHandle.Attrs(ctx)
	if err != nil {
		log.Fatalf("Failed to get bucket attributes: %v", err)
	}
	fmt.Printf("✅ Successfully connected to bucket: %s\n", bucketName)

	// Upload a test file
	testContent := "Hello from the bucket test script!"
	testFileName := fmt.Sprintf("test-file-%d.txt", time.Now().Unix())
	
	object := bucketHandle.Object(testFileName)
	writer := object.NewWriter(ctx)
	
	if _, err := writer.Write([]byte(testContent)); err != nil {
		log.Fatalf("Failed to write test file: %v", err)
	}
	
	if err := writer.Close(); err != nil {
		log.Fatalf("Failed to close writer: %v", err)
	}
	
	fmt.Printf("✅ Successfully uploaded test file: %s\n", testFileName)

	// Make the file public
	if err := object.ACL().Set(ctx, storage.AllUsers, storage.RoleReader); err != nil {
		log.Printf("Warning: Failed to make file public: %v", err)
	}

	// List files in bucket
	fmt.Println("\nListing files in bucket:")
	fmt.Println("-----------------------")
	
	it := bucketHandle.Objects(ctx, nil)
	count := 0
	
	for {
		attrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Fatalf("Error iterating bucket objects: %v", err)
		}
		
		count++
		fmt.Printf("%d. %s (size: %d bytes, created: %s)\n", 
			count, attrs.Name, attrs.Size, attrs.Created.Format("2006-01-02 15:04:05"))
	}
	
	if count == 0 {
		fmt.Println("No files found in bucket (this is unexpected!)")
	} else {
		fmt.Printf("✅ Found %d files in bucket\n", count)
	}

	// Read the test file back
	reader, err := object.NewReader(ctx)
	if err != nil {
		log.Fatalf("Failed to create reader: %v", err)
	}
	defer reader.Close()
	
	content, err := io.ReadAll(reader)
	if err != nil {
		log.Fatalf("Failed to read file: %v", err)
	}
	
	if string(content) != testContent {
		log.Fatalf("File content doesn't match: got %q, want %q", string(content), testContent)
	}
	
	fmt.Printf("✅ Successfully read test file, content matches: %q\n", testContent)

	// Clean up - delete the test file
	if err := object.Delete(ctx); err != nil {
		log.Fatalf("Failed to delete test file: %v", err)
	}
	
	fmt.Printf("✅ Successfully deleted test file: %s\n", testFileName)
	fmt.Println("\n🎉 All tests passed! Your bucket is working correctly.")
}
