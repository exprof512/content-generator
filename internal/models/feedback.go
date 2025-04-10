package models

type Feedback struct {
	Score   int    `json:"score"`
	Content string `json:"content"`
}
