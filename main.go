package main

import (
	"database/sql"
	"log"
	"net/http"
	"note-service/handlers"
	"note-service/repository"

	"github.com/go-chi/chi/v5"
	_ "github.com/lib/pq"
)

func main() {
	db, err := sql.Open("postgres", "postgres://postgres:12345@localhost:5432/notesdb?sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
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

	log.Println("server started on :8000")
	log.Fatal(http.ListenAndServe(":8000", r))
}
