package db

import (
	"database/sql"
	"log"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB

func Init(dataSourceName string) error {
	// Garante que o diretório do banco existe
	dir := filepath.Dir(dataSourceName)
	if dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}

	var err error
	DB, err = sql.Open("sqlite3", dataSourceName+"?_journal_mode=WAL&_foreign_keys=on")
	if err != nil {
		return err
	}

	if err = DB.Ping(); err != nil {
		return err
	}

	log.Printf(" ok : Banco SQLite conectado: %s", dataSourceName)
	return migrate()
}

func migrate() error {
	schema := `
	CREATE TABLE IF NOT EXISTS files (
		id          INTEGER PRIMARY KEY AUTOINCREMENT,
		original_name TEXT    NOT NULL,
		stored_name   TEXT    NOT NULL UNIQUE,
		size          INTEGER NOT NULL,
		mime_type     TEXT    NOT NULL,
		description   TEXT    DEFAULT '',
		download_count INTEGER DEFAULT 0,
		created_at    DATETIME DEFAULT (datetime('now','localtime'))
	);

	CREATE INDEX IF NOT EXISTS idx_files_created_at ON files(created_at);
	CREATE INDEX IF NOT EXISTS idx_files_mime_type  ON files(mime_type);
	`

	if _, err := DB.Exec(schema); err != nil {
		return err
	}

	log.Println(" ok : Migrations aplicadas com sucesso")
	return nil
}

func Close() {
	if DB != nil {
		DB.Close()
	}
}
