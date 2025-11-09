package models

// Post represents a single entry on the bulletin board.
type Post struct {
	ID      string `json:"id"`
	Title   string `json:"title"`
	Content string `json:"content"`
}
