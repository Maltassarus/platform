package service

import (
	"errors"
	"time"

	"platform/internal/model"
	"platform/internal/repository"
)

type PostService struct {
	postRepo repository.PostRepository
}

func NewPostService(postRepo repository.PostRepository) *PostService {
	return &PostService{
		postRepo: postRepo,
	}
}

func (s *PostService) Create(userID int, req *model.CreatePostRequest) (*model.Post, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	status := "published"
	if req.PublishAt != nil && req.PublishAt.After(time.Now()) {
		status = "draft"
	}

	post := &model.Post{
		UserID:    userID,
		Title:     req.Title,
		Content:   req.Content,
		Status:    status,
		PublishAt: req.PublishAt,
	}

	if err := s.postRepo.Create(post); err != nil {
		return nil, err
	}

	return post, nil
}

func (s *PostService) GetByID(id int) (*model.Post, error) {
	post, err := s.postRepo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if post == nil {
		return nil, errors.New("post not found")
	}
	return post, nil
}

func (s *PostService) Update(userID, postID int, req *model.UpdatePostRequest) (*model.Post, error) {
	post, err := s.postRepo.GetByID(postID)
	if err != nil {
		return nil, err
	}
	if post == nil {
		return nil, errors.New("post not found")
	}

	if post.UserID != userID {
		return nil, errors.New("you are not the author of this post")
	}

	if req.Title != "" {
		post.Title = req.Title
	}
	if req.Content != "" {
		post.Content = req.Content
	}

	if err := s.postRepo.Update(post); err != nil {
		return nil, err
	}

	return post, nil
}

func (s *PostService) Delete(userID, postID int) error {
	post, err := s.postRepo.GetByID(postID)
	if err != nil {
		return err
	}
	if post == nil {
		return errors.New("post not found")
	}

	if post.UserID != userID {
		return errors.New("you are not the author of this post")
	}

	return s.postRepo.Delete(postID)
}

func (s *PostService) GetAll(limit, offset int) ([]model.Post, error) {
	if limit <= 0 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}
	return s.postRepo.GetAll(limit, offset)
}

func (s *PostService) GetPendingPosts() ([]model.Post, error) {
	return s.postRepo.GetPendingPosts()
}

func (s *PostService) PublishPost(id int) error {
	return s.postRepo.PublishPost(id)
}
