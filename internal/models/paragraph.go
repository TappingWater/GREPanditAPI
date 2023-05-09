// models/paragraph.go
package models

/**
* Represents a paragraph or set of paragraphs.
* Since multiple questions can share the same paragraph
* or paragraph set this model aims to reduce data
* redundancy.
**/
type Paragraph struct {
	ID   int    `json:"id"`
	Text string `json:"text"`
}
