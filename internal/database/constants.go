package database

// Table names
const (
	WordsTable                     = "words"
	VerbalQuestionsTable           = "verbal_questions"
	VerbalQuestionWordsJoinTable   = "verbal_question_words"
	UsersTable                     = "users"
	VerbalStatsTable               = "verbal_stats"
	UserVerbalStatsJoinTable       = "user_verbal_stats"
	UserMarkedWordsTable           = "user_marked_words"
	UserMarkedVerbalQuestionsTable = "user_marked_verbal_questions"
)

// Words field names
const (
	WordsIDField       = "id"
	WordsWordField     = "word"
	WordsMeaningsField = "meanings"
)

// VerbalQuestions field names
const (
	VerbalQuestionsIDField          = "id"
	VerbalQuestionsCompetenceField  = "competence"
	VerbalQuestionsFramedAsField    = "framed_as"
	VerbalQuestionsTypeField        = "type"
	VerbalQuestionsParagraphField   = "paragraph"
	VerbalQuestionsQuestionField    = "question"
	VerbalQuestionsOptionsField     = "options"
	VerbalQuestionsAnswerField      = "answer"
	VerbalQuestionsWordField        = "word"
	VerbalQuestionsExplanationField = "explanation"
	VerbalQuestionsDifficultyField  = "difficulty"
	VerbalQuestionsWordmapField     = "wordmap"
)

// Join table for users and verbal questions
const (
	VerbalQuestionWordJoinVerbalField = "verbal_question_id"
	VerbalQuestionWordJoinWordField   = "word_id"
)

// User field names
const (
	UserIDField    = "id"
	UserTokenField = "token"
	UserEmailField = "email"
)

// VerbalStats field names
const (
	VerbalStatsIDField       = "id"
	VerbalStatsUserField     = "user_token"
	VerbalStatsQuestionField = "question_id"
	VerbalStatsCorrectField  = "correct"
	VerbalStatsAnswersField  = "answers"
	VerbalStatsDateField     = "date"
)

// User Verbal Stats Join table field names
const (
	UserVerbalStatsJoinIDField     = "id"
	UserVerbalStatsJoinVerbalField = "stats_id"
	UserVerbalStatsJoinUserField   = "user_token"
)

// User Marked Words table field names
const (
	UserMarkedWordsIDField   = "id"
	UserMarkedWordsUserField = "user_token"
	UserMarkedWordsWordField = "word_id"
)

// User Marked Verbal Questions field names
const (
	UserMarkedVerbalQuestionsIDField       = "id"
	UserMarkedVerbalQuestionsUserField     = "user_token"
	UserMarkedVerbalQuestionsQuestionField = "verbal_question"
)
