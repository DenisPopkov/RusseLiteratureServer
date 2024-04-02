package models

type Feed struct {
	ID       int64     `json:"id"`
	Authors  []Author  `json:"authors"`
	Articles []Article `json:"articles"`
	Poets    []Poet    `json:"poets"`
}
