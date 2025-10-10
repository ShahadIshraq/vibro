package handlers

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
	"vibro/src/utils"
)

const (
	maxUploadSize = 10 << 20 // 10 MB
	uploadsDir    = "./uploads"
)

type UploadResponse struct {
	URL      string `json:"url"`
	Filename string `json:"filename"`
}

// UploadImage handles image uploads
func UploadImage() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Limit upload size
		r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)

		// Parse multipart form
		if err := r.ParseMultipartForm(maxUploadSize); err != nil {
			utils.RespondError(w, http.StatusBadRequest, "File too large", err)
			return
		}

		// Get file from form
		file, header, err := r.FormFile("image")
		if err != nil {
			utils.RespondError(w, http.StatusBadRequest, "Invalid file upload", err)
			return
		}
		defer file.Close()

		// Validate file type
		if !isValidImageType(header.Filename) {
			utils.RespondError(w, http.StatusBadRequest, "Invalid file type. Only PNG, JPG, JPEG, GIF, WEBP allowed", nil)
			return
		}

		// Generate unique filename
		filename := generateFilename(header.Filename)
		filePath := filepath.Join(uploadsDir, filename)

		// Create uploads directory if needed
		if err := os.MkdirAll(uploadsDir, 0755); err != nil {
			utils.RespondError(w, http.StatusInternalServerError, "Failed to create upload directory", err)
			return
		}

		// Save file
		dst, err := os.Create(filePath)
		if err != nil {
			utils.RespondError(w, http.StatusInternalServerError, "Failed to save file", err)
			return
		}
		defer dst.Close()

		if _, err := io.Copy(dst, file); err != nil {
			utils.RespondError(w, http.StatusInternalServerError, "Failed to write file", err)
			return
		}

		// Return file URL
		response := UploadResponse{
			URL:      "/uploads/" + filename,
			Filename: filename,
		}

		utils.RespondJSON(w, http.StatusCreated, response)
	}
}

func isValidImageType(filename string) bool {
	ext := filepath.Ext(filename)
	validExts := map[string]bool{
		".png":  true,
		".jpg":  true,
		".jpeg": true,
		".gif":  true,
		".webp": true,
	}
	return validExts[ext]
}

func generateFilename(original string) string {
	ext := filepath.Ext(original)
	hash := sha256.New()
	hash.Write([]byte(fmt.Sprintf("%s-%d", original, time.Now().UnixNano())))
	return hex.EncodeToString(hash.Sum(nil))[:16] + ext
}
