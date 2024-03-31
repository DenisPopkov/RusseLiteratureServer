package models

type Quiz struct {
	ID          int64
	Question    string
	Answers     []int64
	Description string
	Image       string
}
