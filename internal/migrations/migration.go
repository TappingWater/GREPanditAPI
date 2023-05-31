package migrations

import (
	"context"
	"log"

	"github.com/jackc/pgx/v4/pgxpool"
)

func Migrate(db *pgxpool.Pool) {
	ctx := context.Background()

	_, err := db.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS words (
			id SERIAL PRIMARY KEY,
			word TEXT UNIQUE,
			meanings TEXT[],
			examples TEXT[]
		);
	`)

	if err != nil {
		log.Fatalf("Could not create words table: %v", err)
	}

	_, err = db.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS verbal_questions (
			id SERIAL PRIMARY KEY,
			competence INT,
			framed_as INT,
			type INT,
			paragraph TEXT,
			question TEXT,
			options TEXT[],
			answer TEXT[],
			word TEXT[],
			explanation TEXT,
			difficulty INT
		);
	`)

	if err != nil {
		log.Fatalf("Could not create verbal_questions table: %v", err)
	}

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
