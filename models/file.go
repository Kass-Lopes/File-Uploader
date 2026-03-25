package models

import (
	"database/sql"
	"fmt"
	"time"

	"file-uploader/db"
)

type File struct {
	ID            int64     `json:"id"`
	OriginalName  string    `json:"original_name"`
	StoredName    string    `json:"stored_name"`
	Size          int64     `json:"size"`
	MimeType      string    `json:"mime_type"`
	Description   string    `json:"description"`
	DownloadCount int64     `json:"download_count"`
	CreatedAt     time.Time `json:"created_at"`
}

// HumanSize retorna o tamanho do arquivo num formato legível
func (f *File) HumanSize() string {
	const unit = 1024
	b := f.Size
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}

func CreateFile(f *File) (int64, error) {
	query := `
		INSERT INTO files (original_name, stored_name, size, mime_type, description)
		VALUES (?, ?, ?, ?, ?)
	`
	result, err := db.DB.Exec(query, f.OriginalName, f.StoredName, f.Size, f.MimeType, f.Description)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func GetAllFiles() ([]File, error) {
	query := `
		SELECT id, original_name, stored_name, size, mime_type, description, download_count, created_at
		FROM files
		ORDER BY created_at DESC
	`
	rows, err := db.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var files []File
	for rows.Next() {
		var f File
		err := rows.Scan(&f.ID, &f.OriginalName, &f.StoredName, &f.Size, &f.MimeType,
			&f.Description, &f.DownloadCount, &f.CreatedAt)
		if err != nil {
			return nil, err
		}
		files = append(files, f)
	}
	return files, rows.Err()
}

func GetFileByID(id int64) (*File, error) {
	query := `
		SELECT id, original_name, stored_name, size, mime_type, description, download_count, created_at
		FROM files WHERE id = ?
	`
	row := db.DB.QueryRow(query, id)
	var f File
	err := row.Scan(&f.ID, &f.OriginalName, &f.StoredName, &f.Size, &f.MimeType,
		&f.Description, &f.DownloadCount, &f.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &f, nil
}

func IncrementDownload(id int64) error {
	_, err := db.DB.Exec(`UPDATE files SET download_count = download_count + 1 WHERE id = ?`, id)
	return err
}

func DeleteFile(id int64) error {
	_, err := db.DB.Exec(`DELETE FROM files WHERE id = ?`, id)
	return err
}

// Stats contém estatísticas gerais do sistema
type Stats struct {
	TotalFiles     int64
	TotalSize      int64
	TotalDownloads int64
}

// GetStats retorna estatísticas gerais do sistema
func GetStats() (*Stats, error) {
	var s Stats
	err := db.DB.QueryRow(`
		SELECT COUNT(*), COALESCE(SUM(size), 0), COALESCE(SUM(download_count), 0)
		FROM files
	`).Scan(&s.TotalFiles, &s.TotalSize, &s.TotalDownloads)
	return &s, err
}
