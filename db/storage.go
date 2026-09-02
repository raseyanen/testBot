package db

import (
	"database/sql"
	"log"
	"os"
	"time"

	_ "modernc.org/sqlite"
)

type Storage struct {
	db    *sql.DB
	Chats *ChatRepository
}

// NewStorage инициализирует подключение и создает репозитории
func NewStorage(driverName, dataSourceName string) *Storage {
	loc, err := time.LoadLocation(os.Getenv("TZ"))
	if err != nil {
		log.Fatal(err)
	}

	time.Local = loc

	db, err := sql.Open(driverName, dataSourceName)

	if err != nil {
		log.Fatal(err)
	}

	storage := &Storage{
		db:    db,
		Chats: &ChatRepository{db: db},
	}

	err = storage.createTables()
	if err != nil {
		log.Fatal(err)
	}

	return storage
}

func (s *Storage) createTables() error {
	_, err := s.db.Exec(`CREATE TABLE IF NOT EXISTS chats (
		id BIGINT PRIMARY KEY,
		main_topic INTEGER,
		num STRING,
		den STRING,
		title STRING,
		users STRING,
		use_tolstobrow BOOLEAN DEFAULT FALSE,
    	created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	)`)
	if err != nil {
		return err
	}
	return nil
}
