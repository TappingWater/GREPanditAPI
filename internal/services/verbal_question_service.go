package services

import (
	"context"
	"encoding/json"

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

/**
* Creates a new record in the Db for verbal question. It also retrieves
* the id of words based on the string words sent for the question and creates
* a new record in the join table for each word associated with the question.
**/
func (s *VerbalQuestionService) Create(
	ctx context.Context,
	q *models.VerbalQuestionRequest,
) error {
	// Begin a transaction.
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return err
	}
	// Rollback in case of error. This is a no-op if the transaction has been committed.
	defer tx.Rollback(ctx)
	optionsJson, err := json.Marshal(q.Options)
	if err != nil {
		return err
	}
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
			optionsJson,
			pq.Array(q.Answer),
			q.Explanation,
			q.Difficulty).
		Suffix("RETURNING id").
		PlaceholderFormat(squirrel.Dollar)
	sqlQuery, args, err := query.ToSql()
	if err != nil {
		println(err)
		return err
	}
	err = tx.QueryRow(ctx, sqlQuery, args...).Scan(&q.ID)
	if err != nil {
		println(err)
		return err
	}
	// Now associate the words with the new verbal question.
	for _, word := range q.Vocabulary {
		// Get the ID of the word.
		var wordID int
		err = tx.QueryRow(ctx, `SELECT id FROM words WHERE word = $1`, word).Scan(&wordID)
		if err != nil {
			println(err)
			return err
		}
		// Create a new record in the verbal_question_words table.
		_, err = tx.Exec(ctx, `INSERT INTO verbal_question_words (verbal_question_id, word_id) VALUES ($1, $2)`, q.ID, wordID)
		if err != nil {
			println(err)
			return err
		}
	}
	// If we reach this point, all database operations have been successful. Commit the transaction.
	return tx.Commit(ctx)
}

/**
* Retrieve verbal question by its ID. A join operation is done using the join table
* to get the list of words associated with this table.
**/
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
		Where(squirrel.Eq{"q.id": id}).
		PlaceholderFormat(squirrel.Dollar)
	sqlQuery, args, err := query.ToSql()
	if err != nil {
		println(err.Error())
		return nil, err
	}
	var optionsJson []byte
	err = s.DB.QueryRow(ctx, sqlQuery, args...).
		Scan(
			&q.ID,
			&q.Competence,
			&q.FramedAs,
			&q.Type,
			&q.Paragraph,
			&q.Question,
			&optionsJson,
			pq.Array(&q.Answer),
			&q.Explanation,
			&q.Difficulty,
		)
	if err != nil {
		if err == pgx.ErrNoRows {
			println(err.Error())
			return nil, echo.ErrNotFound
		}
		return nil, err
	}
	err = json.Unmarshal(optionsJson, &q.Options)
	if err != nil {
		return nil, err
	}
	// Now get the vocabulary words.
	rows, err := s.DB.Query(ctx, `
		SELECT w.id, w.word, w.meanings
		FROM words AS w
		INNER JOIN verbal_question_words AS vqw ON w.id = vqw.word_id
		WHERE vqw.verbal_question_id = $1
	`, id)
	if err != nil {
		println(err.Error())
		return nil, err
	}
	defer rows.Close()

	q.Vocabulary = make([]models.Word, 0)
	var meaningsJson []byte
	for rows.Next() {
		var word models.Word
		err = rows.Scan(&word.ID, &word.Word, &meaningsJson)
		if err != nil {
			println(err.Error())
			return nil, err
		}
		err = json.Unmarshal(meaningsJson, &word.Meanings)
		if err != nil {
			return nil, err
		}
		q.Vocabulary = append(q.Vocabulary, word)
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

/**
* Retrieve verbal questions at random based on particular parameters
* to display to the user.
**/
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
		"w.id",
		"w.word",
		"w.meanings",
	).
		From("verbal_questions AS q").
		LeftJoin("verbal_question_words AS a ON q.id = a.verbal_question_id").
		LeftJoin("words AS w ON a.word_id = w.id").
		OrderBy("RANDOM()").
		Limit(uint64(limit))
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
	sqlQuery, args, err := query.ToSql()
	if err != nil {
		return nil, err
	}
	rows, err := s.DB.Query(ctx, sqlQuery, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	questionsMap := make(map[int]models.VerbalQuestion)
	wordsMap := make(map[int][]models.Word)
	for rows.Next() {
		var q models.VerbalQuestion
		var word models.Word
		var meaningsJson []byte
		err := rows.Scan(
			&q.ID,
			&q.Competence,
			&q.FramedAs,
			&q.Type,
			&q.Paragraph,
			&q.Question,
			&q.Options,
			&q.Answer,
			&q.Explanation,
			&q.Difficulty,
			&word.ID,
			&word.Word,
			&meaningsJson,
		)
		if err != nil {
			return nil, err
		}
		err = json.Unmarshal(meaningsJson, &word.Meanings)
		if err != nil {
			return nil, err
		}
		// If the question is already in the map, add the word to the word map.
		// If it's not, add the question to the question map and the word to the word map.
		if _, ok := questionsMap[q.ID]; ok {
			wordsMap[q.ID] = append(wordsMap[q.ID], word)
		} else {
			q.Vocabulary = []models.Word{word}
			questionsMap[q.ID] = q
			wordsMap[q.ID] = []models.Word{word}
		}
	}
	// Assign the words from the wordsMap to the vocabulary of the corresponding question in the questionsMap.
	for id, question := range questionsMap {
		question.Vocabulary = wordsMap[id]
		questionsMap[id] = question
	}
	questions := make([]models.VerbalQuestion, 0, len(questionsMap))
	for _, question := range questionsMap {
		questions = append(questions, question)
	}
	return questions, nil
}
