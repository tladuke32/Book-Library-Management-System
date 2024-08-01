package db

//func InitDB(dataSourceName string) (*sql.DB, error) {
//	db, err := sql.Open("sqlite3", dataSourceName)
//	if err != nil {
//		return nil, err
//	}
//
//	if err := CreateUserTable(db); err != nil {
//		return nil, err
//	}
//	if err := CreateSessionTable(db); err != nil {
//		return nil, err
//	}
//	if err := CreateBookTable(db); err != nil {
//		return nil, err
//	}
//
//	return db, nil
//}

//func CreateUserTable(db *sql.DB) error {
//	query := `
//	CREATE TABLE IF NOT EXISTS users (
//	    id INTEGER PRIMARY KEY AUTOINCREMENT,
//	    username TEXT NOT NULL UNIQUE,
//	    password TEXT NOT NULL
//	);
//	`
//	_, err := db.Exec(query)
//	if err != nil {
//		log.Println("Failed to create users table")
//	}
//	return err
//}
//
//func CreateSessionTable(db *sql.DB) error {
//	query := `
//    CREATE TABLE IF NOT EXISTS sessions (
//       id INTEGER PRIMARY KEY AUTOINCREMENT,
//        user_id INTEGER NOT NULL,
//       session_token TEXT NOT NULL,
//       FOREIGN KEY (user_id) REFERENCES users(id)
//    );
//  `
//	_, err := db.Exec(query)
//	if err != nil {
//		log.Println("Failed to create sessions table:", err)
//	}
//	return err
//}
//
//func CreateBookTable(db *sql.DB) error {
//	query := `
//  CREATE TABLE IF NOT EXISTS books (
//        id INTEGER PRIMARY KEY AUTOINCREMENT,
//        title TEXT NOT NULL,
//        author TEXT NOT NULL,
//        published_date TEXT NOT NULL,
//        isbn TEXT NOT NULL,
//        user_id INTEGER NOT NULL,
//        categories TEXT,
//        rating INTEGER,
//        FOREIGN KEY (user_id) REFERENCES users(id)
//    );
//    `
//	_, err := db.Exec(query)
//	if err != nil {
//		log.Println("Failed to create books table:", err)
//	}
//	return err
//}
