package services

import (
	"context"
	"encoding/json"
	"time"

	"github.com/Masterminds/squirrel"
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

func (s *UserVerbalStatsService) GetVocabularyByQuestionIDs(ctx context.Context, ids []int) (map[int][]models.Word, error) {
	query := squirrel.Select(
		"vqw."+database.VerbalQuestionWordJoinVerbalField,
		"w."+database.WordsIDField,
		"w."+database.WordsWordField,
		"w."+database.WordsMeaningsField,
		"w."+database.WordsExamplesField,
	).
		From(database.WordsTable + " AS w").
		Join(database.VerbalQuestionWordsJoinTable + " AS vqw ON w." + database.WordsIDField + " = vqw." + database.VerbalQuestionWordJoinWordField).
		Where(squirrel.Eq{"vqw." + database.VerbalQuestionWordJoinVerbalField: ids}).
		PlaceholderFormat(squirrel.Dollar)
	sqlQuery, args, err := query.ToSql()
	if err != nil {
		return nil, err
	}
	rows, err := s.DB.Query(ctx, sqlQuery, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	vocabulary := make(map[int][]models.Word)
	var meaningsJson []byte
	for rows.Next() {
		var questionID int
		var word models.Word
		err = rows.Scan(&questionID, &word.ID, &word.Word, &meaningsJson, &word.Examples)
		if err != nil {
			return nil, err
		}
		err = json.Unmarshal(meaningsJson, &word.Meanings)
		if err != nil {
			return nil, err
		}
		vocabulary[questionID] = append(vocabulary[questionID], word)
	}
	return vocabulary, nil
}

func (s *UserVerbalStatsService) GetVerbalStatsByUserToken(ctx context.Context, userToken string) ([]models.UserVerbalStat, error) {
	query := squirrel.Select(
		"vs.*",
		"q."+database.VerbalQuestionsCompetenceField,
		"q."+database.VerbalQuestionsFramedAsField,
		"q."+database.VerbalQuestionsTypeField,
		"q."+database.VerbalQuestionsDifficultyField,
	).
		From(database.VerbalStatsTable + " AS vs").
		Join(database.VerbalQuestionsTable + " AS q ON vs." + database.VerbalStatsQuestionField + " = q." + database.VerbalQuestionsIDField).
		Where(squirrel.Eq{"vs." + database.VerbalStatsUserField: userToken}).
		PlaceholderFormat(squirrel.Dollar)
	sqlQuery, args, err := query.ToSql()
	if err != nil {
		return nil, err
	}
	rows, err := s.DB.Query(ctx, sqlQuery, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	verbalStats := make([]models.UserVerbalStat, 0)
	for rows.Next() {
		var verbalStat models.UserVerbalStat
		err := rows.Scan(&verbalStat.ID, &verbalStat.UserToken, &verbalStat.QuestionID, &verbalStat.Correct, &verbalStat.Answers, &verbalStat.Duration, &verbalStat.Date,
			&verbalStat.Competence, &verbalStat.FramedAs, &verbalStat.Type, &verbalStat.Difficulty)
		if err != nil {
			return nil, err
		}
		verbalStats = append(verbalStats, verbalStat)
	}
	// After you get the list of UserVerbalStat:
	questionIDs := make([]int, len(verbalStats))
	for i, verbalStat := range verbalStats {
		questionIDs[i] = verbalStat.QuestionID
	}
	// Get the vocabulary words for each question.
	vocabulary, err := s.GetVocabularyByQuestionIDs(ctx, questionIDs)
	if err != nil {
		return nil, err
	}
	// Add vocabulary words to each question.
	for i, verbalStat := range verbalStats {
		verbalStats[i].Vocabulary = vocabulary[verbalStat.QuestionID]
	}
	return verbalStats, nil
}
