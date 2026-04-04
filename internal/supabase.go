package internal

import (
	"bytes"
	"context"
	"crypto/rand"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"

	"github.com/lucsky/cuid"
)

const maxUploadSize = 5 << 20 // 5MB

var allowedMIMETypes = map[string]bool{
	"image/jpeg": true,
	"image/png":  true,
	"image/webp": true,
}

func UploadToSupabase(ctx context.Context, file multipart.File, handler *multipart.FileHeader) (string, error) {
	// Validate env vars
	projectRef := os.Getenv("SUPABASE_PROJECT_REF")
	if projectRef == "" {
		return "", fmt.Errorf("SUPABASE_PROJECT_REF environment variable not set")
	}
	serviceRoleKey := os.Getenv("SUPABASE_SERVICE_ROLE_KEY")
	if serviceRoleKey == "" {
		return "", fmt.Errorf("SUPABASE_SERVICE_ROLE_KEY environment variable not set")
	}

	// Validate MIME type
	contentType := handler.Header.Get("Content-Type")
	if !allowedMIMETypes[contentType] {
		return "", fmt.Errorf("unsupported file type: %s", contentType)
	}

	// Read file into buffer, enforcing size limit
	limited := io.LimitReader(file, maxUploadSize+1)
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, limited); err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}
	if buf.Len() > maxUploadSize {
		return "", fmt.Errorf("file exceeds maximum size of 5MB")
	}

	// Generate unique file path
	c, err := cuid.NewCrypto(rand.Reader)
	if err != nil {
		return "", fmt.Errorf("failed to generate file ID: %w", err)
	}

	bucket := "images"
	filePath := fmt.Sprintf("uploads/%s", c)
	url := fmt.Sprintf("https://%s.supabase.co/storage/v1/object/%s/%s", projectRef, bucket, filePath)

	// Build and send upload request
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(buf.Bytes()))
	if err != nil {
		return "", fmt.Errorf("failed to create upload request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+serviceRoleKey)
	req.Header.Set("Content-Type", contentType)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("upload request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return "", fmt.Errorf("upload failed with status %s", resp.Status)
	}

	publicURL := fmt.Sprintf("https://%s.supabase.co/storage/v1/object/public/%s/%s", projectRef, bucket, filePath)
	return publicURL, nil
}