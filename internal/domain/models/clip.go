package models

type Clip struct {
	ID    int64      `json:"id"`
	Text  []ClipText `json:"text"`
	Quiz  Quiz       `json:"quiz"`
	Image string     `json:"image"`
}
