package models

type Competence int
type FramedAs int
type QuestionType int
type Difficulty int

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

type VerbalQuestion struct {
	VerbalQuestionID int          `json:"verbal_question_id"`
	Competence       Competence   `json:"competence"`
	FramedAs         FramedAs     `json:"framed_as"`
	Type             QuestionType `json:"type"`
	ParagraphID      int          `json:"paragraph_id"`
	Question         string       `json:"question"`
	Options          []string     `json:"options"`
	Answer           []string     `json:"answer"`
	Explanation      string       `json:"explanation"`
	Difficulty       Difficulty   `json:"difficulty"`
}
