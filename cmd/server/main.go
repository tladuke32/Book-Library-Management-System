package main

import (
	"Book-Library-Management-System/internal/api"
	"Book-Library-Management-System/internal/auth"
	"Book-Library-Management-System/internal/db"
	"Book-Library-Management-System/internal/models"
	"context"
	"database/sql"
	"encoding/json"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"strconv"
	"time"
)

var database *sql.DB

func main() {
	log.Println("Starting server")

	var err error
	database, err = db.InitDB("books.db")
	if err != nil {
		log.Fatal(err)
	}

	r := mux.NewRouter()
	r.HandleFunc("/register", RegisterHandler).Methods("POST")
	r.HandleFunc("/login", LoginHandler).Methods("POST")

	apiRouter := r.PathPrefix("/api").Subrouter()
	apiRouter.Use(AuthMiddleware)
	apiRouter.HandleFunc("/", HomeHandler).Methods("GET")
	apiRouter.HandleFunc("/books", GetBooksHandler).Methods("GET")
	apiRouter.HandleFunc("/books", CreateBookHandler).Methods("POST")
	apiRouter.HandleFunc("/books/{id:[0-9]+}", UpdateBookHandler).Methods("PUT")
	apiRouter.HandleFunc("/books/{id:[0-9]+}", DeleteBookHandler).Methods("DELETE")
	apiRouter.HandleFunc("/import-books", ImportBooksHandler).Methods("GET")
	apiRouter.HandleFunc("/logout", LogoutHandler).Methods("POST")

	apiRouter.HandleFunc("/ws", api.HandleConnections)

	r.PathPrefix("/static").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./static/"))))

	log.Println("Setting up HTTP server...")
	http.Handle("/", r)

	log.Println("Server is running on port 8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/static/index.html", http.StatusFound)
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

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	booksChan := make(chan []models.Book)
	errChan := make(chan error)

	go func() {
		books, err := api.FetchBooks(query)
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
