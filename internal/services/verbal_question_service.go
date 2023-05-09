package services

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"

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

func (s *VerbalQuestionService) Create(ctx context.Context, q *models.VerbalQuestion) error {
	var paragraphID interface{} = nil
	if q.ParagraphID.Valid {
		paragraphID = q.ParagraphID.Int64
	}
	return s.DB.QueryRow(ctx, `
		INSERT INTO verbal_questions (competence, framed_as, type, paragraph_id, question, options, answer, explanation, difficulty)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id;
	`, q.Competence, q.FramedAs, q.Type, paragraphID, q.Question, pq.Array(q.Options), pq.Array(q.Answer), q.Explanation, q.Difficulty).Scan(&q.ID)
}

func (s *VerbalQuestionService) GetByID(ctx context.Context, id int) (*models.VerbalQuestion, error) {
	q := &models.VerbalQuestion{}
	err := s.DB.QueryRow(ctx, `
		SELECT q.id, q.competence, q.framed_as, q.type, q.paragraph_id, COALESCE(p.paragraph_text, ''), q.question, q.options, q.answer, q.explanation, q.difficulty
		FROM verbal_questions AS q
		LEFT JOIN paragraphs AS p ON q.paragraph_id = p.id
		WHERE q.id = $1;
	`, id).Scan(&q.ID, &q.Competence, &q.FramedAs, &q.Type, &q.ParagraphID, &q.ParagraphText, &q.Question, pq.Array(&q.Options), pq.Array(&q.Answer), &q.Explanation, &q.Difficulty)
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
	err := s.DB.QueryRow(ctx, `SELECT COUNT(*) FROM verbal_questions`).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (s *VerbalQuestionService) Random(ctx context.Context, limit, questionType, competence, framedAs, difficulty int, excludeIDs []int) ([]models.VerbalQuestion, error) {
	query := `SELECT * FROM verbal_questions`
	whereClauses := []string{}
	args := []interface{}{limit}
	if questionType != 0 {
		whereClauses = append(whereClauses, "type = $?")
		args = append(args, questionType)
	}
	if competence != 0 {
		whereClauses = append(whereClauses, "competence = $?")
		args = append(args, competence)
	}
	if framedAs != 0 {
		whereClauses = append(whereClauses, "framed_as = $?")
		args = append(args, framedAs)
	}
	if difficulty != 0 {
		whereClauses = append(whereClauses, "difficulty = $?")
		args = append(args, difficulty)
	}
	if len(excludeIDs) > 0 {
		placeholders := strings.Trim(strings.Repeat("$?,", len(excludeIDs)), ",")
		whereClauses = append(whereClauses, fmt.Sprintf("id NOT IN (%s)", placeholders))
		for _, id := range excludeIDs {
			args = append(args, id)
		}
	}
	if len(whereClauses) > 0 {
		query += " WHERE " + strings.Join(whereClauses, " AND ")
	}
	query += " ORDER BY RANDOM() LIMIT $1"
	rows, err := s.DB.Query(ctx, ReplaceSQLPlaceholders(query), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var questions []models.VerbalQuestion
	for rows.Next() {
		var q models.VerbalQuestion
		err = rows.Scan(&q.ID, &q.Competence, &q.FramedAs, &q.Type, &q.ParagraphID, &q.Question, &q.Options, &q.Answer, &q.Explanation, &q.Difficulty)
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

func ReplaceSQLPlaceholders(sql string) string {
	n := 0
	re := regexp.MustCompile(`\$\?`)
	return re.ReplaceAllStringFunc(sql, func(string) string {
		n++
		return "$" + strconv.Itoa(n)
	})
}
