package utils

import (
	"fmt"
	"image"
	"image/jpeg"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/image/draw"
)

// ImageExt reports whether p looks like a raster image we can resize.
func ImageExt(p string) bool {
	switch strings.ToLower(filepath.Ext(p)) {
	case ".jpg", ".jpeg", ".png", ".gif":
		return true
	}
	return false
}

// GenerateThumbnail decodes the image at srcPath, scales it down so its
// longest side is at most maxWidth (keeping aspect ratio), and writes a JPEG
// to dstPath. Images already smaller than maxWidth are written as-is (re-encoded).
func GenerateThumbnail(srcPath, dstPath string, maxWidth, quality int) error {
	src, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer src.Close()

	img, _, err := image.Decode(src)
	if err != nil {
		return err
	}

	bounds := img.Bounds()
	w := bounds.Dx()
	h := bounds.Dy()
	if w <= 0 || h <= 0 {
		return fmt.Errorf("invalid image dimensions %dx%d", w, h)
	}

	if w > maxWidth {
		nh := int(float64(h) * float64(maxWidth) / float64(w))
		if nh < 1 {
			nh = 1
		}
		dst := image.NewRGBA(image.Rect(0, 0, maxWidth, nh))
		draw.CatmullRom.Scale(dst, dst.Bounds(), img, bounds, draw.Over, nil)
		img = dst
		w, h = maxWidth, nh
	}

	if err := os.MkdirAll(filepath.Dir(dstPath), 0o755); err != nil {
		return err
	}
	out, err := os.Create(dstPath)
	if err != nil {
		return err
	}
	defer out.Close()

	if quality < 1 {
		quality = 82
	}
	if err := jpeg.Encode(out, img, &jpeg.Options{Quality: quality}); err != nil {
		os.Remove(dstPath)
		return err
	}
	return nil
}
