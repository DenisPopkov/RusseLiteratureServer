package models

type Quiz struct {
	ID          int64
	Question    string
	Answers     []Answer
	Description string
	Image       string
}
