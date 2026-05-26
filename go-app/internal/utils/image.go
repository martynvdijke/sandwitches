package utils

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"image/png"
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

func RotateImage(filePath string, clockwise bool) error {
	f, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer f.Close()

	img, format, err := image.Decode(f)
	if err != nil {
		return err
	}
	f.Close()

	bounds := img.Bounds()
	srcW := bounds.Dx()
	srcH := bounds.Dy()

	dstW := srcH
	dstH := srcW

	dst := image.NewRGBA(image.Rect(0, 0, dstW, dstH))

	for y := 0; y < srcH; y++ {
		for x := 0; x < srcW; x++ {
			if clockwise {
				dst.Set(dstW-1-y, x, img.At(bounds.Min.X+x, bounds.Min.Y+y))
			} else {
				dst.Set(y, dstH-1-x, img.At(bounds.Min.X+x, bounds.Min.Y+y))
			}
		}
	}

	outFile, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer outFile.Close()

	switch format {
	case "jpeg", "jpg":
		return jpeg.Encode(outFile, dst, &jpeg.Options{Quality: 90})
	case "png":
		return png.Encode(outFile, dst)
	default:
		return fmt.Errorf("unsupported image format: %s", format)
	}
}

func init() {
	_ = image.NewRGBA
	_ = draw.Draw
	_ = color.RGBA{}
}
