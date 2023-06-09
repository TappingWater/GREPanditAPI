package models

import "time"

type UserVerbalStat struct {
	ID         int       `json:"id"`
	UserToken  string    `json:"user_token"`
	QuestionID int       `json:"question_id"`
	Correct    bool      `json:"correct"`
	Answers    []string  `json:"answers"`
	Date       time.Time `json:"time"`
}

type UserMarkedWord struct {
	ID        int    `json:"id"`
	UserToken string `json:"user_token"`
	WordID    int    `json:"word_id"`
}

type UserMarkedVerbalQuestion struct {
	ID               int    `json:"id"`
	UserToken        string `json:"user_token"`
	VerbalQuestionID int    `json:"verbal_question_id"`
}
