package models

import (
	"database/sql"
	"encoding/json"
	"errors"
)

type Competence int
type FramedAs int
type QuestionType int
type Difficulty int

// ENUM types
const (
	Easy Difficulty = iota
	Medium
	Hard
)

const (
	AnalyzingAndDrawingConclusions Competence = iota
	ReasoningFromIncompleteData
	IdentifyingAuthorsAssumptionsPerspective
	UnderstandingMultipleLevelsOfMeaning
	SelectingImportantInfo
	DistinguishMajorMinorPoints
)

const (
	MCQSingleAnswer FramedAs = iota
	MCQMultipleChoices
	SelectSentence
)

const (
	ReadingComprehension QuestionType = iota
	TextCompletion
	SentenceEquivalence
)

// String equivalents for ENUM types
func (d Difficulty) String() string {
	switch d {
	case Easy:
		return "Easy"
	case Medium:
		return "Medium"
	case Hard:
		return "Hard"
	default:
		return "Unknown"
	}
}

func (c Competence) String() string {
	switch c {
	case AnalyzingAndDrawingConclusions:
		return "Analyzing and drawing conclusions"
	case ReasoningFromIncompleteData:
		return "Reasoning from incomplete data"
	case IdentifyingAuthorsAssumptionsPerspective:
		return "Identifying authors assumptions/perspective"
	case UnderstandingMultipleLevelsOfMeaning:
		return "Understanding multiple levels of meaning"
	case SelectingImportantInfo:
		return "Selecting important info"
	case DistinguishMajorMinorPoints:
		return "Distinguish major/minor points"
	default:
		return "Unknown"
	}
}

func (f FramedAs) String() string {
	switch f {
	case MCQSingleAnswer:
		return "MCQSingleAnswer"
	case MCQMultipleChoices:
		return "MCQMultipleChoice"
	case SelectSentence:
		return "SelectSentence"
	default:
		return "Unknown"
	}
}

func (q QuestionType) String() string {
	switch q {
	case ReadingComprehension:
		return "ReadingComprehension"
	case SentenceEquivalence:
		return "SentenceEquivalence"
	case TextCompletion:
		return "TextCompletion"
	default:
		return "Unknown"
	}
}

// Marshal JSON
func (d Difficulty) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.String())
}

func (c Competence) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.String())
}

func (f FramedAs) MarshalJSON() ([]byte, error) {
	return json.Marshal(f.String())
}

func (q QuestionType) MarshalJSON() ([]byte, error) {
	return json.Marshal(q.String())
}

// Unmarshal Json
func (v *VerbalQuestion) UnmarshalJSON(data []byte) error {
	type Alias VerbalQuestion
	aux := struct {
		ParagraphID *int64 `json:"paragraph_id"`
		*Alias
	}{
		Alias: (*Alias)(v),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	if aux.ParagraphID != nil {
		v.ParagraphID.Int64 = *aux.ParagraphID
		v.ParagraphID.Valid = true
	} else {
		v.ParagraphID.Valid = false
	}
	return nil
}

func (d *Difficulty) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	switch s {
	case "Easy":
		*d = Easy
	case "Medium":
		*d = Medium
	case "Hard":
		*d = Hard
	default:
		return errors.New("invalid difficulty value")
	}
	return nil
}

func (c *Competence) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	switch s {
	case "Analyzing and drawing conclusions":
		*c = AnalyzingAndDrawingConclusions
	case "Reasoning from incomplete data":
		*c = ReasoningFromIncompleteData
	case "Identifying authors assumptions/perspective":
		*c = IdentifyingAuthorsAssumptionsPerspective
	case "Understanding multiple levels of meaning":
		*c = UnderstandingMultipleLevelsOfMeaning
	case "Selecting important info":
		*c = SelectingImportantInfo
	case "Distinguish major/minor points":
		*c = DistinguishMajorMinorPoints
	default:
		return errors.New("invalid competence value")
	}
	return nil
}

func (f *FramedAs) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	switch s {
	case "MCQSingleAnswer":
		*f = MCQSingleAnswer
	case "MCQMultipleChoices":
		*f = MCQMultipleChoices
	case "SelectSentence":
		*f = SelectSentence
	default:
		return errors.New("invalid framed_as value")
	}
	return nil
}

func (q *QuestionType) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	switch s {
	case "ReadingComprehension":
		*q = ReadingComprehension
	case "TextCompletion":
		*q = TextCompletion
	case "SentenceEquivalence":
		*q = SentenceEquivalence
	default:
		return errors.New("invalid question type value")
	}
	return nil
}

/**
*   VerbalQuestion:
*     type: object
*     required:
*       - id
*       - competence
*       - framed_as
*       - type
*       - question
*       - options
*       - answer
*       - explanation
*       - difficulty
*     properties:
*       id:
*         type: integer
*         example: 1
*       competence:
*         type: string
*         enum: ["Analyzing and drawing conclusions", "Reasoning from incomplete data", "Identifying authors assumptions/perspective", "Understanding multiple levels of meaning", "Selecting important info", "Distinguish major/minor points"]
*         example: "Analyzing and drawing conclusions"
*       framed_as:
*         type: string
*         enum: ["MCQSingleAnswer", "MCQMultipleChoices", "SelectSentence"]
*         example: "MCQSingleAnswer"
*       type:
*         type: string
*         enum: ["ReadingComprehension", "TextCompletion", "SentenceEquivalence"]
*         example: "ReadingComprehension"
*       paragraph_id:
*         type: integer
*         example: 1
*       paragraph_text:
*         type: string
*         example: "This is a sample paragraph text."
*       question:
*         type: string
*         example: "What is the main idea of the paragraph?"
*       options:
*         type: array
*         items:
*           type: string
*         example: ["Option A", "Option B", "Option C"]
*       answer:
*         type: array
*         items:
*           type: string
*         example: ["Option A"]
*       explanation:
*         type: string
*         example: "Option A is correct because..."
*       difficulty:
*         type: string
*         enum: ["Easy", "Medium", "Hard"]
*         example: "Easy"
**/
type VerbalQuestion struct {
	ID            int            `json:"id"`
	Competence    Competence     `json:"competence"`
	FramedAs      FramedAs       `json:"framed_as"`
	Type          QuestionType   `json:"type"`
	ParagraphID   sql.NullInt64  `json:"paragraph_id,omitempty"`
	ParagraphText sql.NullString `json:"paragraph_text,omitempty"`
	Question      string         `json:"question"`
	Options       []string       `json:"options"`
	Answer        []string       `json:"answer"`
	Explanation   string         `json:"explanation"`
	Difficulty    Difficulty     `json:"difficulty"`
}
