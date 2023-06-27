// services/word_service.go
package services

import (
	"context"
	"encoding/json"

	"github.com/Masterminds/squirrel"
	"github.com/aaaton/golem/v4"
	"github.com/aaaton/golem/v4/dicts/en"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/labstack/echo/v4"
	"grepandit.com/api/internal/database"
	"grepandit.com/api/internal/models"
)

type WordService struct {
	DB *pgxpool.Pool
}

func NewWordService(db *pgxpool.Pool) *WordService {
	return &WordService{DB: db}
}

func (s *WordService) Create(ctx context.Context, w *models.Word) error {
	// Lemmatize to get base forms of words and find variations
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
	query := squirrel.Insert(database.WordsTable).
		Columns(
			database.WordsWordField,
			database.WordsExamplesField,
			database.WordsMeaningsField,
			database.WordsMarkedField).
		Values(
			baseForm,
			w.Examples,
			meaningsJson,
			w.Marked).
		Suffix("RETURNING " + database.WordsIDField).
		PlaceholderFormat(squirrel.Dollar)
	sqlQuery, args, err := query.ToSql()
	if err != nil {
		return err
	}
	return s.DB.QueryRow(ctx, sqlQuery, args...).Scan(&w.ID)
}

func (s *WordService) GetByID(ctx context.Context, id int) (*models.Word, error) {
	w := &models.Word{}
	query := squirrel.Select("*").
		From(database.WordsTable).
		Where(squirrel.Eq{database.WordsIDField: id}).
		PlaceholderFormat(squirrel.Dollar)
	sqlQuery, args, err := query.ToSql()
	if err != nil {
		return nil, err
	}
	var meaningsJson []byte
	err = s.DB.QueryRow(
		ctx,
		sqlQuery,
		args...).Scan(
		&w.ID,
		&w.Word,
		&meaningsJson,
		&w.Examples,
		&w.Marked)
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
	// Lemmatize to get base forms of words and find variations
	lemmatizer, err := golem.New(en.New())
	if err != nil {
		println(err)
		return nil, err
	}
	baseForm := lemmatizer.Lemma(word)
	w := &models.Word{}
	query := squirrel.
		Select("*").
		From(database.WordsTable).
		Where(squirrel.Eq{database.WordsWordField: baseForm}).
		PlaceholderFormat(squirrel.Dollar)
	sqlQuery, args, err := query.ToSql()
	if err != nil {
		return nil, err
	}
	var meaningsJson []byte
	err = s.DB.QueryRow(
		ctx,
		sqlQuery,
		args...).Scan(
		&w.ID,
		&w.Word,
		&meaningsJson,
		&w.Examples,
		&w.Marked)
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

func (s *WordService) MarkWords(ctx context.Context, words []string) error {
	lemmatizer, err := golem.New(en.New())
	if err != nil {
		return err
	}
	baseFormsSet := make(map[string]struct{})
	for _, word := range words {
		baseForm := lemmatizer.Lemma(word)
		baseFormsSet[baseForm] = struct{}{}
	}
	baseForms := make([]string, 0, len(baseFormsSet))
	for baseForm := range baseFormsSet {
		baseForms = append(baseForms, baseForm)
	}
	query := squirrel.Update(database.WordsTable).
		Set("marked", true).
		Where(squirrel.Eq{"word": words}).
		PlaceholderFormat(squirrel.Dollar)
	sqlQuery, args, err := query.ToSql()
	if err != nil {
		return err
	}
	_, err = s.DB.Exec(ctx, sqlQuery, args...)
	if err != nil {
		return err
	}
	return nil
}

func (s *WordService) GetMarkedWords(ctx context.Context) ([]*models.Word, error) {
	// Construct the SQL query
	query := squirrel.
		Select("*").
		From(database.WordsTable).
		Where(squirrel.Eq{database.WordsMarkedField: true}).
		PlaceholderFormat(squirrel.Dollar)
	sqlQuery, args, err := query.ToSql()
	if err != nil {
		return nil, err
	}
	// Execute the SQL query
	rows, err := s.DB.Query(ctx, sqlQuery, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	// Iterate over rows and unmarshal JSON
	var words []*models.Word
	for rows.Next() {
		w := &models.Word{}
		var meaningsJson []byte
		if err := rows.Scan(&w.ID, &w.Word, &meaningsJson, &w.Examples, &w.Marked); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(meaningsJson, &w.Meanings); err != nil {
			return nil, err
		}
		words = append(words, w)
	}
	// Check for errors from iterating over rows.
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return words, nil
}
