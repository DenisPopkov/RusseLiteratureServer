package models

type Author struct {
	ID     int64
	Name   string
	Image  string
	Clip   Clip
	IsFave string
}
