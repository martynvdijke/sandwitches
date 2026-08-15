package api

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

const (
	defaultPageSize = 50
	maxPageSize     = 200
)

type pagination struct {
	limit  int
	offset int
}

// parsePagination reads and validates the limit/offset query parameters with
// defaults limit=50 (capped at 200) and offset=0. Invalid values (non-numeric
// or negative) produce a 422 validation error envelope and return ok=false.
func parsePagination(c *gin.Context) (pagination, bool) {
	pg := pagination{limit: defaultPageSize, offset: 0}

	if v := c.Query("limit"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 0 {
			apiError(c, http.StatusUnprocessableEntity, "limit must be a non-negative integer")
			return pg, false
		}
		if n > maxPageSize {
			n = maxPageSize
		}
		pg.limit = n
	}

	if v := c.Query("offset"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 0 {
			apiError(c, http.StatusUnprocessableEntity, "offset must be a non-negative integer")
			return pg, false
		}
		pg.offset = n
	}

	return pg, true
}

// wantsPagination reports whether the client explicitly requested pagination
// via limit and/or offset query parameters. Only such requests receive the
// {items, total} envelope; bare requests keep the plain-array response shape
// for backward compatibility (e.g. TRMNL polling).
func wantsPagination(c *gin.Context) bool {
	_, hasLimit := c.GetQuery("limit")
	_, hasOffset := c.GetQuery("offset")
	return hasLimit || hasOffset
}
