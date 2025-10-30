package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/lib/pq"
)

func greet(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello World! %s", time.Now())
}
func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func main() {

	dbUser := getenv("DB_USER", "postgres")
	dbPass := getenv("DB_PASS", "postgres")
	dbHost := getenv("DB_HOST", "localhost")
	dbName := getenv("DB_NAME", "postgres")

	connectiondb := fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable", dbUser, dbPass, dbHost, dbName)
	db, err := sql.Open("postgres", connectiondb)
	if err != nil {
		log.Fatal("Error connecting to the database: ", err)
	}
	defer db.Close()

	newMux := http.NewServeMux()
	newMux.HandleFunc("/", greet)
	newMux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		if err := dbPing(db); err != nil {
			http.Error(w, "db not ready: "+err.Error(), http.StatusServiceUnavailable)
			return
		}
		fmt.Fprintln(w, "OK")
	})
	log.Println("listening on :8080")
	if err := http.ListenAndServe(":8080", logging(newMux)); err != nil {
		log.Fatal(err)
	}
}

func dbPing(db *sql.DB) error {
	var err error
	for i := 0; i < 3; i++ {
		if err = db.Ping(); err == nil {
			return nil
		}
		time.Sleep(time.Second)
	}
	return err
}

func logging(mux http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		mux.ServeHTTP(w, r)
		fmt.Printf("Request %s %s took %v\n", r.Method, r.URL.Path, time.Since(start))
	})
}
