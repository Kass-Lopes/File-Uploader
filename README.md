# File-Uploader — em Go + SQLite

## Instala as dependências
	go mod tidy
	go get github.com/mattn/go-sqlite3

## Endpoints

| Método | Rota            | Descrição                    |
|--------|-----------------|------------------------------|
| GET    | `/`             | Página principal             |
| POST   | `/upload`       | Upload de arquivo            |
| GET    | `/download/{id}`| Download do arquivo por ID   |
| POST   | `/delete/{id}`  | Deletar arquivo por ID       |

