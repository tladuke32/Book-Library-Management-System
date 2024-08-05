package main

import (
	"Book-Library-Management-System/internal/api"
	"Book-Library-Management-System/internal/auth"
	"Book-Library-Management-System/internal/db"
	"Book-Library-Management-System/internal/models"
	"context"
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"html/template"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const GoogleBooksAPI = "https://www.googleapis.com/books/v1/volumes?q="

var database *sql.DB
var tmpl = template.Must(template.ParseFiles("static/index.html"))

func main() {
	log.Println("Starting server")

	var err error
	database, err = db.InitDB("books.db")
	if err != nil {
		panic(err)
	}
	defer database.Close()

	r := mux.NewRouter()
	r.HandleFunc("/register", RegisterHandler).Methods("POST")
	r.HandleFunc("/login", LoginHandler).Methods("POST")
	r.HandleFunc("/", HomeHandler).Methods("GET")
	r.HandleFunc("/dashboard", dashboardHandler)
	r.HandleFunc("/logout", LogoutHandler).Methods("POST")
	r.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "static/favicon.ico")
	})
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))

	apiRouter := r.PathPrefix("/api").Subrouter()
	apiRouter.Use(AuthMiddleware)
	apiRouter.HandleFunc("/books", GetBooksHandler).Methods("GET")
	apiRouter.HandleFunc("/books", CreateBookHandler).Methods("POST")
	apiRouter.HandleFunc("/books/{id:[0-9]+}", UpdateBookHandler).Methods("PUT")
	apiRouter.HandleFunc("/books/{id:[0-9]+}", DeleteBookHandler).Methods("DELETE")
	apiRouter.HandleFunc("/import-books", ImportBooksHandler).Methods("GET")
	apiRouter.HandleFunc("/import-books", ImportBooksFromFileHandler).Methods("POST")
	apiRouter.HandleFunc("/export-books", ExportBooksHandler).Methods("GET")
	apiRouter.HandleFunc("/book", HandleNotifyClients).Methods("POST")

	r.HandleFunc("/ws", api.HandleConnections)

	log.Println("Setting up HTTP server...")
	http.Handle("/", r)

	go api.HandleMessages()

	log.Println("Server is running on port 8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}

func ImportBooksFromFileHandler(w http.ResponseWriter, r *http.Request) {
	file, _, err := r.FormFile("file")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	if err := importBooksFromCSV(file); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func importBooksFromCSV(file multipart.File) error {
	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return err
	}

	for _, record := range records {
		publishedDate, err := time.Parse("2006-01-02", record[3])
		if err != nil {
			return err
		}
		rating, err := strconv.Atoi(record[6])
		if err != nil {
			return err
		}
		book := models.Book{
			Title:         record[0],
			Author:        record[1],
			PublishedDate: publishedDate,
			ISBN:          record[4],
			Categories:    record[5],
			Rating:        rating,
		}
		if err := db.InsertBook(database, book); err != nil {
			return err
		}
	}

	return nil
}

func ExportBooksHandler(w http.ResponseWriter, r *http.Request) {
	userID, err := auth.GetUserIDFromSession(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	books, err := db.GetBooksByUserID(database, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Disposition", "attachment; filename=books.json")
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(books)
}

func dashboardHandler(w http.ResponseWriter, r *http.Request) {
	userID, err := auth.GetUserIDFromSession(r)
	if err != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	tmpl.ExecuteTemplate(w, "index.html", userID)
}

func HandleNotifyClients(w http.ResponseWriter, r *http.Request) {
	var book models.Book
	if err := json.NewDecoder(r.Body).Decode(&book); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	api.NotifyClients(book)
	w.WriteHeader(http.StatusOK)
}

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/static/index.html", http.StatusFound)
}

func GetBooksHandler(w http.ResponseWriter, r *http.Request) {
	userID, err := auth.GetUserIDFromSession(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	books, err := db.GetBooksByUserID(database, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(books)
}

func CreateBookHandler(w http.ResponseWriter, r *http.Request) {
	userID, err := auth.GetUserIDFromSession(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var book models.Book
	if err := json.NewDecoder(r.Body).Decode(&book); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	book.UserID = userID
	book.Categories = r.FormValue("categories")
	rating, err := strconv.Atoi(r.FormValue("rating"))
	if err != nil {
		book.Rating = 0
	} else {
		book.Rating = rating
	}
	if err := db.InsertBook(database, book); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	api.NotifyClients(book)
	go api.HandleMessages()
}

func UpdateBookHandler(w http.ResponseWriter, r *http.Request) {
	_, err := auth.GetUserIDFromSession(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var book models.Book
	if err := json.NewDecoder(r.Body).Decode(&book); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	book.ID = id
	book.Categories = r.FormValue("categories")
	rating, err := strconv.Atoi(r.FormValue("rating"))
	if err != nil {
		book.Rating = 0
	} else {
		book.Rating = rating
	}
	if err := db.UpdateBook(database, book); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func DeleteBookHandler(w http.ResponseWriter, r *http.Request) {
	userID, err := auth.GetUserIDFromSession(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := db.DeleteBookByUser(database, id, userID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func ImportBooksHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("query")
	if query == "" {
		http.Error(w, "Query parameter is required", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	booksChan := make(chan []models.Book)
	errChan := make(chan error)

	go func() {
		books, err := FetchBooks(query)
		if err != nil {
			errChan <- err
			return
		}
		booksChan <- books
	}()

	select {
	case <-ctx.Done():
		http.Error(w, "Request timed out", http.StatusRequestTimeout)
		return
	case err := <-errChan:
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	case books := <-booksChan:
		for _, book := range books {
			if err := db.InsertBook(database, book); err != nil {
				log.Println("Failed to insert book:", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(books)
	}
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

	var booksResponse api.GoogleBooksResponse
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

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	var user models.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := auth.RegisterUser(database, user.Username, user.Password); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	var user models.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	dbUser, err := auth.LoginUser(database, user.Username, user.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	if err := auth.CreateSession(dbUser.ID, w, r); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	if err := auth.LogoutUser(w, r); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, err := auth.GetUserIDFromSession(r); err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}
