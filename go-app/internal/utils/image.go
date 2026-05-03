package utils

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
)

func SaveUploadedFile(file *multipart.FileHeader, uploadDir, prefix string) (string, error) {
	src, err := file.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()

	data, err := io.ReadAll(src)
	if err != nil {
		return "", err
	}

	return SaveBytes(data, uploadDir, prefix, filepath.Ext(file.Filename))
}

func SaveBase64Image(dataURL, uploadDir, prefix string) (string, error) {
	parts := strings.SplitN(dataURL, ";base64,", 2)
	if len(parts) != 2 {
		return "", fmt.Errorf("invalid base64 data URL")
	}

	ext := ""
	formatParts := strings.SplitN(parts[0], "/", 2)
	if len(formatParts) == 2 {
		ext = "." + formatParts[1]
	}
	if ext == ".jpeg" {
		ext = ".jpg"
	}

	data, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil {
		return "", err
	}

	return SaveBytes(data, uploadDir, prefix, ext)
}

func SaveBytes(data []byte, uploadDir, prefix, ext string) (string, error) {
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		return "", err
	}

	hash := sha256.Sum256(data)
	filename := hex.EncodeToString(hash[:16]) + ext
	fullPath := filepath.Join(uploadDir, filename)

	if err := os.WriteFile(fullPath, data, 0644); err != nil {
		return "", err
	}

	relativePath := filepath.Join(prefix, filename)
	return "/media/" + relativePath, nil
}
