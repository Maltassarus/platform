package service

import (
	"testing"

	"platform/internal/model"
)

type mockPostRepositoryForComment struct {
	posts map[int]*model.Post
}

func (m *mockPostRepositoryForComment) GetByID(id int) (*model.Post, error) {
	if p, ok := m.posts[id]; ok {
		return p, nil
	}
	return nil, nil
}

func (m *mockPostRepositoryForComment) Create(*model.Post) error               { return nil }
func (m *mockPostRepositoryForComment) Update(*model.Post) error               { return nil }
func (m *mockPostRepositoryForComment) Delete(int) error                       { return nil }
func (m *mockPostRepositoryForComment) GetAll(int, int) ([]model.Post, error)  { return nil, nil }
func (m *mockPostRepositoryForComment) GetPendingPosts() ([]model.Post, error) { return nil, nil }
func (m *mockPostRepositoryForComment) PublishPost(int) error                  { return nil }

type mockCommentRepository struct {
	comments map[int][]model.Comment
	nextID   int
}

func newMockCommentRepository() *mockCommentRepository {
	return &mockCommentRepository{
		comments: make(map[int][]model.Comment),
		nextID:   1,
	}
}

func (m *mockCommentRepository) Create(comment *model.Comment) error {
	comment.ID = m.nextID
	m.nextID++
	m.comments[comment.PostID] = append(m.comments[comment.PostID], *comment)
	return nil
}

func (m *mockCommentRepository) GetByPostID(postID int) ([]model.Comment, error) {
	if comments, ok := m.comments[postID]; ok {
		return comments, nil
	}
	return []model.Comment{}, nil
}

func TestCommentService_Create(t *testing.T) {
	postRepo := &mockPostRepositoryForComment{posts: make(map[int]*model.Post)}
	commentRepo := newMockCommentRepository()
	svc := NewCommentService(commentRepo, postRepo)

	post := &model.Post{ID: 1, Title: "t", Content: "content", Status: "published"}
	postRepo.posts[1] = post

	req := &model.CreateCommentRequest{Content: "Nice post!"}
	comment, err := svc.Create(1, 1, req)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if comment.UserID != 1 || comment.PostID != 1 || comment.Content != "Nice post!" {
		t.Error("comment fields incorrect")
	}

	_, err = svc.Create(1, 999, req)
	if err != ErrPostNotFound {
		t.Errorf("expected ErrPostNotFound, got %v", err)
	}

	req.Content = ""
	_, err = svc.Create(1, 1, req)
	if err == nil {
		t.Error("expected validation error for empty comment")
	}
}

func TestCommentService_GetByPostID(t *testing.T) {
	postRepo := &mockPostRepositoryForComment{posts: make(map[int]*model.Post)}
	commentRepo := newMockCommentRepository()
	svc := NewCommentService(commentRepo, postRepo)

	post := &model.Post{ID: 1}
	postRepo.posts[1] = post
	commentRepo.Create(&model.Comment{PostID: 1, UserID: 1, Content: "c1"})
	commentRepo.Create(&model.Comment{PostID: 1, UserID: 2, Content: "c2"})

	comments, err := svc.GetByPostID(1)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(comments) != 2 {
		t.Errorf("expected 2 comments, got %d", len(comments))
	}

	_, err = svc.GetByPostID(999)
	if err != ErrPostNotFound {
		t.Errorf("expected ErrPostNotFound, got %v", err)
	}
}
