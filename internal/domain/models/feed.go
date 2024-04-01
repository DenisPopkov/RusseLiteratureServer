package models

type Feed struct {
	ID       int64
	Authors  []Author
	Articles []Article
	Poets    []Poet
}
