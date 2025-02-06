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

type apiCfg struct {
	Connections map[string]Connection `json:"connections"`
}

func initMap() apiCfg {
	Connections := make(map[string]Connection)
	apiCfg := apiCfg{
		Connections: Connections,
	}
	return apiCfg
}

func main() {
	cfg := initMap()
	http.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "You have reached the baccon!")
	})
	http.HandleFunc("GET /postgres/", cfg.GetPostgres)
	http.HandleFunc("GET /postgres/{id}", cfg.GetPostgresID)
	http.HandleFunc("POST /postgres/", cfg.CreatePostgres)
	http.HandleFunc("DELETE /postgres/{id}", cfg.DeletePostgres)
	http.HandleFunc("GET /postgres/{id}/write", cfg.TestWritePostgres)
	http.HandleFunc("GET /postgres/{id}/query", cfg.TestQueryPostgres)

	// Prometheus Handler
	http.Handle("GET /metrics", promhttp.Handler())

	fmt.Println("Server starting on port 8080...")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Printf("Failed to start server: %v\n", err)
	}
}
