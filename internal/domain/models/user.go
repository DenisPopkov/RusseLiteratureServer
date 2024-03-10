package models

type User struct {
	ID       int64
	Phone    string
	PassHash []byte
}
