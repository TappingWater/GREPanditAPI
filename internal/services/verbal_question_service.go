package services

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/Masterminds/squirrel"

	"github.com/aaaton/golem/v4"
	"github.com/aaaton/golem/v4/dicts/en"
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
	// Lemmetize to get base forms of words and find variations
	lemmatizer, err := golem.New(en.New())
	if err != nil {
		return err
	}
	// Convert vocab list to base forms
	vocabBaseForms := make(map[string]string)
	for _, word := range q.Vocabulary {
		lemmatized := lemmatizer.Lemma(word)
		vocabBaseForms[lemmatized] = word
	}
	// Find variations in paragraph and options
	variations := make(map[string]string)
	// Check paragraph
	// Split paragraph by whitespaces and periods
	paragraphWords := strings.FieldsFunc(q.Paragraph.String, func(r rune) bool {
		return r == ' ' || r == '.' || r == ',' || r == '!' || r == '(' || r == ')'
	})
	for _, word := range paragraphWords {
		lemmatized := lemmatizer.Lemma(word)
		if _, ok := vocabBaseForms[lemmatized]; ok {
			variations[word] = lemmatized
		}
	}
	// Iterate over Options
	for _, option := range q.Options {
		optionWords := strings.FieldsFunc(option.Value, func(r rune) bool {
			return r == ' ' || r == '.' || r == ',' || r == '!' || r == '(' || r == ')'
		})
		for _, word := range optionWords {
			lemmatized := lemmatizer.Lemma(word)
			if _, ok := vocabBaseForms[lemmatized]; ok {
				variations[word] = lemmatized
			}
		}
	}
	wordmapJson, err := json.Marshal(variations)
	if err != nil {
		return err
	}
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
			"difficulty",
			"wordmap").
		Values(
			q.Competence,
			q.FramedAs,
			q.Type,
			q.Paragraph,
			q.Question,
			optionsJson,
			pq.Array(q.Answer),
			q.Explanation,
			q.Difficulty,
			wordmapJson).
		Suffix("RETURNING id").
		PlaceholderFormat(squirrel.Dollar)
	sqlQuery, args, err := query.ToSql()
	if err != nil {
		return err
	}
	err = tx.QueryRow(ctx, sqlQuery, args...).Scan(&q.ID)
	if err != nil {
		return err
	}
	// Now associate the words with the new verbal question.
	for word := range vocabBaseForms {
		// Get the ID of the word.
		var wordID int
		err = tx.QueryRow(ctx, `SELECT id FROM words WHERE word = $1`, word).Scan(&wordID)
		if err != nil {
			return err
		}
		// Create a new record in the verbal_question_words table.
		_, err = tx.Exec(ctx, `INSERT INTO verbal_question_words (verbal_question_id, word_id) VALUES ($1, $2)`, q.ID, wordID)
		if err != nil {
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
		"q.wordmap",
	).
		From("verbal_questions AS q").
		Where(squirrel.Eq{"q.id": id}).
		PlaceholderFormat(squirrel.Dollar)
	sqlQuery, args, err := query.ToSql()
	if err != nil {
		return nil, err
	}
	var optionsJson []byte
	var wordMapJson []byte
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
			&wordMapJson,
		)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, echo.ErrNotFound
		}
		return nil, err
	}
	err = json.Unmarshal(optionsJson, &q.Options)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(wordMapJson, &q.VocabWordMap)
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
		return nil, err
	}
	defer rows.Close()

	q.Vocabulary = make([]models.Word, 0)
	var meaningsJson []byte
	for rows.Next() {
		var word models.Word
		err = rows.Scan(&word.ID, &word.Word, &meaningsJson)
		if err != nil {
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
	// Retrieve 5 random question IDs based on parameters
	sb := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	query := sb.Select("id").
		From("verbal_questions as q").
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
	questionIDs := make([]int, 0)
	for rows.Next() {
		var id int
		err := rows.Scan(&id)
		if err != nil {
			return nil, err
		}
		questionIDs = append(questionIDs, id)
	}
	// Based on retrieved question ids we can execute another query to get all the
	// words using the join table and fill up the vocabulary for each question
	vocabularyQuery := sb.Select(
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
		"q.wordmap",
		"w.id",
		"w.word",
		"w.meanings",
	).
		From("verbal_questions AS q").
		Join("verbal_question_words AS a ON q.id = a.verbal_question_id").
		Join("words AS w ON a.word_id = w.id").
		Where(squirrel.Eq{"q.id": questionIDs})

	vocabularySQL, vocabularyArgs, err := vocabularyQuery.ToSql()
	if err != nil {
		return nil, err
	}
	vocabularyRows, err := s.DB.Query(ctx, vocabularySQL, vocabularyArgs...)
	if err != nil {
		return nil, err
	}
	defer vocabularyRows.Close()
	questionsMap := make(map[int]models.VerbalQuestion)
	for vocabularyRows.Next() {
		var q models.VerbalQuestion
		var word models.Word
		var meaningsJSON []byte
		err := vocabularyRows.Scan(
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
			&q.VocabWordMap,
			&word.ID,
			&word.Word,
			&meaningsJSON,
		)
		if err != nil {
			return nil, err
		}
		err = json.Unmarshal(meaningsJSON, &word.Meanings)
		if err != nil {
			return nil, err
		}
		if existingQ, ok := questionsMap[q.ID]; ok {
			existingQ.Vocabulary = append(existingQ.Vocabulary, word)
			questionsMap[q.ID] = existingQ
		} else {
			q.Vocabulary = []models.Word{word}
			questionsMap[q.ID] = q
		}
	}
	questions := make([]models.VerbalQuestion, 0, len(questionsMap))
	for _, question := range questionsMap {
		questions = append(questions, question)
	}
	return questions, nil
}
