package models

type Answer struct {
	ID      int64  `json:"id"`
	Text    string `json:"text"`
	IsRight bool   `json:"isRight"`
}
