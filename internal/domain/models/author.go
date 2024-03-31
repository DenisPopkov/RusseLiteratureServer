package models

type Author struct {
	ID     int64
	Name   string
	Text   string
	Image  string
	Clip   int64
	IsFave bool
}
