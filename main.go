package main

import (
	"fmt"
	"net/http"

	"github.com/go-pg/pg/v10"
	_ "github.com/lib/pq"
)

type Connection struct {
	ID           string `json:"id,omitempty"`
	Host         string `json:"host"`
	Port         string `json:"port"`
	Name         string `json:"name"`
	User         string `json:"user"`
	Password     string `json:"password"`
	Status       string `json:"status,omitempty"`
	Database     string `json:"database"`
	PostgresConn *pg.DB `json:"-"` // ignore this field in JSON
}

var connections = make(map[string]Connection)

func main() {
	http.HandleFunc("GET /postgres/", GetPostgres)
	http.HandleFunc("GET /postgres/{id}", GetPostgresID)
	http.HandleFunc("POST /postgres/", CreatePostgres)
	http.HandleFunc("DELETE /postgres/{id}", DeletePostgres)
	http.HandleFunc("GET /postgres/{id}/write", TestWritePostgres)
	http.HandleFunc("GET /postgres/{id}/query", TestQueryPostgres)

	fmt.Println("Server starting on port 8080...")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Printf("Failed to start server: %v\n", err)
	}
}
