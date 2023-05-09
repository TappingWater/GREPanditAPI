package models

type Word struct {
	ID       int      `json:"id"`
	Word     string   `json:"word"`
	Meanings []string `json:"meanings"`
	Examples []string `json:"examples"`
}
