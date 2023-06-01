package models

type Meaning struct {
	Meaning  string   `json:"meaning"`
	Examples []string `json:"examples"`
	Type     string   `json:"type"`
	Synonyms []string `json:"synonyms"`
}

type Word struct {
	ID       int       `json:"id"`
	Word     string    `json:"word"`
	Meanings []Meaning `json:"meanings"`
}
