package utils

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/image/draw"
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

// MaxImageDimension is the longest side (px) uploaded images are scaled down to.
const MaxImageDimension = 1600

// SaveBytes stores data under uploadDir, returning its /media relative path.
// Raster images (jpg/jpeg/png) are resized to at most MaxImageDimension on the
// longest side and re-encoded as JPEG (quality 82) so uploads stay small.
// Other formats (e.g. animated gifs) and undecodable files are stored verbatim.
func SaveBytes(data []byte, uploadDir, prefix, ext string) (string, error) {
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		return "", err
	}

	saveData := data
	saveExt := ext
	if processed, ok := CompressImage(data, MaxImageDimension, 82); ok {
		saveData = processed
		saveExt = ".jpg"
	}

	hash := sha256.Sum256(saveData)
	filename := hex.EncodeToString(hash[:16]) + saveExt
	fullPath := filepath.Join(uploadDir, filename)

	if err := os.WriteFile(fullPath, saveData, 0644); err != nil {
		return "", err
	}

	relativePath := filepath.Join(prefix, filename)
	return "/media/" + relativePath, nil
}

// CompressImage decodes data and, if it is a raster image (jpg/jpeg/png),
// scales it so its longest side is at most maxDim and re-encodes it as a JPEG
// at the given quality. Returns ok=false (data unchanged) for non-image or
// undecodable input.
func CompressImage(data []byte, maxDim, quality int) ([]byte, bool) {
	img, format, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return data, false
	}
	switch format {
	case "jpeg", "png":
	default:
		return data, false
	}

	bounds := img.Bounds()
	w := bounds.Dx()
	h := bounds.Dy()
	if w <= 0 || h <= 0 {
		return data, false
	}

	if w > maxDim {
		nh := int(float64(h) * float64(maxDim) / float64(w))
		if nh < 1 {
			nh = 1
		}
		dst := image.NewRGBA(image.Rect(0, 0, maxDim, nh))
		draw.CatmullRom.Scale(dst, dst.Bounds(), img, bounds, draw.Over, nil)
		img = dst
		w, h = maxDim, nh
	}

	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: quality}); err != nil {
		return data, false
	}
	return buf.Bytes(), true
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
