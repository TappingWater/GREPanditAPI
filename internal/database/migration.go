package database

import (
	"context"
	"log"

	"github.com/jackc/pgx/v4/pgxpool"
)

func Migrate(db *pgxpool.Pool) {
	ctx := context.Background()

	// Create table for words
	_, err := db.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS `+WordsTable+` (
			`+WordsIDField+` SERIAL PRIMARY KEY,
			`+WordsWordField+` TEXT UNIQUE,
			`+WordsMeaningsField+` JSONB
		);
	`)

	if err != nil {
		log.Fatalf("Could not create "+WordsTable+" table: %v", err)
	}

	// Create table for verbal questions
	_, err = db.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS `+VerbalQuestionsTable+` (
			`+VerbalQuestionsIDField+` SERIAL PRIMARY KEY,
			`+VerbalQuestionsCompetenceField+` INT,
			`+VerbalQuestionsFramedAsField+` INT,
			`+VerbalQuestionsTypeField+` INT,
			`+VerbalQuestionsParagraphField+` TEXT,
			`+VerbalQuestionsQuestionField+` TEXT,
			`+VerbalQuestionsOptionsField+` JSONB,
			`+VerbalQuestionsAnswerField+` TEXT[],
			`+VerbalQuestionsWordField+` TEXT[],
			`+VerbalQuestionsExplanationField+` TEXT,
			`+VerbalQuestionsDifficultyField+` INT,
			`+VerbalQuestionsWordmapField+` JSONB
		);
	`)

	if err != nil {
		log.Fatalf("Could not create "+VerbalQuestionsTable+" table: %v", err)
	}

	// Create join table for verbal questions and words
	_, err = db.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS `+VerbalQuestionWordsJoinTable+` (
			`+VerbalQuestionWordJoinVerbalField+` INT REFERENCES `+VerbalQuestionsTable+`(`+VerbalQuestionsIDField+`),
			`+VerbalQuestionWordJoinWordField+` INT REFERENCES `+WordsTable+`(`+WordsIDField+`),
			PRIMARY KEY (`+VerbalQuestionWordJoinVerbalField+`, `+VerbalQuestionWordJoinWordField+`)
		);
	`)

	if err != nil {
		log.Fatalf("Could not create "+VerbalQuestionWordsJoinTable+" table: %v", err)
	}

	// Create user table
	_, err = db.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS `+UsersTable+` (
			`+UserIDField+` SERIAL PRIMARY KEY,
			`+UserTokenField+` TEXT NOT NULL UNIQUE,
			`+UserEmailField+` TEXT NOT NULL UNIQUE
		);
	`)

	if err != nil {
		log.Fatalf("Could not create "+UsersTable+" table: %v", err)
	}

	// Create verbal stats table
	_, err = db.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS `+VerbalStatsTable+` (
			`+VerbalStatsIDField+` SERIAL PRIMARY KEY,
			`+VerbalStatsUserField+` TEXT NOT NULL REFERENCES `+UsersTable+`(`+UserTokenField+`),
			`+VerbalStatsQuestionField+` INT NOT NULL REFERENCES `+VerbalQuestionsTable+`(`+VerbalQuestionsIDField+`),
			`+VerbalStatsCorrectField+` BOOLEAN,
			`+VerbalStatsAnswersField+` TEXT[],
			`+VerbalStatsDateField+` TIMESTAMP
		);
	`)

	if err != nil {
		log.Fatalf("Could not create "+VerbalStatsTable+" table: %v", err)
	}

	// Create user verbal stats join table
	_, err = db.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS `+UserVerbalStatsJoinTable+` (
			`+UserVerbalStatsJoinVerbalField+` INT REFERENCES `+VerbalQuestionsTable+`(`+VerbalQuestionsIDField+`),
			`+UserVerbalStatsJoinUserField+` TEXT REFERENCES `+UsersTable+`(`+UserTokenField+`),
			PRIMARY KEY (`+UserVerbalStatsJoinVerbalField+`, `+UserVerbalStatsJoinUserField+`)
		);
	`)

	if err != nil {
		log.Fatalf("Could not create "+UserVerbalStatsJoinTable+" table: %v", err)
	}

	// Create user marked words table
	_, err = db.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS `+UserMarkedWordsTable+` (
			`+UserMarkedWordsIDField+` SERIAL PRIMARY KEY,
			`+UserMarkedWordsUserField+` TEXT NOT NULL REFERENCES `+UsersTable+`(`+UserTokenField+`),
			`+UserMarkedWordsWordField+` INT NOT NULL REFERENCES `+WordsTable+`(`+WordsIDField+`),
			UNIQUE (`+UserMarkedWordsUserField+`, `+UserMarkedWordsWordField+`)
		);
	`)

	if err != nil {
		log.Fatalf("Could not create "+UserMarkedWordsTable+" table: %v", err)
	}

	// Create user marked verbal questions table
	_, err = db.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS `+UserMarkedVerbalQuestionsTable+` (
			`+UserMarkedVerbalQuestionsIDField+` SERIAL PRIMARY KEY,
			`+UserMarkedVerbalQuestionsUserField+` TEXT NOT NULL REFERENCES `+UsersTable+`(`+UserTokenField+`),
			`+UserMarkedVerbalQuestionsQuestionField+` INT,
			UNIQUE (`+UserMarkedVerbalQuestionsUserField+`, `+UserMarkedVerbalQuestionsQuestionField+`)
		);
	`)

	if err != nil {
		log.Fatalf("Could not create "+UserMarkedVerbalQuestionsTable+" table: %v", err)
	}

	// Create needed indexes for querying and improving performance
	_, err = db.Exec(ctx, `
		CREATE INDEX IF NOT EXISTS idx_word ON `+WordsTable+`(`+WordsWordField+`);
		CREATE INDEX IF NOT EXISTS idx_competence ON `+VerbalQuestionsTable+`(`+VerbalQuestionsCompetenceField+`);
		CREATE INDEX IF NOT EXISTS idx_framed_as ON `+VerbalQuestionsTable+`(`+VerbalQuestionsFramedAsField+`);
		CREATE INDEX IF NOT EXISTS idx_type ON `+VerbalQuestionsTable+`(`+VerbalQuestionsTypeField+`);
		CREATE INDEX IF NOT EXISTS idx_user_token_users ON `+UsersTable+`(`+UserTokenField+`);
		CREATE INDEX IF NOT EXISTS idx_user_token_user_verbal_stats ON `+UserVerbalStatsJoinTable+`(`+UserVerbalStatsJoinUserField+`);
		CREATE INDEX IF NOT EXISTS idx_user_token_user_marked_words ON `+UserMarkedWordsTable+`(`+UserMarkedWordsUserField+`);
		CREATE INDEX IF NOT EXISTS idx_user_token_user_marked_verbal_questions ON `+UserMarkedVerbalQuestionsTable+`(`+UserMarkedVerbalQuestionsUserField+`);
	`)

	if err != nil {
		log.Fatalf("Could not create indices: %v", err)
	}
}