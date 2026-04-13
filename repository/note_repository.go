package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"note-service/models"
)

var ErrNoteNotFound = errors.New("note not found")

type NoteRepository struct {
	db *sql.DB
}

func NewNoteRepository(db *sql.DB) *NoteRepository {
	return &NoteRepository{db: db}
}

func (r *NoteRepository) Create(note models.Note) (models.Note, error) {
	query := `
	INSERT INTO notes (title, content)
	VALUES($1, $2)
	RETURNING id, created_at
	`
	err := r.db.QueryRow(
		query,
		note.Title,
		note.Content,
	).Scan(&note.ID, &note.CreatedAt)
	return note, err
}

func (r *NoteRepository) GetAll() ([]models.Note, error) {
	query := `
	SELECT id, title, content, created_at
	FROM notes
	ORDER BY id
	`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var notes []models.Note
	for rows.Next() {
		var note models.Note
		err := rows.Scan(
			&note.ID,
			&note.Title,
			&note.Content,
			&note.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		notes = append(notes, note)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return notes, nil
}

func (r *NoteRepository) GetByID(id int) (models.Note, error) {
	query := `
	SELECT id, title, content, created_at
	FROM notes
	WHERE id= $1
	`
	var note models.Note
	err := r.db.QueryRow(query, id).Scan(
		&note.ID,
		&note.Title,
		&note.Content,
		&note.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return models.Note{}, ErrNoteNotFound
		}
		return models.Note{}, fmt.Errorf("query failed: %w", err)
	}
	return note, nil
}

func (r *NoteRepository) Delete(id int) error {
	query := `
	DELETE FROM notes
	WHERE id = $1
	`
	result, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("delete failed: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return ErrNoteNotFound
	}

	return nil
}
