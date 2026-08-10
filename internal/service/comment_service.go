package service

import (
	"platform/internal/model"
	"platform/internal/repository"
)

type CommentService struct {
	commentRepo repository.CommentRepository
	postRepo    repository.PostRepository
}

func NewCommentService(commentRepo repository.CommentRepository, postRepo repository.PostRepository) *CommentService {
	return &CommentService{
		commentRepo: commentRepo,
		postRepo:    postRepo,
	}
}

func (s *CommentService) Create(userID, postID int, req *model.CreateCommentRequest) (*model.Comment, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	post, err := s.postRepo.GetByID(postID)
	if err != nil {
		return nil, err
	}
	if post == nil {
		return nil, ErrPostNotFound
	}

	comment := &model.Comment{
		PostID:  postID,
		UserID:  userID,
		Content: req.Content,
	}

	if err := s.commentRepo.Create(comment); err != nil {
		return nil, err
	}

	return comment, nil
}

func (s *CommentService) GetByPostID(postID int) ([]model.Comment, error) {
	post, err := s.postRepo.GetByID(postID)
	if err != nil {
		return nil, err
	}
	if post == nil {
		return nil, ErrPostNotFound
	}

	return s.commentRepo.GetByPostID(postID)
}
