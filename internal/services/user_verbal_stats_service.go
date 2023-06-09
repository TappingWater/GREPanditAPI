package services

import (
	"context"
	"encoding/json"

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

func (s *UserVerbalStatsService) Create(ctx context.Context, stat *models.UserVerbalStat, wordIDs []int) error {
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
	answersJSON, err := json.Marshal(stat.Answers)
	if err != nil {
		return err
	}
	err = tx.QueryRow(ctx, query, stat.UserToken, stat.QuestionID, stat.Correct, answersJSON, stat.Date).Scan(&stat.ID)
	if err != nil {
		return err
	}
	// Add records to the join table
	for _, wordID := range wordIDs {
		err = s.addToJoinTable(ctx, tx, stat.QuestionID, wordID)
		if err != nil {
			return err
		}
	}
	// Commit the transaction.
	err = tx.Commit(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (s *UserVerbalStatsService) addToJoinTable(ctx context.Context, tx pgx.Tx, questionID, wordID int) error {
	query := `
		INSERT INTO ` + database.UserVerbalStatsJoinTable + ` (` +
		database.UserVerbalStatsJoinVerbalField + `, ` +
		database.UserVerbalStatsJoinWordField + `)
		VALUES ($1, $2)`
	_, err := tx.Exec(ctx, query, questionID, wordID)
	return err
}

func (s *UserVerbalStatsService) GetMarkedWordsByUserToken(ctx context.Context, userToken string) ([]models.UserMarkedWord, error) {
	query := `
		SELECT * FROM ` + database.UserMarkedWordsTable + `
		WHERE ` + database.UserMarkedWordsUserField + ` = $1`
	rows, err := s.DB.Query(ctx, query, userToken)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	markedWords := make([]models.UserMarkedWord, 0)
	for rows.Next() {
		var markedWord models.UserMarkedWord
		err := rows.Scan(&markedWord.ID, &markedWord.UserToken, &markedWord.WordID)
		if err != nil {
			return nil, err
		}
		markedWords = append(markedWords, markedWord)
	}
	return markedWords, nil
}

func (s *UserVerbalStatsService) GetMarkedVerbalQuestionsByUserToken(ctx context.Context, userToken string) ([]models.UserMarkedVerbalQuestion, error) {
	query := `
		SELECT * FROM ` + database.UserMarkedVerbalQuestionsTable + `
		WHERE ` + database.UserMarkedVerbalQuestionsUserField + ` = $1`
	rows, err := s.DB.Query(ctx, query, userToken)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	markedQuestions := make([]models.UserMarkedVerbalQuestion, 0)
	for rows.Next() {
		var markedQuestion models.UserMarkedVerbalQuestion
		err := rows.Scan(&markedQuestion.ID, &markedQuestion.UserToken, &markedQuestion.VerbalQuestionID)
		if err != nil {
			return nil, err
		}
		markedQuestions = append(markedQuestions, markedQuestion)
	}
	return markedQuestions, nil
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
