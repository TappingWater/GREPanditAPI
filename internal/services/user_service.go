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
