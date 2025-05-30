package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

// POST /api/v1/files/upload
func UploadFile(w http.ResponseWriter, r *http.Request) {
	// 1) Ensure upload dir exists
	uploadDir := "./uploads"
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		http.Error(w, "could not create upload dir: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 2) Parse the multipart form
	if err := r.ParseMultipartForm(50 << 20); err != nil {
		http.Error(w, "bad multipart form: "+err.Error(), http.StatusBadRequest)
		return
	}

	// 3) Grab the file
	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "missing file field: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	// 4) Create destination file
	dstPath := filepath.Join(uploadDir, header.Filename)
	dst, err := os.Create(dstPath)
	if err != nil {
		http.Error(w, "cannot save file: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		http.Error(w, "failed to write file: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 5) Return its public URL
	url := fmt.Sprintf("http://%s/uploads/%s", r.Host, header.Filename)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"url": url})
}
