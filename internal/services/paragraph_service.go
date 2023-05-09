// services/paragraph_service.go
package services

import (
	"context"

	"github.com/jackc/pgx/v4/pgxpool"
	"grepandit.com/api/internal/models"
)

type ParagraphService struct {
	DB *pgxpool.Pool
}

func NewParagraphService(db *pgxpool.Pool) *ParagraphService {
	return &ParagraphService{DB: db}
}

func (s *ParagraphService) Create(p *models.Paragraph) error {
	ctx := context.Background()
	query := `INSERT INTO paragraphs (paragraph_text) VALUES ($1)`
	_, err := s.DB.Exec(ctx, query, p.Text)
	return err
}

func (s *ParagraphService) GetByID(id int) (*models.Paragraph, error) {
	ctx := context.Background()
	p := &models.Paragraph{}
	query := `SELECT * FROM paragraphs WHERE id = $1`
	err := s.DB.QueryRow(ctx, query, id).Scan(&p.ID, &p.Text)
	return p, err
}

func (s *ParagraphService) Update(p *models.Paragraph) error {
	ctx := context.Background()
	query := `UPDATE paragraphs SET paragraph_text=$2 WHERE id = $1`
	_, err := s.DB.Exec(ctx, query, p.ID, p.Text)
	return err
}

func (s *ParagraphService) Delete(id int) error {
	ctx := context.Background()
	query := `DELETE FROM paragraphs WHERE id = $1`
	_, err := s.DB.Exec(ctx, query, id)
	return err
}
