package models

type User struct {
	ID          int64
	PhoneNumber string
	PassHash    []byte
}
