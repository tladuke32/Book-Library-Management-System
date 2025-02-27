package db

import (
	"database/sql"
	"log"

	"Book-Library-Management-System/internal/models"
	_ "github.com/mattn/go-sqlite3"
)

func InitDB(dataSourceName string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", dataSourceName)
	if err != nil {
		return nil, err
	}

	if err := CreateUserTable(db); err != nil {
		return nil, err
	}
	if err := CreateSessionTable(db); err != nil {
		return nil, err
	}
	if err := CreateBookTable(db); err != nil {
		return nil, err
	}

	return db, nil
}

func CreateUserTable(db *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS users (
	    id INTEGER PRIMARY KEY AUTOINCREMENT,
	    username TEXT NOT NULL UNIQUE,
	    password TEXT NOT NULL
	);
	`
	_, err := db.Exec(query)
	if err != nil {
		log.Println("Failed to create users table")
	}
	return err
}

func CreateSessionTable(db *sql.DB) error {
	query := `
    CREATE TABLE IF NOT EXISTS sessions (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        user_id INTEGER NOT NULL,
        session_token TEXT NOT NULL,
        FOREIGN KEY (user_id) REFERENCES users(id)
    );
    `
	_, err := db.Exec(query)
	if err != nil {
		log.Println("Failed to create sessions table:", err)
	}
	return err
}

func CreateBookTable(db *sql.DB) error {
	createBookTableSQL := `
	CREATE TABLE IF NOT EXISTS books (
	    id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
	    title TEXT,
	    author TEXT,
	    published_date TEXT,
	    isbn TEXT,
	    categories TEXT,
	    rating INTEGER,
		user_id INTEGER,
		FOREIGN KEY(user_id) REFERENCES users(id)                            
	);`

	log.Println("Creating book table...")
	statement, err := db.Prepare(createBookTableSQL)
	if err != nil {
		log.Println("Failed to prepare statement:", err)
		return err
	}
	_, err = statement.Exec()
	if err != nil {
		log.Println("Failed to execute statement:", err)
		return err
	}
	log.Println("Book table created")
	return nil
}

func InsertBook(db *sql.DB, book models.Book) error {
	insertBookSQL := `INSERT INTO books (title, author, published_date, isbn) VALUES (?, ?, ?, ?)`
	statement, err := db.Prepare(insertBookSQL)
	if err != nil {
		return err
	}
	_, err = statement.Exec(book.Title, book.Author, book.PublishedDate, book.ISBN, book.Categories, book.Rating, book.UserID)
	return err
}

func GetBooksByUserID(db *sql.DB, userID int) ([]models.Book, error) {
	rows, err := db.Query(`SELECT id, title, author, published_date, isbn FROM books`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var books []models.Book
	for rows.Next() {
		var book models.Book
		if err := rows.Scan(&book.ID, &book.Title, &book.Author, &book.PublishedDate, &book.ISBN, &book.Categories, &book.Rating); err != nil {
			return nil, err
		}
		books = append(books, book)
	}
	return books, nil
}

func UpdateBook(db *sql.DB, book models.Book) error {
	updateBookSQL := `UPDATE books SET title = ?, author = ?, published_date = ?, isbn = ? WHERE id = ?`
	statement, err := db.Prepare(updateBookSQL)
	if err != nil {
		return err
	}
	_, err = statement.Exec(book.Title, book.Author, book.PublishedDate, book.ISBN, book.Categories, book.Rating, book.ID, book.UserID)
	return err
}

func DeleteBookByUser(db *sql.DB, id int, userID int) error {
	_, err := db.Exec(`DELETE FROM books WHERE id = ?`, id, userID)
	return err
}
