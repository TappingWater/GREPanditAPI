package services

import (
	"context"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/labstack/echo/v4"
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
	query := `
		INSERT INTO ` + database.UsersTable + ` (` +
		database.UserTokenField + `, ` +
		database.UserEmailField + `)
		VALUES ($1, $2)
		RETURNING ` + database.UserIDField

	return s.DB.QueryRow(ctx, query, u.Token, u.Email).Scan(&u.ID)
}

func (s *UserService) Update(ctx context.Context, id int, u *models.User) error {
	query := `
		UPDATE ` + database.UsersTable + `
		SET ` + database.UserTokenField + ` = $1, ` +
		database.UserEmailField + ` = $2
		WHERE ` + database.UserIDField + ` = $3`

	_, err := s.DB.Exec(ctx, query, u.Token, u.Email, id)
	return err
}

func (s *UserService) GetByID(ctx context.Context, id int) (*models.User, error) {
	u := &models.User{}
	query := `
		SELECT * FROM ` + database.UsersTable + `
		WHERE ` + database.UserIDField + ` = $1`

	err := s.DB.QueryRow(ctx, query, id).Scan(&u.ID, &u.Token, &u.Email)
	if err != nil {
		if err.Error() == "no rows in result set" {
			return nil, echo.ErrNotFound
		}
		return nil, err
	}

	return u, nil
}

func (s *UserService) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	u := &models.User{}
	query := `
		SELECT * FROM ` + database.UsersTable + `
		WHERE ` + database.UserEmailField + ` = $1`

	err := s.DB.QueryRow(ctx, query, email).Scan(&u.ID, &u.Token, &u.Email)
	if err != nil {
		if err.Error() == "no rows in result set" {
			return nil, echo.ErrNotFound
		}
		return nil, err
	}

	return u, nil
}

func (s *UserService) GetByUserToken(ctx context.Context, userToken string) (*models.User, error) {
	u := &models.User{}
	query := `
		SELECT * FROM ` + database.UsersTable + `
		WHERE ` + database.UserTokenField + ` = $1`

	err := s.DB.QueryRow(ctx, query, userToken).Scan(&u.ID, &u.Token, &u.Email)
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

func (s *UserService) GetMarkedWordsByUserToken(ctx context.Context, userToken string) ([]models.UserMarkedWord, error) {
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
