package main

import (
	"fmt"
	"net/http"

	"github.com/go-pg/pg/v10"
	_ "github.com/lib/pq"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Connection struct {
	ID                       string               `json:"id,omitempty"`
	Host                     string               `json:"host"`
	Port                     string               `json:"port"`
	Name                     string               `json:"name"`
	User                     string               `json:"user"`
	Password                 string               `json:"password"`
	Status                   string               `json:"status,omitempty"`
	Database                 string               `json:"database"`
	PostgresConn             *pg.DB               `json:"-"` // ignore this field in JSON
	ProcessedRequests        prometheus.Counter   `json:"-"` // ignore this field in JSON
	LastWriteRequestDuration prometheus.Histogram `json:"last_write_request,omitempty"`
	LastReadRequestDuration  prometheus.Histogram `json:"last_read_request,omitempty"`
}

var Connections = make(map[string]Connection)

func main() {
	http.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "You have reached the baccon!")
	})
	http.HandleFunc("GET /postgres/", GetPostgres)
	http.HandleFunc("GET /postgres/{id}", GetPostgresID)
	http.HandleFunc("POST /postgres/", CreatePostgres)
	http.HandleFunc("DELETE /postgres/{id}", DeletePostgres)
	http.HandleFunc("GET /postgres/{id}/write", TestWritePostgres)
	http.HandleFunc("GET /postgres/{id}/query", TestQueryPostgres)

	// Prometheus Handler
	http.Handle("GET /metrics", promhttp.Handler())

	fmt.Println("Server starting on port 8080...")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Printf("Failed to start server: %v\n", err)
	}
}
