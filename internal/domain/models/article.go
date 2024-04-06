package models

type Article struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Image       string `json:"image"`
	Clip        int64  `json:"clip"`
	IsFave      bool   `json:"isFave"`
}
