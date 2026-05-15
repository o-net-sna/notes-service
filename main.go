package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"note-service/handlers"
	"note-service/repository"
	"os"
	"time"

	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promhttp"
	_ "github.com/lib/pq"
)


var (
    httpRequestsTotal = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "http_requests_total",
            Help: "Total number of HTTP requests",
        },
        []string{"method", "endpoint", "status_code"},
    )
    httpRequestDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "http_request_duration_seconds",
            Help:    "HTTP request duration in seconds",
            Buckets: prometheus.DefBuckets,
        },
        []string{"method", "endpoint"},
    )
    httpRequestsInFlight = prometheus.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: "http_requests_in_flight",
            Help: "Current number of HTTP requests in flight",
        },
        []string{"method", "endpoint"},
    )
)

func init() {
    prometheus.MustRegister(httpRequestsTotal)
    prometheus.MustRegister(httpRequestDuration)
    prometheus.MustRegister(httpRequestsInFlight)
}

func metricsMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        endpoint := r.URL.Path
        if endpoint == "/" {
            endpoint = "/whoami"
        }
        
        method := r.Method
        
        // Track requests in flight
        httpRequestsInFlight.WithLabelValues(method, endpoint).Inc()
        defer httpRequestsInFlight.WithLabelValues(method, endpoint).Dec()
        
        start := time.Now()
        rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
        
        next.ServeHTTP(rw, r)
        
        duration := time.Since(start).Seconds()
        
        // Record metrics
        httpRequestDuration.WithLabelValues(method, endpoint).Observe(duration)
        httpRequestsTotal.WithLabelValues(method, endpoint, fmt.Sprintf("%d", rw.statusCode)).Inc()
    })
}

type responseWriter struct {
    http.ResponseWriter
    statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
    rw.statusCode = code
    rw.ResponseWriter.WriteHeader(code)
}

func main() {
	databaseURL := getenv("DATABASE_URL", "postgres://postgres:12345@localhost:5432/notesdb?sslmode=disable")
	serverAddr := getenv("SERVER_ADDR", ":8000")

	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if err := waitForDatabase(db, 30*time.Second); err != nil {
		log.Fatal(err)
	}

	if err := repository.RunMigrations(db); err != nil {
		log.Fatal(err)
	}

	repo := repository.NewNoteRepository(db)
	noteHandler := handlers.NewNoteHandler(repo)

	r := chi.NewRouter()

	r.Get("/notes", noteHandler.GetNotes)
	r.Post("/notes", noteHandler.CreateNote)
	r.Get("/notes/{id}", noteHandler.GetNoteByID)
	r.Delete("/notes/{id}", noteHandler.DeleteNote)
	r.Get("/whoami", func(w http.ResponseWriter, r *http.Request) {
		hostname, err := os.Hostname()
		if err != nil {
			hostname = "unknown"
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{
			"service":  "notes-service",
			"hostname": hostname,
		})
	})
	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), time.Second)
		defer cancel()

		if err := db.PingContext(ctx); err != nil {
			http.Error(w, "database unavailable", http.StatusServiceUnavailable)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	})

	r.Handle("/metrics", promhttp.Handler())
	rWithMetrics := metricsMiddleware(r)

	log.Printf("server started on %s", serverAddr)
	log.Fatal(http.ListenAndServe(serverAddr, rWithMetrics))
}

func waitForDatabase(db *sql.DB, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)

	for {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		err := db.PingContext(ctx)
		cancel()

		if err == nil {
			return nil
		}

		if time.Now().After(deadline) {
			return err
		}

		log.Printf("database is not ready yet: %v", err)
		time.Sleep(time.Second)
	}
}

func getenv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
