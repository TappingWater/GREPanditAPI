package models

import (
	"encoding/json"
	"errors"
)

type Competence int
type FramedAs int
type QuestionType int
type Difficulty int

// ENUM types
const (
	Easy Difficulty = iota + 1
	Medium
	Hard
)

const (
	AnalyzingAndDrawingConclusions Competence = iota + 1
	ReasoningFromIncompleteData
	IdentifyingAuthorsAssumptionsPerspective
	UnderstandingMultipleLevelsOfMeaning
	SelectingImportantInfo
	DistinguishMajorMinorPoints
)

const (
	MCQSingleAnswer FramedAs = iota + 1
	MCQMultipleChoices
	SelectSentence
)

const (
	ReadingComprehension QuestionType = iota + 1
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
		return "Identifying author's assumptions/perspective"
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
	case "MCQMultipleChoice":
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
* Model that is used to represent a single option that the
* user can select
**/
type Option struct {
	Value         string `json:"value"`
	Correct       bool   `json:"correct"`
	Justification string `json:"justification"`
}

/**
* Model that represents a question in the verbal reading portion
* of the GRE exam.
**/
type VerbalQuestion struct {
	ID           int               `json:"id"`
	Competence   Competence        `json:"competence"`
	FramedAs     FramedAs          `json:"framed_as"`
	Type         QuestionType      `json:"type"`
	Paragraph    string            `json:"paragraph"`
	Question     string            `json:"question"`
	Options      []Option          `json:"options"`
	Difficulty   Difficulty        `json:"difficulty"`
	Vocabulary   []Word            `json:"vocabulary"`
	VocabWordMap map[string]string `json:"wordmap"`
}

/**
* Represents the Data used by the POST endpoint.
* Vocabulary is passed as a list of words here.
**/
type VerbalQuestionRequest struct {
	ID         int          `json:"id"`
	Competence Competence   `json:"competence"`
	FramedAs   FramedAs     `json:"framed_as"`
	Type       QuestionType `json:"type"`
	Paragraph  string       `json:"paragraph"`
	Question   string       `json:"question"`
	Options    []Option     `json:"options"`
	Difficulty Difficulty   `json:"difficulty"`
	Vocabulary []string     `json:"vocabulary"`
}

type RandomQuestionsRequest struct {
	Limit        int          `json:"limit"`
	QuestionType QuestionType `json:"type,omitempty"`
	Competence   Competence   `json:"competence,omitempty"`
	FramedAs     FramedAs     `json:"framed_as,omitempty"`
	Difficulty   Difficulty   `json:"difficulty,omitempty"`
	ExcludeIDs   []int        `json:"exclude_ids,omitempty"`
}
