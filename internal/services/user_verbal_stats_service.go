package services

import (
	"context"
	"time"

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
	query := `
		INSERT INTO ` + database.VerbalStatsTable + ` (` +
		database.VerbalStatsUserField + `, ` +
		database.VerbalStatsQuestionField + `, ` +
		database.VerbalStatsCorrectField + `, ` +
		database.VerbalStatsAnswersField + `, ` +
		database.VerbalStatsDurationField + `, ` +
		database.VerbalStatsDateField + `)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING ` + database.VerbalStatsIDField
	return s.DB.QueryRow(ctx, query, userToken, stat.QuestionID, stat.Correct, stat.Answers, stat.Duration, time.Now()).Scan(&stat.ID)
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
		err := rows.Scan(&verbalStat.ID, &verbalStat.UserToken, &verbalStat.QuestionID, &verbalStat.Correct, &verbalStat.Answers, &verbalStat.Duration, &verbalStat.Date)
		if err != nil {
			return nil, err
		}
		verbalStats = append(verbalStats, verbalStat)
	}
	return verbalStats, nil
}
