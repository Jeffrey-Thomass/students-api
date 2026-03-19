package sqlite

import (
	"database/sql"

	"github.com/Jeffrey-Thomass/students-api/internal/config"
	_ "github.com/mattn/go-sqlite3"
)

type Sqlite struct {
	Db *sql.DB
}

func New(cfg *config.Config) (*Sqlite, error) {
	db, err := sql.Open("sqlite3", cfg.StoragePath)
	if err != nil {
		return nil, err
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS students (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	name TEXT,
	age INTEGER,
	email TEXT
	)`)
	if err != nil {
		return nil, err
	}

	return &Sqlite{
		Db: db,
	}, nil

}
