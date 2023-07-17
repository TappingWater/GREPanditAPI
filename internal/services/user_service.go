package services

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/labstack/echo/v4"
	"github.com/lib/pq"
	"grepandit.com/api/internal/database"
	"grepandit.com/api/internal/models"
)

type UserService struct {
	DB *pgxpool.Pool
}

func NewUserService(db *pgxpool.Pool) *UserService {
	return &UserService{DB: db}
}

func (s *UserService) Create(ctx context.Context, u *models.User) error {
	queryInsert := `
		INSERT INTO ` + database.UsersTable + ` (` +
		database.UserTokenField + `, ` +
		database.UserEmailField + `, ` +
		database.UserVerbalAbilityField + `)
		VALUES ($1, $2, $3)
		RETURNING ` + database.UserIDField
	err := s.DB.QueryRow(ctx, queryInsert, u.Token, u.Email, nil).Scan(&u.ID)
	if err != nil {
		// Check if it is a unique constraint violation error
		if pgErr, ok := err.(*pq.Error); ok {
			if pgErr.Code.Name() == "unique_violation" {
				return fmt.Errorf("User already exists")
			}
		}
		// Return other database-related errors as is
		return err
	}
	return nil
}
func (s *UserService) Update(ctx context.Context, u *models.User) error {
	queryUpdate := `
		UPDATE ` + database.UsersTable + `
		SET ` +
		database.UserTokenField + ` = $1, ` +
		database.UserEmailField + ` = $2, ` +
		database.UserVerbalAbilityField + ` = $3
		WHERE ` + database.UserIDField + ` = $4
		RETURNING ` + database.UserIDField
	err := s.DB.QueryRow(ctx, queryUpdate, u.Token, u.Email, u.VerbalAbility, u.ID).Scan(&u.ID)
	if err != nil {
		// Return other database-related errors as is
		return err
	}
	return nil
}

func (s *UserService) Get(ctx context.Context, userToken string) (*models.User, error) {
	u := &models.User{}
	query := `
		SELECT * FROM ` + database.UsersTable + `
		WHERE ` + database.UserTokenField + ` = $1`

	err := s.DB.QueryRow(ctx, query, userToken).Scan(&u.ID, &u.Token, &u.Email, &u.VerbalAbility)
	if err != nil {
		if err.Error() == "no rows in result set" {
			return nil, echo.ErrNotFound
		}
		return nil, err
	}
	return u, nil
}

// AddMarkedWords adds marked words for a user to the database.
func (s *UserService) AddMarkedWords(ctx context.Context, userToken string, wordIDs []int) error {
	// Create slice of user tokens for batch insert.
	userTokens := make([]string, len(wordIDs))
	for i := range userTokens {
		userTokens[i] = userToken
	}
	query := `
		INSERT INTO ` + database.UserMarkedWordsTable + ` (` +
		database.UserMarkedWordsUserField + `, ` +
		database.UserMarkedWordsWordField + `)
		SELECT * FROM UNNEST($1::TEXT[], $2::INT[])
		ON CONFLICT (` + database.UserMarkedWordsUserField + `, ` +
		database.UserMarkedWordsWordField + `) DO NOTHING`
	_, err := s.DB.Exec(ctx, query, userTokens, wordIDs)
	if err != nil {
		print(err.Error())
		return err
	}
	return nil
}

// AddMarkedQuestions adds marked questions for a user to the database.
func (s *UserService) AddMarkedQuestions(ctx context.Context, userToken string, questionIDs []int) error {
	// Create slice of user tokens for batch insert.
	userTokens := make([]string, len(questionIDs))
	for i := range userTokens {
		userTokens[i] = userToken
	}

	query := `
		INSERT INTO ` + database.UserMarkedVerbalQuestionsTable + ` (` +
		database.UserMarkedVerbalQuestionsUserField + `, ` +
		database.UserMarkedVerbalQuestionsQuestionField + `)
		SELECT * FROM UNNEST($1::TEXT[], $2::INT[])
		ON CONFLICT (` + database.UserMarkedVerbalQuestionsUserField + `, ` +
		database.UserMarkedVerbalQuestionsQuestionField + `) DO NOTHING`

	_, err := s.DB.Exec(ctx, query, userTokens, questionIDs)
	if err != nil {
		return err
	}

	return nil
}

// removes marked words for a user to the database.
func (s *UserService) RemoveMarkedWords(ctx context.Context, userToken string, wordIDs []int) error {
	array := pq.Array(wordIDs)
	query := `
		DELETE FROM ` + database.UserMarkedWordsTable + `
		WHERE ` + database.UserMarkedWordsUserField + ` = $1
		AND ` + database.UserMarkedWordsWordField + ` = ANY($2)`
	_, err := s.DB.Exec(ctx, query, userToken, array)
	if err != nil {
		return err
	}
	return nil
}

// removes marked questions for a user to the database.
func (s *UserService) RemoveMarkedQuestions(ctx context.Context, userToken string, questionIDs []int) error {
	array := pq.Array(questionIDs)
	query := `
		DELETE FROM ` + database.UserMarkedVerbalQuestionsTable + `
		WHERE ` + database.UserMarkedVerbalQuestionsUserField + ` = $1
		AND ` + database.UserMarkedVerbalQuestionsQuestionField + ` = ANY($2)`
	_, err := s.DB.Exec(ctx, query, userToken, array)
	if err != nil {
		return err
	}
	return nil
}

func (s *UserService) GetMarkedWordsByUserToken(ctx context.Context, userToken string) ([]models.UserMarkedWord, error) {
	query := squirrel.
		Select("u.*, w.*").
		From(database.UserMarkedWordsTable + " AS u").
		Join(database.WordsTable + " AS w ON u.word_id = w.id").
		Where(squirrel.Eq{"u.user_token": userToken}).
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
	var markedWords []models.UserMarkedWord
	for rows.Next() {
		var markedWord models.UserMarkedWord
		err := rows.Scan(
			&markedWord.ID,
			&markedWord.UserToken,
			&markedWord.WordID,
			&markedWord.Word.ID,
			&markedWord.Word.Word,
			&markedWord.Word.Meanings,
			&markedWord.Word.Examples,
			&markedWord.Word.Marked)
		if err != nil {
			return nil, err
		}
		markedWords = append(markedWords, markedWord)
	}
	return markedWords, nil
}

func (s *UserService) GetMarkedVerbalQuestionsByUserToken(ctx context.Context, userToken string) ([]models.UserMarkedVerbalQuestion, error) {
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

func (s *UserService) GetProblematicWordsByUserToken(ctx context.Context, userToken string) ([]models.Word, error) {
	// Query to get all incorrect stats for the user
	incorrectStatsQuery := squirrel.Select(
		"vs."+database.VerbalStatsQuestionField,
	).
		From(database.VerbalStatsTable+" AS vs").
		Where(
			squirrel.Eq{"vs." + database.VerbalStatsUserField: userToken},
			squirrel.Eq{"vs." + database.VerbalStatsCorrectField: false},
		).
		PlaceholderFormat(squirrel.Dollar)
	incorrectSqlQuery, args, err := incorrectStatsQuery.ToSql()
	if err != nil {
		return nil, err
	}
	rows, err := s.DB.Query(ctx, incorrectSqlQuery, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	// Get all the questions that the user answered incorrectly
	questionIDs := make([]int, 0)
	for rows.Next() {
		var questionID int
		err := rows.Scan(&questionID)
		if err != nil {
			return nil, err
		}
		questionIDs = append(questionIDs, questionID)
	}
	// Query to get all word IDs from verbal_question_words with those question IDs
	wordIDsQuery := squirrel.Select("vw." + database.VerbalQuestionWordJoinWordField).
		From(database.VerbalQuestionWordsJoinTable + " AS vw").
		Where(squirrel.Eq{"vw." + database.VerbalQuestionWordJoinVerbalField: questionIDs}).
		PlaceholderFormat(squirrel.Dollar)
	wordIDsSqlQuery, wordIDsArgs, err := wordIDsQuery.ToSql()
	if err != nil {
		return nil, err
	}
	wordIDsRows, err := s.DB.Query(ctx, wordIDsSqlQuery, wordIDsArgs...)
	if err != nil {
		return nil, err
	}
	defer wordIDsRows.Close()
	// Get a list of unique word ids
	uniqueWordIDs := make(map[int]struct{})
	for wordIDsRows.Next() {
		var wordID int
		err := wordIDsRows.Scan(&wordID)
		if err != nil {
			return nil, err
		}
		uniqueWordIDs[wordID] = struct{}{}
	}
	uniqueIDS := make([]int, 0, len(uniqueWordIDs))
	for k, _ := range uniqueWordIDs {
		uniqueIDS = append(uniqueIDS, k)
	}
	// Query to get all words from words table with those word IDs
	wordsQuery := squirrel.Select("*").
		From(database.WordsTable).
		Where(squirrel.Eq{database.WordsIDField: uniqueIDS}).
		PlaceholderFormat(squirrel.Dollar)
	wordsSqlQuery, wordsArgs, err := wordsQuery.ToSql()
	if err != nil {
		return nil, err
	}
	wordRows, err := s.DB.Query(ctx, wordsSqlQuery, wordsArgs...)
	if err != nil {
		return nil, err
	}
	defer wordRows.Close()
	words := make([]models.Word, 0)
	for wordRows.Next() {
		var w models.Word
		var meaningsJson []byte
		err := wordRows.Scan(&w.ID, &w.Word, &meaningsJson, &w.Examples, &w.Marked)
		if err != nil {
			return nil, err
		}
		err = json.Unmarshal(meaningsJson, &w.Meanings)
		if err != nil {
			return nil, err
		}
		words = append(words, w)
	}
	return words, nil
}
