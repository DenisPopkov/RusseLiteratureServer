package models

type Poet struct {
	ID     int64  `json:"id"`
	Name   string `json:"name"`
	Image  string `json:"image"`
	Clip   int64  `json:"clip"`
	IsFave bool   `json:"isFave"`
}
