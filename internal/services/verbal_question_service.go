package services

import (
	"context"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/lib/pq"
	"grepandit.com/api/internal/models"
)

type VerbalQuestionService struct {
	DB *pgxpool.Pool
}

func NewVerbalQuestionService(db *pgxpool.Pool) *VerbalQuestionService {
	return &VerbalQuestionService{DB: db}
}

func (s *VerbalQuestionService) Create(q *models.VerbalQuestion) error {
	ctx := context.Background()
	query := `INSERT INTO verbal_questions (competence, framed_as, type, paragraph_id, question, options, answer, explanation, difficulty) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`
	_, err := s.DB.Exec(ctx, query, q.Competence, q.FramedAs, q.Type, q.ParagraphID, q.Question, pq.Array(q.Options), pq.Array(q.Answer), q.Explanation, q.Difficulty)
	return err
}

func (s *VerbalQuestionService) Get(id int) (*models.VerbalQuestion, error) {
	ctx := context.Background()
	q := &models.VerbalQuestion{}
	query := `SELECT * FROM verbal_questions WHERE id = $1`
	err := s.DB.QueryRow(ctx, query, id).Scan(&q.VerbalQuestionID, &q.Competence, &q.FramedAs, &q.Type, &q.ParagraphID, &q.Question, pq.Array(&q.Options), pq.Array(&q.Answer), &q.Explanation, &q.Difficulty)
	return q, err
}

func (s *VerbalQuestionService) Update(q *models.VerbalQuestion) error {
	ctx := context.Background()
	query := `UPDATE verbal_questions SET competence=$2, framed_as=$3, type=$4, paragraph_id=$5, question=$6, options=$7, answer=$8, explanation=$9, difficulty=$10 WHERE verbal_question_id = $1`
	_, err := s.DB.Exec(ctx, query, q.VerbalQuestionID, q.Competence, q.FramedAs, q.Type, q.ParagraphID, q.Question, pq.Array(q.Options), pq.Array(q.Answer), q.Explanation, q.Difficulty)
	return err
}

func (s *VerbalQuestionService) Delete(id int) error {
	ctx := context.Background()
	query := `DELETE FROM verbal_questions WHERE verbal_question_id = $1`
	_, err := s.DB.Exec(ctx, query, id)
	return err
}
