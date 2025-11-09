package service

import (
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/hajime110419/go-bbs-api/internal/models"
	"github.com/hajime110419/go-bbs-api/internal/repository"
	"github.com/hajime110419/go-bbs-api/internal/utils"
)

type PostService struct {
	repo *repository.PostRepository
}

func NewPostService(db *sql.DB) *PostService {
	return &PostService{
		repo: repository.NewPostRepository(db),
	}
}

// GetAllPosts retrieves all posts, applying business rules
func (s *PostService) GetAllPosts() ([]models.Post, error) {
	return s.repo.GetAll()
}

// CreatePost handles business logic for creating a post
func (s *PostService) CreatePost(title, content string) (*models.Post, error) {
	post := &models.Post{
		ID:      uuid.New().String(),
		Title:   utils.Sanitize(title),
		Content: utils.Sanitize(content),
	}

	if post.Title == "" || post.Content == "" {
		return nil, ErrInvalidPost
	}

	if err := s.repo.Create(post); err != nil {
		return nil, err
	}

	return post, nil
}

var ErrInvalidPost = fmt.Errorf("invalid post: title and content required")
