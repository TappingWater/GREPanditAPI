// services/word_service.go
package services

import (
	"context"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/lib/pq"
	"grepandit.com/api/internal/models"
)

type WordService struct {
	DB *pgxpool.Pool
}

func NewWordService(db *pgxpool.Pool) *WordService {
	return &WordService{DB: db}
}

func (s *WordService) Create(w *models.Word) error {
	ctx := context.Background()
	query := `INSERT INTO words (word, meanings) VALUES ($1, $2)`
	_, err := s.DB.Exec(ctx, query, w.Word, pq.Array(w.Meanings), pq.Array(w.Examples))
	return err
}

func (s *WordService) GetByID(id int) (*models.Word, error) {
	ctx := context.Background()
	w := &models.Word{}
	query := `SELECT * FROM words WHERE id = $1`
	err := s.DB.QueryRow(ctx, query, id).Scan(&w.ID, &w.Word, pq.Array(&w.Meanings), pq.Array(&w.Examples))
	return w, err
}

func (s *WordService) GetByWord(word string) (*models.Word, error) {
	ctx := context.Background()
	w := &models.Word{}
	query := `SELECT * FROM words WHERE word = $1`
	err := s.DB.QueryRow(ctx, query, word).Scan(&w.ID, &w.Word, pq.Array(&w.Meanings), pq.Array(&w.Examples))
	return w, err
}

func (s *WordService) Update(w *models.Word) error {
	ctx := context.Background()
	query := `UPDATE words SET word=$2, meanings=$3 WHERE id = $1`
	_, err := s.DB.Exec(ctx, query, w.ID, w.Word, pq.Array(w.Meanings), pq.Array(w.Examples))
	return err
}

func (s *WordService) Delete(id int) error {
	ctx := context.Background()
	query := `DELETE FROM words WHERE id = $1`
	_, err := s.DB.Exec(ctx, query, id)
	return err
}
