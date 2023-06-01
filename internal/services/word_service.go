// services/word_service.go
package services

import (
	"context"
	"encoding/json"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/labstack/echo/v4"
	"grepandit.com/api/internal/models"
)

type WordService struct {
	DB *pgxpool.Pool
}

func NewWordService(db *pgxpool.Pool) *WordService {
	return &WordService{DB: db}
}

func (s *WordService) Create(ctx context.Context, w *models.Word) error {
	meaningsJson, err := json.Marshal(w.Meanings)
	if err != nil {
		return err
	}
	query := `INSERT INTO words (word, meanings) VALUES ($1, $2) RETURNING id`
	return s.DB.QueryRow(ctx, query, w.Word, meaningsJson).Scan(&w.ID)
}

func (s *WordService) GetByID(ctx context.Context, id int) (*models.Word, error) {
	w := &models.Word{}
	query := `SELECT * FROM words WHERE id = $1`
	var meaningsJson []byte
	err := s.DB.QueryRow(ctx, query, id).Scan(&w.ID, &w.Word, &meaningsJson)
	if err != nil {
		if err.Error() == "no rows in result set" {
			return nil, echo.ErrNotFound
		}
		return nil, err
	}
	err = json.Unmarshal(meaningsJson, &w.Meanings)
	if err != nil {
		return nil, err
	}
	return w, nil
}

func (s *WordService) GetByWord(ctx context.Context, word string) (*models.Word, error) {
	w := &models.Word{}
	query := `SELECT * FROM words WHERE word = $1`
	var meaningsJson []byte
	err := s.DB.QueryRow(ctx, query, word).Scan(&w.ID, &w.Word, &meaningsJson)
	if err != nil {
		if err.Error() == "no rows in result set" {
			return nil, echo.ErrNotFound
		}
		return nil, err
	}
	err = json.Unmarshal(meaningsJson, &w.Meanings)
	if err != nil {
		return nil, err
	}
	return w, nil
}
