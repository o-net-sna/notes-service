package repository

import "database/sql"

func RunMigrations(db *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS notes (
		id SERIAL PRIMARY KEY,
		title TEXT NOT NULL,
		content TEXT,
		created_at TIMESTAMP DEFAULT NOW()
	);
	`
	_, err := db.Exec(query)
	return err
}
