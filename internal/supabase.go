package internal

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"

	"github.com/lucsky/cuid"
)

func UploadToSupabase(file multipart.File, handler *multipart.FileHeader) (string, error) {
	// Read file into a buffer
	var buf bytes.Buffer
	_, err := io.Copy(&buf, file)
	if err != nil {
		return "", err
	}

	// Define upload details
	projectRef := os.Getenv("SUPABASE_PROJECT_REF")
	bucket := "images"
	c, err := cuid.NewCrypto(rand.Reader)
    if err != nil {
        return "", err
    }
	filePath := fmt.Sprintf("uploads/%s", c)
	url := fmt.Sprintf("https://%s.supabase.co/storage/v1/object/%s/%s", projectRef, bucket, filePath)

	// Upload file to Supabase
	req, err := http.NewRequest("POST", url, bytes.NewReader(buf.Bytes()))
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", "Bearer "+os.Getenv("SUPABASE_SERVICE_ROLE_KEY"))
	req.Header.Set("Content-Type", handler.Header.Get("Content-Type"))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil || resp.StatusCode >= 300 {
		return "", fmt.Errorf("upload failed with status %s", resp.Status)
	}

	publicURL := fmt.Sprintf("https://%s.supabase.co/storage/v1/object/public/%s/%s", projectRef, bucket, filePath)
	return publicURL, nil
}
