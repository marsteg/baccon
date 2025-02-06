package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"time"

	"github.com/go-pg/pg/v10"
	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus"
)

func (c apiCfg) GetPostgres(w http.ResponseWriter, r *http.Request) {
	var connList []Connection
	for _, conn := range c.Connections {
		connList = append(connList, conn)
	}

	jsonData, err := json.Marshal(connList)
	if err != nil {
		http.Error(w, "Error converting data to JSON", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonData)
}

func (c apiCfg) CreatePostgres(w http.ResponseWriter, r *http.Request) {
	var conn Connection
	res, err := httputil.DumpRequest(r, true)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Create PG Incoming Request body: %s\n", string(res))

	if err := json.NewDecoder(r.Body).Decode(&conn); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	conn.ID = uuid.New().String()
	addr := conn.Host + ":" + conn.Port
	conn.PostgresConn = pg.Connect(&pg.Options{
		User:     conn.User,
		Addr:     addr,
		Password: conn.Password,
		Database: conn.Database,
	})
	conn.ProcessedRequests = prometheus.NewCounter(prometheus.CounterOpts{
		Name: fmt.Sprintf("conn_%s_processed_requests", conn.Name),
		Help: "The total number of processed requests of this Connection",
	})
	prometheus.MustRegister(conn.ProcessedRequests)

	conn.LastWriteRequestDuration = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    fmt.Sprintf("conn_%s_last_write_request_duration", conn.Name),
		Help:    "The last write request duration of this Connection",
		Buckets: []float64{0.05, 0.1, 0.2, 0.3, 0.4, 0.5, 0.6, 1, 2, 2.5, 5, 10},
	})
	prometheus.MustRegister(conn.LastWriteRequestDuration)

	conn.LastReadRequestDuration = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    fmt.Sprintf("conn_%s_last_read_request_duration", conn.Name),
		Help:    "The last read request duration of this Connection",
		Buckets: []float64{0.05, 0.1, 0.2, 0.3, 0.4, 0.5, 0.6, 1, 2, 2.5, 5, 10},
	})
	prometheus.MustRegister(conn.LastReadRequestDuration)

	timer := prometheus.NewTimer(conn.LastWriteRequestDuration)

	err = conn.PostgresConn.Ping(r.Context())
	if err != nil {
		http.Error(w, "Cannot connect to DB: "+err.Error(), http.StatusRequestTimeout)
		return
	}
	conn.Status = "connected"

	// create a new table
	_, err = conn.PostgresConn.Exec("CREATE TABLE IF NOT EXISTS " + conn.Name + " (id serial PRIMARY KEY, Tests VARCHAR(50));")
	if err != nil {
		http.Error(w, "Cannot create table: "+err.Error(), http.StatusInternalServerError)
		return
	}
	conn.ProcessedRequests.Inc()
	timer.ObserveDuration()

	jsonData, err := json.Marshal(conn)
	if err != nil {
		http.Error(w, "Error converting data to JSON", http.StatusInternalServerError)
		return
	}
	c.Connections[conn.ID] = conn
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonData)
}

func (c apiCfg) GetPostgresID(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Path[len("/postgres/"):]
	fmt.Printf("GET ID Incoming Request body: %s\n", id)

	conn, exists := c.Connections[id]
	if !exists {
		http.Error(w, "Connection not found", http.StatusNotFound)
		return
	}

	jsonData, err := json.Marshal(conn)
	if err != nil {
		http.Error(w, "Error converting data to JSON", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonData)
}

func (c apiCfg) DeletePostgres(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Path[len("/postgres/"):]
	fmt.Printf("DEL PG Incoming Request body: %s\n", id)

	conn, exists := c.Connections[id]
	if !exists {
		http.Error(w, "Connection not found", http.StatusNotFound)
		return
	}

	_, err := conn.PostgresConn.Exec("DROP TABLE " + conn.Name + ";")
	if err != nil {
		http.Error(w, "Cannot insert data: "+err.Error(), http.StatusInternalServerError)
		return
	}

	err = conn.PostgresConn.Close()
	if err != nil {
		http.Error(w, "Cannot close connection: "+err.Error(), http.StatusInternalServerError)
		return
	}

	conn.ProcessedRequests.Inc()
	delete(c.Connections, id)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNoContent)
}

func (c apiCfg) TestWritePostgres(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Path[len("/postgres/"):]
	id = id[:len(id)-len("/write")]
	fmt.Printf("Test Write  PG Incoming Request body: %s\n", id)

	conn, exists := c.Connections[id]
	if !exists {
		http.Error(w, "Connection not found", http.StatusNotFound)
		return
	}
	// create Table with example data
	timer := prometheus.NewTimer(conn.LastWriteRequestDuration)
	t := time.Now().UTC()
	ts := t.Format("2006-01-02 15:04:05")
	_, err := conn.PostgresConn.Exec("INSERT INTO " + conn.Name + " (Tests) VALUES ('" + ts + "');")
	if err != nil {
		http.Error(w, "Cannot insert data: "+err.Error(), http.StatusInternalServerError)
		return
	}

	conn.ProcessedRequests.Inc()
	timer.ObserveDuration()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
}

type PostgresQueryResponse struct {
	Data         []interface{} `json:"data"`
	RowsReturned int           `json:"rows_returned"`
}

func (c apiCfg) TestQueryPostgres(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Path[len("/postgres/"):]
	id = id[:len(id)-len("/query")]
	fmt.Printf("Test Query PG Incoming Request body: %s\n", id)

	conn, exists := c.Connections[id]
	if !exists {
		http.Error(w, "Connection not found", http.StatusNotFound)
		return
	}
	var response PostgresQueryResponse
	var results []map[string]interface{}
	timer := prometheus.NewTimer(conn.LastReadRequestDuration)
	_, err := conn.PostgresConn.Query(&results, "SELECT Tests FROM "+conn.Name+";")
	if err != nil {
		http.Error(w, "Cannot query data: "+err.Error(), http.StatusInternalServerError)
		return
	}
	response.RowsReturned = len(results)
	response.Data = make([]interface{}, len(results))
	for i, v := range results {
		response.Data[i] = v
	}

	jsonData, err := json.Marshal(response)
	if err != nil {
		http.Error(w, "Error converting data to JSON", http.StatusInternalServerError)
		return
	}
	timer.ObserveDuration()
	conn.ProcessedRequests.Inc()
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonData)
}
