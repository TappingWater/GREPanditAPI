package services

import (
	"context"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/labstack/echo/v4"
	"grepandit.com/api/internal/models"
)

type ParagraphService struct {
	DB *pgxpool.Pool
}

func NewParagraphService(db *pgxpool.Pool) *ParagraphService {
	return &ParagraphService{DB: db}
}

func (s *ParagraphService) Create(ctx context.Context, p *models.Paragraph) error {
	query := `INSERT INTO paragraphs (paragraph_text) VALUES ($1) RETURNING id`
	return s.DB.QueryRow(ctx, query, p.Text).Scan(&p.ID)
}

func (s *ParagraphService) GetByID(ctx context.Context, id int) (*models.Paragraph, error) {
	p := &models.Paragraph{}
	query := `SELECT * FROM paragraphs WHERE id = $1`
	err := s.DB.QueryRow(ctx, query, id).Scan(&p.ID, &p.Text)
	if err != nil {
		if err.Error() == "no rows in result set" {
			return nil, echo.ErrNotFound
		}
		return nil, err
	}
	return p, nil
}
