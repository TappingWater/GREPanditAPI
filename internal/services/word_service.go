// services/word_service.go
package services

import (
	"context"
	"encoding/json"

	"github.com/aaaton/golem/v4"
	"github.com/aaaton/golem/v4/dicts/en"
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
	// Lemmetize to get base forms of words and find variations
	lemmatizer, err := golem.New(en.New())
	if err != nil {
		println(err)
		return err
	}
	baseForm := lemmatizer.Lemma(w.Word)
	meaningsJson, err := json.Marshal(w.Meanings)
	if err != nil {
		return err
	}
	query := `INSERT INTO words (word, meanings) VALUES ($1, $2) RETURNING id`
	return s.DB.QueryRow(ctx, query, baseForm, meaningsJson).Scan(&w.ID)
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
	// Lemmetize to get base forms of words and find variations
	lemmatizer, err := golem.New(en.New())
	if err != nil {
		println(err)
		return nil, err
	}
	baseForm := lemmatizer.Lemma(word)
	w := &models.Word{}
	query := `SELECT * FROM words WHERE word = $1`
	var meaningsJson []byte
	err = s.DB.QueryRow(ctx, query, baseForm).Scan(&w.ID, &w.Word, &meaningsJson)
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
