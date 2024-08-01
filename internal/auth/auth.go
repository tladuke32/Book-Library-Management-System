package auth

import (
	"Book-Library-Management-System/internal/models"
	"database/sql"
	"fmt"
	"github.com/gorilla/sessions"
	"golang.org/x/crypto/bcrypt"
	"net/http"
)

var store = sessions.NewCookieStore([]byte("super-secret-key"))

func RegisterUser(db *sql.DB, username, password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	query := `INSERT INTO users (username, password) VALUES (?, ?)`
	_, err = db.Exec(query, username, hashedPassword)
	if err != nil {
		return err
	}
	return nil
}

func LoginUser(db *sql.DB, username, password string) (*models.User, error) {
	var user models.User
	query := `SELECT id, username, password FROM users WHERE username = ?`
	err := db.QueryRow(query, username).Scan(&user.ID, &user.Username, &user.Password)
	if err != nil {
		return nil, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return nil, fmt.Errorf("invalid password")
	}

	return &user, nil
}

func CreateSession(userID int, w http.ResponseWriter, r *http.Request) error {
	session, _ := store.Get(r, "session")
	session.Values["user_id"] = userID
	return session.Save(r, w)
}

func GetUserIDFromSession(r *http.Request) (int, error) {
	session, _ := store.Get(r, "session")
	userID, ok := session.Values["user_id"].(int)
	if !ok {
		return 0, fmt.Errorf("no user ID in session")
	}
	return userID, nil
}

func LogoutUser(w http.ResponseWriter, r *http.Request) error {
	session, _ := store.Get(r, "session")
	delete(session.Values, "user_id")
	return session.Save(r, w)
}
