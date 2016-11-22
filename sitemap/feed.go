package sitemap

import (
	"encoding/json"
	"time"

	"github.com/shuyaoyimei/gofeed/extensions"
)

// Feed is an RSS Feed
type Feed struct {
	Title    string  `json:"title,omitempty"`
	Items    []*Item `json:"items,omitempty"`
	Language string  `json:"language,omitempty"`
	Version  string  `json:"version,omitempty"`
}

func (f Feed) String() string {
	json, _ := json.MarshalIndent(f, "", "    ")
	return string(json)
}

// Item is an RSS Item
type Item struct {
	Title         string         `json:"title,omitempty"`
	Link          string         `json:"link,omitempty"`
	Image         *Image         `json:"image,omitempty"`
	PubDate       string         `json:"pubDate,omitempty"`
	PubDateParsed *time.Time     `json:"pubDateParsed,omitempty"`
	Extensions    ext.Extensions `json:"extensions,omitempty"`
}

// Image is an image that represents the feed
type Image struct {
	Link string `json:"link,omitempty"`
}

//News is a mid status for item
type News struct {
	Name            string `json:"name,omitempty"`
	Title           string `json:"title,omitempty"`
	Language        string `json:"language,omitempty"`
	PublicationDate string `json:"publicationdate,omitempty"`
}
