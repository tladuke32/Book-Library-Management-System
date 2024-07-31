package api

import (
	"Book-Library-Management-System/internal/models"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

const GoogleBooksAPI = "https://www.googleapis.com/books/v1/volumes?q="

type GoogleBooksResponse struct {
	Items []struct {
		VolumeInfo struct {
			Title               string   `json:"title"`
			Authors             []string `json:"authors"`
			PublishedDate       string   `json:"publishedDate"`
			IndustryIdentifiers []struct {
				Type       string `json:"type"`
				Identifier string `json:"identifier"`
			} `json:"industryIdentifiers"`
		} `json:"volumeInfo"`
	} `json:"items"`
}

func FetchBooks(query string) ([]models.Book, error) {
	url := fmt.Sprintf("%s%s", GoogleBooksAPI, query)
	resp, err := http.Get(url)
	if err != nil {
		log.Println("Failed to fetch data from Google Books API:", err)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch data: %s", resp.Status)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var booksResponse GoogleBooksResponse
	if err := json.Unmarshal(body, &booksResponse); err != nil {
		return nil, err
	}

	var books []models.Book
	for _, item := range booksResponse.Items {
		book := models.Book{
			Title:         item.VolumeInfo.Title,
			Author:        strings.Join(item.VolumeInfo.Authors, ", "),
			PublishedDate: item.VolumeInfo.PublishedDate,
		}
		for _, identifier := range item.VolumeInfo.IndustryIdentifiers {
			if identifier.Type == "ISBN_13" {
				book.ISBN = identifier.Identifier
				break
			}
		}
		books = append(books, book)
	}

	return books, nil
}
