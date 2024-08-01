package models

type Book struct {
	ID            int    `json:"id"`
	Title         string `json:"title"`
	Author        string `json:"author"`
	PublishedDate string `json:"published_date"`
	ISBN          string `json:"isbn"`
	UserID        int    `json:"user_id"`
	Categories    string `json:"categories"`
	Rating        int    `json:"rating"`
}
