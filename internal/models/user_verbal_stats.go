package models

import "time"

type UserVerbalStat struct {
	ID         int          `json:"id"`
	UserToken  string       `json:"u_id"`
	QuestionID int          `json:"question_id"`
	Correct    bool         `json:"correct"`
	Answers    []string     `json:"answers"`
	Duration   int          `json:"duration"`
	Date       time.Time    `json:"time"`
	Competence Competence   `json:"competence"`
	FramedAs   FramedAs     `json:"framed_as"`
	Type       QuestionType `json:"type"`
	Difficulty Difficulty   `json:"difficulty"`
	Vocabulary []Word       `json:"vocabulary"`
}

type UserMarkedWord struct {
	ID        int    `json:"id"`
	UserToken string `json:"user_token"`
	WordID    int    `json:"word_id"`
	Word      Word   `json:"word"`
}

type UserMarkedVerbalQuestion struct {
	ID               int    `json:"id"`
	UserToken        string `json:"user_token"`
	VerbalQuestionID int    `json:"verbal_question_id"`
}
