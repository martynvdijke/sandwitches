package handlers

import (
	"encoding/xml"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/martynvdijke/sandwitches-go/internal/database"
)

type rssFeed struct {
	XMLName xml.Name `xml:"rss"`
	Version string   `xml:"version,attr"`
	Channel rssChannel
}

type rssChannel struct {
	Title       string    `xml:"title"`
	Link        string    `xml:"link"`
	Description string    `xml:"description"`
	Items       []rssItem `xml:"item"`
}

type rssItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
}

func LatestRecipesFeed(c *gin.Context) {
	var recipes []database.Recipe
	database.DB.Order("created_at DESC").Limit(5).Find(&recipes)

	base := baseURL()

	items := make([]rssItem, len(recipes))
	for i, r := range recipes {
		items[i] = rssItem{
			Title:       r.Title,
			Link:        base + "/recipes/" + r.Slug,
			Description: r.Description,
		}
	}

	feed := rssFeed{
		Version: "2.0",
		Channel: rssChannel{
			Title:       "Sandwitches - Latest Recipes",
			Link:        base + "/",
			Description: "Updates on the newest recipes added to Sandwitches.",
			Items:       items,
		},
	}

	c.Header("Content-Type", "application/rss+xml; charset=utf-8")
	c.XML(http.StatusOK, feed)
}

func baseURL() string {
	url := os.Getenv("BASE_URL")
	if url == "" {
		url = "http://localhost"
	}
	return url
}

func init() {
	_ = time.Now
}
