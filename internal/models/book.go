package models

import "time"

type Book struct {
	ID            int       `json:"id"`
	Title         string    `json:"title"`
	Author        string    `json:"author"`
	PublishedDate time.Time `json:"published_date"`
	ISBN          string    `json:"isbn"`
	Categories    string    `json:"categories"`
	Rating        int       `json:"rating"`
	UserID        int       `json:"user_id"`
}
