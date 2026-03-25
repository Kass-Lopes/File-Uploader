package main

import (
	"log"
	"net/http"
	"os"

	"file-uploader/db"
	"file-uploader/handlers"
)

func main() {
	// Configurações via variáveis de ambiente
	port := getEnv("PORT", "8080")
	dbPath := getEnv("DB_PATH", "./data/files.db")
	uploadDir := getEnv("UPLOAD_DIR", "./uploads")

	// Inicializa o banco de dados
	if err := db.Init(dbPath); err != nil {
		log.Fatalf("Falha ao inicializar banco de dados: %v", err)
	}
	defer db.Close()

	// Garante que o diretório de uploads existe
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		log.Fatalf("Falha ao criar diretório de uploads: %v", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", handlers.HandleIndex)
	mux.HandleFunc("/upload", handlers.HandleUpload)
	mux.HandleFunc("/download/", handlers.HandleDownload)
	mux.HandleFunc("/delete/", handlers.HandleDelete)

	// Arquivos estáticos
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))

	log.Printf("Servidor rodando em http://localhost:%s", port)
	log.Printf("Upload dir: %s | DB: %s", uploadDir, dbPath)

	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatalf("Erro no servidor: %v", err)
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
