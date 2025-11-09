package repository

import (
	"database/sql"

	"github.com/hajime110419/go-bbs-api/internal/models"
)

// PostRepository handles all database operations for posts.
// It provides an abstraction over the data access layer.
type PostRepository struct {
	db *sql.DB
}

// NewPostRepository creates a new PostRepository with the given database connection.
func NewPostRepository(db *sql.DB) *PostRepository {
	return &PostRepository{db: db}
}

// GetAll retrieves all posts from the database, ordered by creation time (descending).
// "rowid" is an implicit auto-incrementing column in SQLite. Ordering by it
// in descending order retrieves the most recent posts first.
func (r *PostRepository) GetAll() ([]models.Post, error) {
	rows, err := r.db.Query("SELECT id, title, content FROM posts ORDER BY rowid DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	posts := make([]models.Post, 0)
	for rows.Next() {
		var p models.Post
		if err := rows.Scan(&p.ID, &p.Title, &p.Content); err != nil {
			return nil, err
		}
		posts = append(posts, p)
	}

	return posts, rows.Err()
}

// Create inserts a new post into the database.
// It uses a prepared statement to prevent SQL injection vulnerabilities.
func (r *PostRepository) Create(post *models.Post) error {
	stmt, err := r.db.Prepare("INSERT INTO posts(id, title, content) VALUES(?, ?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(post.ID, post.Title, post.Content)
	return err
}
