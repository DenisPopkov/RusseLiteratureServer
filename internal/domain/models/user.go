package models

type User struct {
	ID       int64  `json:"id"`
	Phone    string `json:"phone"`
	PassHash []byte `json:"passHash"`
	Name     string `json:"name"`
	Image    string `json:"image"`
}
