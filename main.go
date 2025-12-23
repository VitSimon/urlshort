package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "modernc.org/sqlite"
)

const (
	dbPath  = "shortener.db"
	addr    = ":8080"
	timeout = 5 * time.Second
)

var db *sql.DB

// Initialize DB
func initDB() error {
	var err error
	db, err = sql.Open("sqlite", dbPath+"?_journal_mode=WAL")
	if err != nil {
		return err
	}
	db.SetMaxOpenConns(1)
	_, err = db.Exec(`
    CREATE TABLE IF NOT EXISTS uris (
      code TEXT PRIMARY KEY,
      original TEXT NOT NULL
    )
  `)
	return err
}

// Handler : create short URL link
func createHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	original := r.FormValue("u")
	if original == "" {
		http.Error(w, "Missing url", http.StatusBadRequest)
		return
	}

	code := fmt.Sprintf("%x", time.Now().UnixNano())[:6] // simple code
	_, err := db.Exec("INSERT INTO uris(code, original) VALUES(?, ?)", code, original)
	if err != nil {
		log.Printf("DB insert error: %v", err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	w.Write([]byte(fmt.Sprintf("Short URL code: %s\n", code)))
}

// Handler : Redirect
func redirectHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	code := r.URL.Path[1:] // remove leading "/"
	if code == "" {
		http.Error(w, "Missing code", http.StatusBadRequest)
		return
	}

	var original string
	err := db.QueryRow("SELECT original FROM uris WHERE code = ?", code).Scan(&original)
	if err != nil {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	http.Redirect(w, r, original, http.StatusFound)
}

func main() {
	// Initialize DB
	if err := initDB(); err != nil {
		log.Fatalf("DB init failed: %v", err)
	}
	defer db.Close()

	mux := http.NewServeMux()
	mux.HandleFunc("/c", createHandler)
	mux.HandleFunc("/", redirectHandler)

	server := &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  timeout,
		WriteTimeout: timeout,
		IdleTimeout:  60 * time.Second,
	}

	// Shutdown and program end handling
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Printf("Server listening on %s", addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("ListenAndServe: %v", err)
		}
	}()

	<-stop
	log.Println("Shutting down server...")

	if err := server.Close(); err != nil {
		log.Fatalf("Server Close failed: %v", err)
	}
	log.Println("Server stopped.")
}
