// services/verbal_question_service.go
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
	"grepandit.com/api/internal/database"
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
	paragraphWords := strings.FieldsFunc(q.Paragraph, func(r rune) bool {
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
	query := squirrel.Insert(database.VerbalQuestionsTable).
		Columns(
			database.VerbalQuestionsCompetenceField,
			database.VerbalQuestionsFramedAsField,
			database.VerbalQuestionsTypeField,
			database.VerbalQuestionsParagraphField,
			database.VerbalQuestionsQuestionField,
			database.VerbalQuestionsOptionsField,
			database.VerbalQuestionsDifficultyField,
			database.VerbalQuestionsWordmapField).
		Values(
			q.Competence,
			q.FramedAs,
			q.Type,
			q.Paragraph,
			q.Question,
			optionsJson,
			q.Difficulty,
			wordmapJson).
		Suffix("RETURNING " + database.VerbalQuestionsIDField).
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
		err = tx.QueryRow(ctx, "SELECT "+database.WordsIDField+" FROM "+database.WordsTable+" WHERE "+database.WordsWordField+" = $1", word).Scan(&wordID)
		if err != nil {
			return err
		}
		// Create a new record in the verbal_question_words table.
		_, err = tx.Exec(ctx, "INSERT INTO "+database.VerbalQuestionWordsJoinTable+" ("+database.VerbalQuestionWordJoinVerbalField+", "+database.VerbalQuestionWordJoinWordField+") VALUES ($1, $2)", q.ID, wordID)
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
		database.VerbalQuestionsIDField,
		database.VerbalQuestionsCompetenceField,
		database.VerbalQuestionsFramedAsField,
		database.VerbalQuestionsTypeField,
		database.VerbalQuestionsParagraphField,
		database.VerbalQuestionsQuestionField,
		database.VerbalQuestionsOptionsField,
		database.VerbalQuestionsDifficultyField,
		database.VerbalQuestionsWordmapField,
	).
		From(database.VerbalQuestionsTable).
		Where(squirrel.Eq{database.VerbalQuestionsIDField: id}).
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
		SELECT w.`+database.WordsIDField+`, w.`+database.WordsWordField+`, w.`+database.WordsMeaningsField+`, w.`+database.WordsExamplesField+`
		FROM `+database.WordsTable+` AS w
		INNER JOIN `+database.VerbalQuestionWordsJoinTable+` AS vqw ON w.`+database.WordsIDField+` = vqw.`+database.VerbalQuestionWordJoinWordField+`
		WHERE vqw.`+database.VerbalQuestionWordJoinVerbalField+` = $1
	`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	q.Vocabulary = make([]models.Word, 0)
	var meaningsJson []byte
	for rows.Next() {
		var word models.Word
		err = rows.Scan(&word.ID, &word.Word, &meaningsJson, &word.Examples)
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

func (s *VerbalQuestionService) GetByIDs(
	ctx context.Context,
	ids []int,
) ([]*models.VerbalQuestion, error) {
	questions := make([]*models.VerbalQuestion, 0)
	query := squirrel.Select(
		database.VerbalQuestionsIDField,
		database.VerbalQuestionsCompetenceField,
		database.VerbalQuestionsFramedAsField,
		database.VerbalQuestionsTypeField,
		database.VerbalQuestionsParagraphField,
		database.VerbalQuestionsQuestionField,
		database.VerbalQuestionsOptionsField,
		database.VerbalQuestionsDifficultyField,
		database.VerbalQuestionsWordmapField,
	).
		From(database.VerbalQuestionsTable).
		Where(squirrel.Eq{database.VerbalQuestionsIDField: ids}).
		PlaceholderFormat(squirrel.Dollar)
	sqlQuery, args, err := query.ToSql()
	if err != nil {
		return nil, err
	}
	rows, err := s.DB.Query(ctx, sqlQuery, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		q := &models.VerbalQuestion{}
		var optionsJson []byte
		var wordMapJson []byte
		err = rows.Scan(
			&q.ID,
			&q.Competence,
			&q.FramedAs,
			&q.Type,
			&q.Paragraph,
			&q.Question,
			&optionsJson,
			&q.Difficulty,
			&wordMapJson,
		)
		if err != nil {
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
		wordsRows, err := s.DB.Query(ctx, `
			SELECT w.`+database.WordsIDField+`, w.`+database.WordsWordField+`, w.`+database.WordsMeaningsField+`, w.`+database.WordsExamplesField+`
			FROM `+database.WordsTable+` AS w
			INNER JOIN `+database.VerbalQuestionWordsJoinTable+` AS vqw ON w.`+database.WordsIDField+` = vqw.`+database.VerbalQuestionWordJoinWordField+`
			WHERE vqw.`+database.VerbalQuestionWordJoinVerbalField+` = $1
		`, q.ID)
		if err != nil {
			return nil, err
		}
		q.Vocabulary = make([]models.Word, 0)
		var meaningsJson []byte
		for wordsRows.Next() {
			var word models.Word
			err = wordsRows.Scan(&word.ID, &word.Word, &meaningsJson, &word.Examples)
			if err != nil {
				return nil, err
			}
			err = json.Unmarshal(meaningsJson, &word.Meanings)
			if err != nil {
				return nil, err
			}
			q.Vocabulary = append(q.Vocabulary, word)
		}
		wordsRows.Close()
		questions = append(questions, q)
	}
	return questions, nil
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
	query := sb.Select(database.VerbalQuestionsIDField).
		From(database.VerbalQuestionsTable + " as q").
		OrderBy("RANDOM()").
		Limit(uint64(limit))
	if questionType != 0 {
		query = query.Where(squirrel.Eq{database.VerbalQuestionsTypeField: questionType})
	}
	if competence != 0 {
		query = query.Where(squirrel.Eq{database.VerbalQuestionsCompetenceField: competence})
	}
	if framedAs != 0 {
		query = query.Where(squirrel.Eq{database.VerbalQuestionsFramedAsField: framedAs})
	}
	if difficulty != 0 {
		query = query.Where(squirrel.Eq{database.VerbalQuestionsDifficultyField: difficulty})
	}
	if len(excludeIDs) > 0 {
		query = query.Where(squirrel.NotEq{database.VerbalQuestionsIDField: excludeIDs})
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
		"q."+database.VerbalQuestionsIDField,
		"q."+database.VerbalQuestionsCompetenceField,
		"q."+database.VerbalQuestionsFramedAsField,
		"q."+database.VerbalQuestionsTypeField,
		"q."+database.VerbalQuestionsParagraphField,
		"q."+database.VerbalQuestionsQuestionField,
		"q."+database.VerbalQuestionsOptionsField,
		"q."+database.VerbalQuestionsDifficultyField,
		"q."+database.VerbalQuestionsWordmapField,
		"w."+database.WordsIDField,
		"w."+database.WordsWordField,
		"w."+database.WordsMeaningsField,
		"w."+database.WordsExamplesField,
	).
		From(database.VerbalQuestionsTable + " AS q").
		Join(database.VerbalQuestionWordsJoinTable + " AS a ON q." + database.VerbalQuestionsIDField + " = a." + database.VerbalQuestionWordJoinVerbalField).
		Join(database.WordsTable + " AS w ON a." + database.VerbalQuestionWordJoinWordField + " = w." + database.WordsIDField).
		Where(squirrel.Eq{"q." + database.VerbalQuestionsIDField: questionIDs})

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
			&q.Difficulty,
			&q.VocabWordMap,
			&word.ID,
			&word.Word,
			&meaningsJSON,
			&word.Examples,
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
