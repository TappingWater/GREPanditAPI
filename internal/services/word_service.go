// services/word_service.go
package services

import (
	"context"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/labstack/echo/v4"
	"github.com/lib/pq"
	"grepandit.com/api/internal/models"
)

type WordService struct {
	DB *pgxpool.Pool
}

func NewWordService(db *pgxpool.Pool) *WordService {
	return &WordService{DB: db}
}

func (s *WordService) Create(ctx context.Context, w *models.Word) error {
	query := `INSERT INTO words (word, meanings, examples) VALUES ($1, $2, $3) RETURNING id`
	return s.DB.QueryRow(ctx, query, w.Word, pq.Array(w.Meanings), pq.Array(w.Examples)).Scan(&w.ID)
}

func (s *WordService) GetByID(ctx context.Context, id int) (*models.Word, error) {
	w := &models.Word{}
	query := `SELECT * FROM words WHERE id = $1`
	err := s.DB.QueryRow(ctx, query, id).Scan(&w.ID, &w.Word, pq.Array(&w.Meanings), pq.Array(&w.Examples))
	if err != nil {
		if err.Error() == "no rows in result set" {
			return nil, echo.ErrNotFound
		}
		return nil, err
	}
	return w, nil
}

func (s *WordService) GetByWord(ctx context.Context, word string) (*models.Word, error) {
	w := &models.Word{}
	query := `SELECT * FROM words WHERE word = $1`
	err := s.DB.QueryRow(ctx, query, word).Scan(&w.ID, &w.Word, pq.Array(&w.Meanings), pq.Array(&w.Examples))
	if err != nil {
		if err.Error() == "no rows in result set" {
			return nil, echo.ErrNotFound
		}
		return nil, err
	}
	return w, nil
}
