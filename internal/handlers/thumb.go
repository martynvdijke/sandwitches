package handlers

import (
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/martynvdijke/sandwitches-go/internal/utils"
)

const thumbCacheDir = ".thumbs"

// Thumbnail serves a width-limited, cached JPEG of any file under MediaRoot.
// Original files are never modified; resized copies live in <MediaRoot>/.thumbs.
// Falls back to the original file when the source is not a decodable image.
func Thumbnail(c *gin.Context) {
	rel := strings.TrimPrefix(c.Param("path"), "/")
	rel = filepath.Clean(rel)
	if rel == "." || strings.HasPrefix(rel, "..") || filepath.IsAbs(rel) {
		c.Status(http.StatusNotFound)
		return
	}

	srcPath := filepath.Join(mediaRoot, rel)
	info, err := os.Stat(srcPath)
	if err != nil || info.IsDir() {
		c.Status(http.StatusNotFound)
		return
	}

	width := 800
	if w, err := strconv.Atoi(c.DefaultQuery("w", "800")); err == nil && w >= 32 && w <= 2560 {
		width = w
	}

	if !utils.ImageExt(srcPath) {
		c.File(srcPath)
		return
	}

	cachePath := filepath.Join(mediaRoot, thumbCacheDir, strconv.Itoa(width), rel+".jpg")

	if _, err := os.Stat(cachePath); err != nil {
		tmpPath := cachePath + ".tmp"
		if err := utils.GenerateThumbnail(srcPath, tmpPath, width, 82); err != nil {
			os.Remove(tmpPath)
			c.File(srcPath) // fall back to original
			return
		}
		if err := os.Rename(tmpPath, cachePath); err != nil {
			os.Remove(tmpPath)
			c.File(srcPath)
			return
		}
	}

	c.Header("Cache-Control", "public, max-age=86400")
	c.File(cachePath)
}
