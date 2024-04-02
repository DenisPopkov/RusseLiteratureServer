package models

type Quiz struct {
	ID          int64    `json:"id"`
	Question    string   `json:"question"`
	Description string   `json:"description"`
	Image       string   `json:"image"`
	Answers     []Answer `json:"answers"`
}
