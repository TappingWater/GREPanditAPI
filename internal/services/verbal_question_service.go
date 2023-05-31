package services

import (
	"context"

	"github.com/Masterminds/squirrel"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/labstack/echo/v4"
	"github.com/lib/pq"
	"grepandit.com/api/internal/models"
)

type VerbalQuestionService struct {
	DB *pgxpool.Pool
}

func NewVerbalQuestionService(db *pgxpool.Pool) *VerbalQuestionService {
	return &VerbalQuestionService{DB: db}
}

func (s *VerbalQuestionService) Create(
	ctx context.Context,
	q *models.VerbalQuestion,
) error {
	query := squirrel.Insert("verbal_questions").
		Columns(
			"competence",
			"framed_as",
			"type",
			"paragraph",
			"question",
			"options",
			"answer",
			"explanation",
			"difficulty").
		Values(
			q.Competence,
			q.FramedAs,
			q.Type,
			q.Paragraph,
			q.Question,
			pq.Array(q.Options),
			pq.Array(q.Answer),
			q.Explanation,
			q.Difficulty).
		Suffix("RETURNING id").
		PlaceholderFormat(squirrel.Dollar)
	sqlQuery, args, err := query.ToSql()
	if err != nil {
		return err
	}
	err = s.DB.QueryRow(ctx, sqlQuery, args...).Scan(&q.ID)
	return err
}

func (s *VerbalQuestionService) GetByID(
	ctx context.Context,
	id int,
) (*models.VerbalQuestion, error) {
	q := &models.VerbalQuestion{}
	query := squirrel.Select(
		"q.id",
		"q.competence",
		"q.framed_as",
		"q.type",
		"q.paragraph",
		"q.question",
		"q.options",
		"q.answer",
		"q.explanation",
		"q.difficulty",
	).
		From("verbal_questions AS q").
		PlaceholderFormat(squirrel.Dollar)
	sqlQuery, args, err := query.ToSql()
	if err != nil {
		return nil, err
	}
	err = s.DB.QueryRow(ctx, sqlQuery, args...).
		Scan(&q.ID, &q.Competence, &q.FramedAs, &q.Type, &q.Paragraph, &q.Question, pq.Array(&q.Options), pq.Array(&q.Answer), &q.Explanation, &q.Difficulty)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, echo.ErrNotFound
		}
		return nil, err
	}
	return q, nil
}

func (s *VerbalQuestionService) Count(ctx context.Context) (int, error) {
	var count int
	err := s.DB.QueryRow(ctx, `SELECT COUNT(*) FROM verbal_questions`).
		Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (s *VerbalQuestionService) Random(
	ctx context.Context,
	limit int,
	questionType models.QuestionType,
	competence models.Competence,
	framedAs models.FramedAs,
	difficulty models.Difficulty,
	excludeIDs []int,
) ([]models.VerbalQuestion, error) {
	sb := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	query := sb.Select(
		"q.id",
		"q.competence",
		"q.framed_as",
		"q.type",
		"q.paragraph",
		"q.question",
		"q.options",
		"q.answer",
		"q.explanation",
		"q.difficulty",
	).From("verbal_questions AS q")
	if questionType != 0 {
		query = query.Where(squirrel.Eq{"q.type": questionType})
	}
	if competence != 0 {
		query = query.Where(squirrel.Eq{"q.competence": competence})
	}
	if framedAs != 0 {
		query = query.Where(squirrel.Eq{"q.framed_as": framedAs})
	}
	if difficulty != 0 {
		query = query.Where(squirrel.Eq{"q.difficulty": difficulty})
	}
	if len(excludeIDs) > 0 {
		query = query.Where(squirrel.NotEq{"q.id": excludeIDs})
	}
	query = query.OrderBy("RANDOM()").Limit(uint64(limit))
	sqlQuery, args, err := query.ToSql()
	if err != nil {
		println(err)
		return nil, err
	}
	rows, err := s.DB.Query(ctx, sqlQuery, args...)
	if err != nil {
		println(err)
		return nil, err
	}
	defer rows.Close()
	var questions []models.VerbalQuestion
	for rows.Next() {
		var q models.VerbalQuestion
		err = rows.Scan(
			&q.ID,
			&q.Competence,
			&q.FramedAs,
			&q.Type,
			&q.Paragraph,
			&q.Question,
			pq.Array(&q.Options),
			pq.Array(&q.Answer),
			&q.Explanation,
			&q.Difficulty,
		)
		if err != nil {
			return nil, err
		}
		questions = append(questions, q)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return questions, nil
}
