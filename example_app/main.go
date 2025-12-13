package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	_ "github.com/mattn/go-sqlite3"
	_ "github.com/sijms/go-ora/v2"
)

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func main() {
	// Oracle connection info from requirement
	user := flag.String("user", getEnv("ORA_USER", "LEARN1"), "Oracle username")
	pass := flag.String("pass", getEnv("ORA_PASS", "Welcome"), "Oracle password")
	host := flag.String("host", getEnv("ORA_HOST", "localhost"), "Oracle host")
	port := flag.String("port", getEnv("ORA_PORT", "1521"), "Oracle port")
	service := flag.String("service", getEnv("ORA_SERVICE", "XE"), "Oracle service name")

	// Flags for testing/flexibility
	driver := flag.String("driver", "oracle", "Database driver (oracle, sqlite3)")
	dsnOverride := flag.String("dsn", "", "DSN override (for sqlite or custom)")
	mockService := flag.String("mock-service", "", "URL of the mock service")

	flag.Parse()

	var dsn string
	if *dsnOverride != "" {
		dsn = *dsnOverride
	} else {
		// Construct Oracle DSN (go-ora URL format)
		dsn = fmt.Sprintf("oracle://%s:%s@%s:%s/%s", *user, *pass, *host, *port, *service)
	}

	log.Printf("Connecting to DB: %s (%s)", *driver, dsn)
	db, err := sql.Open(*driver, dsn)
	if err != nil {
		log.Fatalf("Failed to open DB: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping DB: %v", err)
	}

	// API: Update Data
	// GET /update?id=1&status=new_status
	http.HandleFunc("/update", func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Query().Get("id")
		status := r.URL.Query().Get("status")

		if id == "" || status == "" {
			http.Error(w, "Missing id or status", http.StatusBadRequest)
			return
		}

		var query string
		if *driver == "sqlite3" {
			query = "UPDATE users SET status = ? WHERE id = ?"
		} else {
			query = "UPDATE users SET status = :1 WHERE id = :2"
		}

		_, err := db.Exec(query, status, id)
		if err != nil {
			log.Printf("Update error: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"result": "ok"}`))
	})

	// API: Update Data via JSON body and custom header (POST)
	// POST /update-json
	// Header: X-Request-ID required
	// Body: {"id": "1", "status": "updated"}
	http.HandleFunc("/update-json", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		reqID := r.Header.Get("X-Request-ID")
		if reqID == "" {
			http.Error(w, "missing X-Request-ID", http.StatusBadRequest)
			return
		}

		var payload struct {
			ID     string `json:"id"`
			Status string `json:"status"`
		}

		dec := json.NewDecoder(r.Body)
		if err := dec.Decode(&payload); err != nil {
			http.Error(w, "invalid json", http.StatusBadRequest)
			return
		}

		if payload.ID == "" || payload.Status == "" {
			http.Error(w, "missing id or status", http.StatusBadRequest)
			return
		}

		var query string
		if *driver == "sqlite3" {
			query = "UPDATE users SET status = ? WHERE id = ?"
		} else {
			query = "UPDATE users SET status = :1 WHERE id = :2"
		}

		_, err := db.Exec(query, payload.Status, payload.ID)
		if err != nil {
			log.Printf("Update-json error: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		respBody, _ := json.Marshal(map[string]string{
			"id":         payload.ID,
			"status":     payload.Status,
			"request_id": reqID,
		})
		w.WriteHeader(http.StatusOK)
		w.Write(respBody)
	})

	// API: Read Data
	// GET /read?id=1
	// Requirement: "read data ... and expected something, if not return error"
	// Logic: If status is "error", return 500. Else return 200 with JSON.
	http.HandleFunc("/read", func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Query().Get("id")
		if id == "" {
			http.Error(w, "Missing id", http.StatusBadRequest)
			return
		}

		var query string
		if *driver == "sqlite3" {
			query = "SELECT status FROM users WHERE id = ?"
		} else {
			query = "SELECT status FROM users WHERE id = :1"
		}

		var status string
		err := db.QueryRow(query, id).Scan(&status)
		if err != nil {
			if err == sql.ErrNoRows {
				http.Error(w, "Not found", http.StatusNotFound)
			} else {
				log.Printf("Read error: %v", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}

		// Simulate "Expected Something" logic
		if status == "bad" || status == "error" {
			http.Error(w, "Data in invalid state", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"id": "%s", "status": "%s"}`, id, status)
	})

	// API: Call Mock
	// GET /call-mock
	http.HandleFunc("/call-mock", func(w http.ResponseWriter, r *http.Request) {
		if *mockService == "" {
			http.Error(w, "Mock service URL not configured", http.StatusInternalServerError)
			return
		}

		mockURL := *mockService + "/mock-test"
		userType := r.URL.Query().Get("user_type")
		if userType != "" {
			mockURL += "?user_type=" + userType
		}

		log.Printf("Calling mock: %s", mockURL)
		resp, err := http.Get(mockURL)
		if err != nil {
			log.Printf("Call mock error: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Proxy headers
		for k, v := range resp.Header {
			for _, vv := range v {
				w.Header().Add(k, vv)
			}
		}
		w.WriteHeader(resp.StatusCode)
		w.Write(body)
	})

	log.Println("Server listening on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
