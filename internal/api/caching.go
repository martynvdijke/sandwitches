package api

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
)

// etagFor computes a content hash wrapped in quotes suitable for the ETag
// header.
func etagFor(body []byte) string {
	sum := sha256.Sum256(body)
	return `"` + hex.EncodeToString(sum[:16]) + `"`
}

// serveWithETag marshals payload, sets an ETag header, and responds 304 when
// the client sent a matching If-None-Match header; otherwise it writes the
// JSON body.
func serveWithETag(c *gin.Context, payload interface{}) {
	body, err := json.Marshal(payload)
	if err != nil {
		apiError(c, http.StatusInternalServerError, "Failed to encode response")
		return
	}
	serveETagBody(c, body)
}

// serveETagBody writes the given pre-marshaled JSON body with an ETag header
// (and 304 short-circuit when the client's If-None-Match matches).
func serveETagBody(c *gin.Context, body []byte) {
	etag := etagFor(body)
	c.Header("ETag", etag)
	if match := c.GetHeader("If-None-Match"); match != "" && match == etag {
		c.Status(http.StatusNotModified)
		return
	}
	c.Data(http.StatusOK, "application/json; charset=utf-8", body)
}

// dayRecipeCache caches the rendered recipe-of-the-day payload per local
// calendar day. The cache is invalidated when the recipe count changes so new
// or removed recipes take effect.
type dayRecipeCache struct {
	mu    sync.Mutex
	day   string
	count int64
	body  []byte
}

var recipeOfTheDayCache dayRecipeCache
