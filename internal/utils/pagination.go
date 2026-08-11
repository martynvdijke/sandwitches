package utils

import (
	"math"
	"strconv"

	"github.com/gin-gonic/gin"
)

type Pagination struct {
	Page       int
	PerPage    int
	Total      int64
	TotalPages int
	HasPrev    bool
	HasNext    bool
	PrevPage   int
	NextPage   int
	Pages      []int
	Offset     int
}

func GetPagination(c *gin.Context, defaultPerPage int) (page, perPage, offset int) {
	page, _ = strconv.Atoi(c.Query("page"))
	if page < 1 {
		page = 1
	}
	perPage = defaultPerPage
	if pp := c.Query("per_page"); pp != "" {
		if v, err := strconv.Atoi(pp); err == nil && v > 0 && v <= 100 {
			perPage = v
		}
	}
	offset = (page - 1) * perPage
	return
}

func NewPagination(page, perPage int, total int64) Pagination {
	totalPages := int(math.Ceil(float64(total) / float64(perPage)))
	if totalPages < 1 {
		totalPages = 1
	}

	p := Pagination{
		Page:       page,
		PerPage:    perPage,
		Total:      total,
		TotalPages: totalPages,
		Offset:     (page - 1) * perPage,
		HasPrev:    page > 1,
		HasNext:    page < totalPages,
		PrevPage:   page - 1,
		NextPage:   page + 1,
	}

	start := 1
	end := totalPages
	if totalPages > 7 {
		start = page - 3
		if start < 1 {
			start = 1
		}
		end = start + 6
		if end > totalPages {
			end = totalPages
			start = end - 6
			if start < 1 {
				start = 1
			}
		}
	}
	for i := start; i <= end; i++ {
		p.Pages = append(p.Pages, i)
	}

	return p
}
