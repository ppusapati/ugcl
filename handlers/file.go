package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"cloud.google.com/go/storage"
)

const (
	projectID  = "ugcl-461407" // Your GCP project ID
	bucketName = "sreeugcl"    // Replace with your GCS bucket
)

func UploadFile(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	// Parse the multipart form
	if err := r.ParseMultipartForm(50 << 20); err != nil {
		http.Error(w, "bad multipart form: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Get the uploaded file
	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "missing file field: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Initialize GCS client
	client, err := storage.NewClient(ctx)
	if err != nil {
		http.Error(w, "failed to create GCS client: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer client.Close()

	// Create object in bucket
	object := client.Bucket(bucketName).Object(header.Filename)
	writer := object.NewWriter(ctx)

	// Optional: Make the file publicly accessible
	object.ACL().Set(ctx, storage.AllUsers, storage.RoleReader)

	// Upload file content
	if _, err := io.Copy(writer, file); err != nil {
		http.Error(w, "failed to upload to GCS: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if err := writer.Close(); err != nil {
		http.Error(w, "failed to finalize upload: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Return public GCS URL
	url := fmt.Sprintf("https://storage.googleapis.com/%s/%s", bucketName, header.Filename)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"url": url})
}
