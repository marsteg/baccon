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
)

func GetPostgres(w http.ResponseWriter, r *http.Request) {
	var connList []Connection
	for _, conn := range connections {
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

func CreatePostgres(w http.ResponseWriter, r *http.Request) {
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

	err = conn.PostgresConn.Ping(r.Context())
	if err != nil {
		http.Error(w, "Cannot connect to DB: "+err.Error(), http.StatusRequestTimeout)
		return
	}
	conn.Status = "connected"

	connections[conn.ID] = conn

	// create a new table
	_, err = conn.PostgresConn.Exec("CREATE TABLE IF NOT EXISTS baccon (id serial PRIMARY KEY, Tests VARCHAR(50));")
	if err != nil {
		http.Error(w, "Cannot create table: "+err.Error(), http.StatusInternalServerError)
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

func GetPostgresID(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Path[len("/postgres/"):]
	fmt.Printf("GET ID Incoming Request body: %s\n", id)

	conn, exists := connections[id]
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

func DeletePostgres(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Path[len("/postgres/"):]
	fmt.Printf("DEL PG Incoming Request body: %s\n", id)

	conn, exists := connections[id]
	if !exists {
		http.Error(w, "Connection not found", http.StatusNotFound)
		return
	}

	err := conn.PostgresConn.Close()
	if err != nil {
		http.Error(w, "Cannot close connection: "+err.Error(), http.StatusInternalServerError)
		return
	}

	delete(connections, id)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNoContent)
}

func TestWritePostgres(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Path[len("/postgres/"):]
	id = id[:len(id)-len("/write")]
	fmt.Printf("Test Write  PG Incoming Request body: %s\n", id)

	conn, exists := connections[id]
	if !exists {
		http.Error(w, "Connection not found", http.StatusNotFound)
		return
	}
	// create Table with example data
	t := time.Now().UTC()
	ts := t.Format("2006-01-02 15:04:05")
	_, err := conn.PostgresConn.Exec("INSERT INTO baccon (Tests) VALUES ('" + ts + "');")
	if err != nil {
		http.Error(w, "Cannot insert data: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
}

type PostgresQueryResponse struct {
	ResponseRow  map[int]string `json:"response_row"`
	RowsReturned int            `json:"rows_returned"`
}

func TestQueryPostgres(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Path[len("/postgres/"):]
	id = id[:len(id)-len("/query")]
	fmt.Printf("Test Query PG Incoming Request body: %s\n", id)

	conn, exists := connections[id]
	if !exists {
		http.Error(w, "Connection not found", http.StatusNotFound)
		return
	}
	var response PostgresQueryResponse
	var name string
	result, err := conn.PostgresConn.Query(pg.Scan(&name), "SELECT Tests FROM baccon;")
	if err != nil {
		http.Error(w, "Cannot query data: "+err.Error(), http.StatusInternalServerError)
		return
	}
	response.RowsReturned = result.RowsReturned()
	// TODO: find a way to add the actual returned data to the response

	jsonData, err := json.Marshal(response)
	if err != nil {
		http.Error(w, "Error converting data to JSON", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonData)
}
