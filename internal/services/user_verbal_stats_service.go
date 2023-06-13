package services

import (
	"context"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"grepandit.com/api/internal/database"
	"grepandit.com/api/internal/models"
)

type UserVerbalStatsService struct {
	DB *pgxpool.Pool
}

func NewUserVerbalStatsService(db *pgxpool.Pool) *UserVerbalStatsService {
	return &UserVerbalStatsService{DB: db}
}

func (s *UserVerbalStatsService) Create(ctx context.Context, stat *models.UserVerbalStat, userToken string) error {
	// Begin a transaction.
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return err
	}
	// Rollback the transaction in case of error. This is a no-op if the transaction has been committed.
	defer tx.Rollback(ctx)
	query := `
		INSERT INTO ` + database.VerbalStatsTable + ` (` +
		database.VerbalStatsUserField + `, ` +
		database.VerbalStatsQuestionField + `, ` +
		database.VerbalStatsCorrectField + `, ` +
		database.VerbalStatsAnswersField + `, ` +
		database.VerbalStatsDateField + `)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING ` + database.VerbalStatsIDField

	err = tx.QueryRow(ctx, query, userToken, stat.QuestionID, stat.Correct, stat.Answers, time.Now()).Scan(&stat.ID)
	if err != nil {
		return err
	}
	// Add records to the join table
	err = s.addToJoinTable(ctx, tx, stat.QuestionID, userToken)
	if err != nil {
		return err
	}
	// Commit the transaction.
	err = tx.Commit(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (s *UserVerbalStatsService) addToJoinTable(ctx context.Context, tx pgx.Tx, questionID int, userToken string) error {
	query := `
		INSERT INTO ` + database.UserVerbalStatsJoinTable + ` (` +
		database.UserVerbalStatsJoinVerbalField + `, ` +
		database.UserVerbalStatsJoinUserField + `)
		VALUES ($1, $2)`
	_, err := tx.Exec(ctx, query, questionID, userToken)
	return err
}

func (s *UserVerbalStatsService) GetVerbalStatsByUserToken(ctx context.Context, userToken string) ([]models.UserVerbalStat, error) {
	query := `
		SELECT * FROM ` + database.VerbalStatsTable + `
		WHERE ` + database.VerbalStatsUserField + ` = $1`
	rows, err := s.DB.Query(ctx, query, userToken)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	verbalStats := make([]models.UserVerbalStat, 0)
	for rows.Next() {
		var verbalStat models.UserVerbalStat
		err := rows.Scan(&verbalStat.ID, &verbalStat.UserToken, &verbalStat.QuestionID, &verbalStat.Correct, &verbalStat.Answers, &verbalStat.Date)
		if err != nil {
			return nil, err
		}
		verbalStats = append(verbalStats, verbalStat)
	}
	return verbalStats, nil
}
