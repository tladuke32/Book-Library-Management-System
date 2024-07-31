package main

import (
	"Book-Library-Management-System/internal/api"
	"Book-Library-Management-System/internal/db"
	"Book-Library-Management-System/internal/models"
	"database/sql"
	"encoding/json"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"strconv"
)

var database *sql.DB

func main() {
	log.Println("Starting server")

	database := db.InitDB("books.db")
	if database == nil {
		log.Fatal("Failed to initialize database")
	} else {
		log.Println("Successfully initialized database")
	}

	err := db.CreateBookTable(database)
	if err != nil {
		log.Fatal("Failed to create books table", err)
	} else {
		log.Println("Books table created")
	}

	r := mux.NewRouter()
	r.HandleFunc("/", HomeHandler).Methods("GET")
	r.HandleFunc("/books", GetBooksHandler).Methods("GET")
	r.HandleFunc("/books", CreateBookHandler).Methods("POST")
	r.HandleFunc("/books/{id:[0-9]+}", UpdateBookHandler).Methods("PUT")
	r.HandleFunc("/books/{id:[0-9]+}", DeleteBookHandler).Methods("DELETE")
	r.HandleFunc("/import-books", ImportBooksHandler).Methods("GET")

	log.Println("Setting up HTTP server...")
	http.Handle("/", r)

	log.Println("Server is running on port 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Welcome to the Book Library!"))
}

func GetBooksHandler(w http.ResponseWriter, r *http.Request) {
	books, err := db.GetBooks(database)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(books)
}

func CreateBookHandler(w http.ResponseWriter, r *http.Request) {

	var book models.Book
	if err := json.NewDecoder(r.Body).Decode(&book); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := db.InsertBook(database, book); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func UpdateBookHandler(w http.ResponseWriter, r *http.Request) {
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

	if err := db.UpdateBook(database, book); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func DeleteBookHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := db.DeleteBook(database, id); err != nil {
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

	books, err := api.FetchBooks(query)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	log.Println(books)
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
