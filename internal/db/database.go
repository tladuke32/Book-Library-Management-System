package db

import (
	"database/sql"
	"log"

	"Book-Library-Management-System/internal/models"
	_ "github.com/mattn/go-sqlite3"
)

func InitDB(filepath string) *sql.DB {
	db, err := sql.Open("sqlite3", filepath)
	if err != nil {
		log.Println("Failed to open database:", err)
		return nil
	}

	if err := db.Ping(); err != nil {
		log.Println("Failed to ping database:", err)
		return nil
	}

	return db
}

func CreateBookTable(db *sql.DB) error {
	createBookTableSQL := `CREATE TABLE IF NOT EXISTS books (
	    "id" INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
	    "title" TEXT,
	    "author" TEXT,
	    "published_date" TEXT,
	    "isbn" TEXT
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
	_, err = statement.Exec(book.Title, book.Author, book.PublishedDate, book.ISBN)
	return err
}

func GetBooks(db *sql.DB) ([]models.Book, error) {
	row, err := db.Query("SELECT id, title, author, published_date, isbn FROM books")
	if err != nil {
		return nil, err
	}
	defer row.Close()

	var books []models.Book
	for row.Next() {
		var book models.Book
		row.Scan(&book.ID, &book.Title, &book.Author, &book.PublishedDate, &book.ISBN)
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
	_, err = statement.Exec(book.Title, book.Author, book.PublishedDate, book.ISBN, book.ID)
	return err
}

func DeleteBook(db *sql.DB, id int) error {
	deleteBookSQL := `DELETE FROM books WHERE id = ?`
	statement, err := db.Prepare(deleteBookSQL)
	if err != nil {
		return err
	}
	_, err = statement.Exec(id)
	return err
}
