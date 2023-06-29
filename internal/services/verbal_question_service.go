// services/verbal_question_service.go
package services

import (
	"context"
	"encoding/json"
	"math/rand"
	"strings"

	"github.com/Masterminds/squirrel"
	"github.com/aaaton/golem/v4"
	"github.com/aaaton/golem/v4/dicts/en"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/labstack/echo/v4"
	"github.com/stitchfix/mab"
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

type UserRewardSource struct {
	user *models.User
}

func (u *UserRewardSource) GetRewards(ctx context.Context, banditContext interface{}) ([]mab.Dist, error) {
	// Possible Combinations
	difficulties := [3]string{"Easy", "Medium", "Hard"}
	qTypes := [3]string{"ReadingComprehension", "TextCompletion", "SentenceEquivalence"}
	combinations := make([]string, 9)
	counter := 0
	for _, diff := range difficulties {
		for _, qType := range qTypes {
			combinations[counter] = diff + "_" + qType
			counter++
		}
	}
	distributions := make([]mab.Dist, len(combinations))
	for i, combination := range combinations {
		ability, ok := u.user.VerbalAbility[combination]
		var reward float64
		if ok {
			reward = 1 - float64(ability/u.user.VerbalAbilityCount[combination]) // We subtract from 1 to recommend questions the user is not good at
		} else {
			reward = 0.5 + rand.Float64()*(1-0.5)
		}
		distributions[i] = mab.Point(reward)
	}
	return distributions, nil
}

/**
* Fetch a list of questions that are adaptive based on the user
**/
func (s *VerbalQuestionService) GetAdaptiveQuestions(ctx context.Context, userToken string,
	numQuestions int, excludeIds []int) ([]*models.VerbalQuestion, error) {
	// Get the user's performance stats
	us := NewUserService(s.DB)
	user, err := us.Get(ctx, userToken)
	if err != nil {
		return nil, err
	}
	// Create a reward source
	rewardSource := &UserRewardSource{user: user}
	// Initialize a new epsilon-greedy bandit with epsilon=0.4 and the reward source
	strategy := mab.NewEpsilonGreedy(0.2)
	sampler := mab.NewSha1Sampler()
	bandit := mab.Bandit{
		RewardSource: rewardSource,
		Strategy:     strategy,
		Sampler:      sampler,
	}
	difficulties := [3]string{"Easy", "Medium", "Hard"}
	qTypes := [3]string{"ReadingComprehension", "TextCompletion", "SentenceEquivalence"}
	combinations := make([]string, 9)
	counter := 0
	for _, diff := range difficulties {
		for _, qType := range qTypes {
			combinations[counter] = diff + "_" + qType
			counter++
		}
	}
	// Use the bandit to choose the competencies for the questions
	selectedCombinations := make([]string, numQuestions)
	for i := 0; i < numQuestions; i++ {
		result, err := bandit.SelectArm(context.Background(), userToken, nil)
		if err != nil {
			return nil, err
		}
		selectedCombinations[i] = combinations[result.Arm]
	}
	// Get a question from each selected competency
	questions := make([]*models.VerbalQuestion, numQuestions)
	successCount := 0
	for _, combination := range selectedCombinations {
		parts := strings.Split(combination, "_")
		question, err := s.GetByCriteria(ctx, parts[0], parts[1], excludeIds)
		if err != nil {
			// If an error occurred, just move to the next one.
			print(err.Error())
			continue
		}
		// Only assign to the slice when we have a successful question.
		questions[successCount] = question
		successCount++
	}
	print(successCount)
	// If we have less than numQuestions, resize the slice.
	if successCount < numQuestions {
		questions = questions[:successCount]
	}
	// New UserVerbalStatsService instance
	uvss := NewUserVerbalStatsService(s.DB)
	// Get the question IDs
	questionIDs := make([]int, successCount)
	for i, question := range questions {
		questionIDs[i] = question.ID
	}
	vocabulary, err := uvss.GetVocabularyByQuestionIDs(ctx, questionIDs)
	if err != nil {
		return nil, err
	}
	for i, question := range questions {
		questions[i].Vocabulary = vocabulary[question.ID]
	}
	return questions, nil
}

/**
* Retrieve verbal questions at random based on particular parameters
* to display to the user.
**/
func (s *VerbalQuestionService) GetByCriteria(
	ctx context.Context,
	difficulty string,
	qType string,
	excludeIDs []int,
) (*models.VerbalQuestion, error) {
	difficultyEnum, err := models.StringToDifficulty(difficulty)
	if err != nil {
		return nil, err
	}
	qTypeEnum, err := models.StringToQuestionType(qType)
	if err != nil {
		return nil, err
	}
	// Build the SQL query
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
		PlaceholderFormat(squirrel.Dollar)
	query = query.Where(squirrel.Eq{database.VerbalQuestionsTypeField: qTypeEnum})
	query = query.Where(squirrel.Eq{database.VerbalQuestionsDifficultyField: difficultyEnum})
	if len(excludeIDs) > 0 {
		query = query.Where(squirrel.NotEq{database.VerbalQuestionsIDField: excludeIDs})
	}
	// Execute the SQL query
	sqlQuery, args, err := query.ToSql()
	if err != nil {
		return nil, err
	}
	q := &models.VerbalQuestion{}
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
	return q, nil
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

func (s *VerbalQuestionService) GetQuestionsOnVocab(
	ctx context.Context,
	userToken string,
	excludeQuestionIDs []int,
	wordIDs []int,
) ([]*models.VerbalQuestion, error) {
	// Query the question-word join table for question ids based on the word ids
	query := squirrel.Select(database.VerbalQuestionWordJoinVerbalField).
		From(database.VerbalQuestionWordsJoinTable).
		Where(squirrel.Eq{database.VerbalQuestionWordJoinWordField: wordIDs}).
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
	// Collect the question IDs
	questionIDs := make([]int, 0)
	for rows.Next() {
		var id int
		err = rows.Scan(&id)
		if err != nil {
			return nil, err
		}
		// Exclude question IDs from the exclude list
		if !contains(excludeQuestionIDs, id) {
			questionIDs = append(questionIDs, id)
		}
	}
	// If there are more than 5 question IDs, randomly select 5
	if len(questionIDs) > 5 {
		rand.Shuffle(len(questionIDs), func(i, j int) { questionIDs[i], questionIDs[j] = questionIDs[j], questionIDs[i] })
		questionIDs = questionIDs[:5]
	}
	// Call the GetByIDs function with the final question IDs
	return s.GetByIDs(ctx, questionIDs)
}

// Helper function to check if a slice contains a value
func contains(slice []int, val int) bool {
	for _, item := range slice {
		if item == val {
			return true
		}
	}
	return false
}
