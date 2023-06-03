package migrations

import (
	"context"
	"log"

	"github.com/jackc/pgx/v4/pgxpool"
)

func Migrate(db *pgxpool.Pool) {
	ctx := context.Background()

	// Create table for word
	_, err := db.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS words (
			id SERIAL PRIMARY KEY,
			word TEXT UNIQUE,
			meanings JSONB
		);
	`)

	if err != nil {
		log.Fatalf("Could not create words table: %v", err)
	}

	// Create table for verbal questions
	_, err = db.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS verbal_questions (
			id SERIAL PRIMARY KEY,
			competence INT,
			framed_as INT,
			type INT,
			paragraph TEXT,
			question TEXT,
			options JSONB,
			answer TEXT[],
			word TEXT[],
			explanation TEXT,
			difficulty INT,
			wordmap JSONB
		);
	`)

	if err != nil {
		log.Fatalf("Could not create verbal_questions table: %v", err)
	}

	// Create join table for verbal questions and words
	_, err = db.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS verbal_question_words (
			verbal_question_id INT REFERENCES verbal_questions(id),
			word_id INT REFERENCES words(id),
			PRIMARY KEY (verbal_question_id, word_id)
		);
	`)

	if err != nil {
		log.Fatalf("Could not create verbal_question_words table: %v", err)
	}

	// Create needed indexes for querying and improving performance
	_, err = db.Exec(ctx, `
		CREATE INDEX IF NOT EXISTS idx_word ON words(word);
		CREATE INDEX IF NOT EXISTS idx_competence ON verbal_questions(competence);
		CREATE INDEX IF NOT EXISTS idx_framed_as ON verbal_questions(framed_as);
		CREATE INDEX IF NOT EXISTS idx_type ON verbal_questions(type);
	`)

	if err != nil {
		log.Fatalf("Could not create indices: %v", err)
	}
}
