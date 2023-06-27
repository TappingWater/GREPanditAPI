package models

type Meaning struct {
	Meaning string `json:"meaning"`
	Type    string `json:"type"`
}

type Word struct {
	ID       int       `json:"id"`
	Word     string    `json:"word"`
	Meanings []Meaning `json:"meanings"`
	Examples []string  `json:"examples"`
	Marked   bool      `json:"marked"`
}

type WordMap struct {
	BaseForm  string `json:"base_form"`
	Variation string `json:"variation"`
}

type MarkWordsReq struct {
	Words []string `json:"words"`
}
