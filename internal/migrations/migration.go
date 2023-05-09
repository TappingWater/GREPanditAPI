package migrations

import (
	"context"
	"log"

	"github.com/jackc/pgx/v4/pgxpool"
)

func Migrate(db *pgxpool.Pool) {
	ctx := context.Background()
	_, err := db.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS verbal_questions (
			id SERIAL PRIMARY KEY,
			competence INT,
			framed_as INT,
			type INT,
			paragraph_id INT,
			question TEXT,
			options TEXT[],
			answer TEXT[],
			explanation TEXT,
			difficulty INT
		);
	`)

	if err != nil {
		log.Fatalf("Could not create verbal_questions table: %v", err)
	}

	// Add other table creations or migrations here
}
