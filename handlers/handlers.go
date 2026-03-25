package handlers

import (
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"file-uploader/models"
)

const (
	MaxUploadSize = 50 << 20 // aceita max 50 MB
	UploadDir     = "./uploads"
)

var tmpl *template.Template

func init() {
	funcMap := template.FuncMap{
		"humanSize": humanSize,
		"formatDate": func(t time.Time) string {
			return t.Format("02/01/2006 15:04")
		},
		"mimeIcon":  mimeIcon,
		"mimeColor": mimeColor,
	}
	tmpl = template.Must(template.New("").Funcs(funcMap).ParseGlob("templates/*.html"))
}

func HandleIndex(w http.ResponseWriter, r *http.Request) {
	files, err := models.GetAllFiles()
	if err != nil {
		http.Error(w, "Erro ao buscar arquivos", http.StatusInternalServerError)
		log.Println("GetAllFiles:", err)
		return
	}

	stats, err := models.GetStats()
	if err != nil {
		stats = &models.Stats{}
	}

	data := map[string]interface{}{
		"Files": files,
		"Stats": stats,
		"Flash": r.URL.Query().Get("msg"),
		"Error": r.URL.Query().Get("err"),
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.ExecuteTemplate(w, "index.html", data); err != nil {
		log.Println("Template error:", err)
	}
}

func HandleUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, MaxUploadSize+1024)
	if err := r.ParseMultipartForm(MaxUploadSize); err != nil {
		http.Redirect(w, r, "/?err=Arquivo+muito+grande+(máximo+50MB)", http.StatusSeeOther)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Redirect(w, r, "/?err=Nenhum+arquivo+selecionado", http.StatusSeeOther)
		return
	}
	defer file.Close()

	description := strings.TrimSpace(r.FormValue("description"))

	buf := make([]byte, 512)
	n, _ := file.Read(buf)
	mimeType := http.DetectContentType(buf[:n])
	file.Seek(0, io.SeekStart)

	ext := filepath.Ext(header.Filename)
	storedName := fmt.Sprintf("%d_%s%s",
		time.Now().UnixNano(),
		sanitizeName(strings.TrimSuffix(header.Filename, ext)),
		ext,
	)

	destPath := filepath.Join(UploadDir, storedName)
	dest, err := os.Create(destPath)
	if err != nil {
		log.Println("Criar arquivo:", err)
		http.Redirect(w, r, "/?err=Erro+ao+salvar+arquivo", http.StatusSeeOther)
		return
	}
	defer dest.Close()

	size, err := io.Copy(dest, file)
	if err != nil {
		os.Remove(destPath)
		http.Redirect(w, r, "/?err=Erro+ao+gravar+arquivo", http.StatusSeeOther)
		return
	}

	f := &models.File{
		OriginalName: header.Filename,
		StoredName:   storedName,
		Size:         size,
		MimeType:     mimeType,
		Description:  description,
	}

	id, err := models.CreateFile(f)
	if err != nil {
		os.Remove(destPath)
		log.Println("CreateFile:", err)
		http.Redirect(w, r, "/?err=Erro+ao+registrar+no+banco", http.StatusSeeOther)
		return
	}

	log.Printf("Upload OK: id=%d nome=%q size=%d mime=%s", id, header.Filename, size, mimeType)
	http.Redirect(w, r, "/?msg=Arquivo+enviado+com+sucesso!", http.StatusSeeOther)
}

func HandleDownload(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDFromPath(r.URL.Path, "/download/")
	if err != nil {
		http.NotFound(w, r)
		return
	}

	f, err := models.GetFileByID(id)
	if err != nil || f == nil {
		http.NotFound(w, r)
		return
	}

	filePath := filepath.Join(UploadDir, f.StoredName)
	fh, err := os.Open(filePath)
	if err != nil {
		http.Error(w, "Arquivo não encontrado no disco", http.StatusNotFound)
		return
	}
	defer fh.Close()

	models.IncrementDownload(id)

	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, f.OriginalName))
	w.Header().Set("Content-Type", f.MimeType)
	w.Header().Set("Content-Length", strconv.FormatInt(f.Size, 10))
	io.Copy(w, fh)
}

func HandleDelete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	id, err := parseIDFromPath(r.URL.Path, "/delete/")
	if err != nil {
		http.NotFound(w, r)
		return
	}

	f, err := models.GetFileByID(id)
	if err != nil || f == nil {
		http.NotFound(w, r)
		return
	}

	os.Remove(filepath.Join(UploadDir, f.StoredName))

	if err := models.DeleteFile(id); err != nil {
		log.Println("DeleteFile:", err)
		http.Redirect(w, r, "/?err=Erro+ao+deletar", http.StatusSeeOther)
		return
	}

	log.Printf("Delete OK: id=%d nome=%q", id, f.OriginalName)
	http.Redirect(w, r, "/?msg=Arquivo+removido+com+sucesso", http.StatusSeeOther)
}

func parseIDFromPath(path, prefix string) (int64, error) {
	s := strings.TrimPrefix(path, prefix)
	s = strings.Trim(s, "/")
	return strconv.ParseInt(s, 10, 64)
}

func sanitizeName(name string) string {
	replacer := strings.NewReplacer(" ", "_", "/", "_", "\\", "_")
	s := replacer.Replace(name)
	if len(s) > 40 {
		s = s[:40]
	}
	return s
}

func humanSize(b int64) string {
	const unit = 1024
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

func mimeIcon(mime string) string {
	switch {
	case strings.HasPrefix(mime, "image/"):
		return "image"
	case strings.HasPrefix(mime, "video/"):
		return "video"
	case strings.HasPrefix(mime, "audio/"):
		return "audio"
	case mime == "application/pdf":
		return "pdf"
	case strings.Contains(mime, "zip") || strings.Contains(mime, "tar") || strings.Contains(mime, "gzip"):
		return "archive"
	case strings.Contains(mime, "text"):
		return "text"
	default:
		return "file"
	}
}

func mimeColor(mime string) string {
	switch {
	case strings.HasPrefix(mime, "image/"):
		return "#7c3aed"
	case strings.HasPrefix(mime, "video/"):
		return "#db2777"
	case strings.HasPrefix(mime, "audio/"):
		return "#059669"
	case mime == "application/pdf":
		return "#dc2626"
	case strings.Contains(mime, "zip") || strings.Contains(mime, "tar"):
		return "#d97706"
	default:
		return "#2563eb"
	}
}
