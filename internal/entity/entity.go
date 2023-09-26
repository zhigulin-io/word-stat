package entity

type Stat struct {
	PostID int    `json:"postId"`
	Word   string `json:"word"`
	Count  int    `json:"count"`
}
