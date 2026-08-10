package service

import (
	"errors"
	"testing"
	"time"

	"platform/internal/model"
)

type mockPostRepository struct {
	posts   map[int]*model.Post
	nextID  int
	pending []model.Post
}

func newMockPostRepository() *mockPostRepository {
	return &mockPostRepository{
		posts:  make(map[int]*model.Post),
		nextID: 1,
	}
}

func (m *mockPostRepository) Create(post *model.Post) error {
	post.ID = m.nextID
	m.nextID++
	m.posts[post.ID] = post
	return nil
}

func (m *mockPostRepository) GetByID(id int) (*model.Post, error) {
	if p, ok := m.posts[id]; ok {
		copy := *p
		return &copy, nil
	}
	return nil, nil
}

func (m *mockPostRepository) Update(post *model.Post) error {
	if _, ok := m.posts[post.ID]; !ok {
		return errors.New("not found")
	}
	m.posts[post.ID] = post
	return nil
}

func (m *mockPostRepository) Delete(id int) error {
	delete(m.posts, id)
	return nil
}

func (m *mockPostRepository) GetAll(limit, offset int) ([]model.Post, error) {
	var res []model.Post
	for _, p := range m.posts {
		if p.Status == "published" {
			res = append(res, *p)
		}
	}
	return res, nil
}

func (m *mockPostRepository) GetPendingPosts() ([]model.Post, error) {
	return m.pending, nil
}

func (m *mockPostRepository) PublishPost(id int) error {
	if p, ok := m.posts[id]; ok {
		if p.Status == "draft" {
			p.Status = "published"
			return nil
		}
		return errors.New("already published or not draft")
	}
	return errors.New("not found")
}

func (m *mockPostRepository) setPendingPosts(posts []model.Post) {
	m.pending = posts
}

func TestPostService_Create(t *testing.T) {
	repo := newMockPostRepository()
	svc := NewPostService(repo)

	req := &model.CreatePostRequest{
		Title:   "Test Title",
		Content: "Test content (more than 10 chars)",
	}
	post, err := svc.Create(1, req)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if post.Status != "published" {
		t.Errorf("expected status 'published', got '%s'", post.Status)
	}
	if post.PublishAt != nil {
		t.Error("publish_at should be nil for immediate publish")
	}

	future := time.Now().Add(24 * time.Hour)
	req.PublishAt = &future
	post, err = svc.Create(2, req)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if post.Status != "draft" {
		t.Errorf("expected status 'draft', got '%s'", post.Status)
	}
	if post.PublishAt == nil || post.PublishAt.Unix() != future.Unix() {
		t.Errorf("publish_at not set correctly")
	}

	req.Title = "12"
	_, err = svc.Create(3, req)
	if err == nil {
		t.Error("expected validation error for short title")
	}
}

func TestPostService_GetByID_HideDraft(t *testing.T) {
	repo := newMockPostRepository()
	svc := NewPostService(repo)

	post := &model.Post{
		UserID:  1,
		Title:   "Draft",
		Content: "Draft content",
		Status:  "draft",
	}
	repo.Create(post)

	got, err := svc.GetByID(post.ID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if got == nil || got.ID != post.ID {
		t.Error("author should see draft")
	}
}

func TestPostService_Update_Delete(t *testing.T) {
	repo := newMockPostRepository()
	svc := NewPostService(repo)

	post := &model.Post{
		UserID:  1,
		Title:   "Original",
		Content: "Original content",
		Status:  "published",
	}
	repo.Create(post)

	updateReq := &model.UpdatePostRequest{
		Title:   "Updated",
		Content: "Updated content",
	}
	updated, err := svc.Update(1, post.ID, updateReq)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if updated.Title != "Updated" || updated.Content != "Updated content" {
		t.Error("post not updated correctly")
	}

	_, err = svc.Update(2, post.ID, updateReq)
	if err != ErrNotAuthor {
		t.Errorf("expected ErrNotAuthor, got %v", err)
	}

	_, err = svc.Update(1, 999, updateReq)
	if err != ErrPostNotFound {
		t.Errorf("expected ErrPostNotFound, got %v", err)
	}

	err = svc.Delete(1, post.ID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	deleted, _ := repo.GetByID(post.ID)
	if deleted != nil {
		t.Error("post should be deleted")
	}

	post2 := &model.Post{UserID: 3, Title: "t", Content: "content", Status: "published"}
	repo.Create(post2)
	err = svc.Delete(1, post2.ID)
	if err != ErrNotAuthor {
		t.Errorf("expected ErrNotAuthor, got %v", err)
	}
}

func TestPostService_PublishPending(t *testing.T) {
	repo := newMockPostRepository()
	svc := NewPostService(repo)

	past := time.Now().Add(-time.Hour)
	post := &model.Post{
		UserID:    1,
		Title:     "Pending",
		Content:   "Content",
		Status:    "draft",
		PublishAt: &past,
	}
	repo.Create(post)

	repo.setPendingPosts([]model.Post{*post})

	pending, err := svc.GetPendingPosts()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(pending) != 1 || pending[0].ID != post.ID {
		t.Error("pending posts not returned correctly")
	}

	err = svc.PublishPost(post.ID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	updated, _ := repo.GetByID(post.ID)
	if updated.Status != "published" {
		t.Errorf("expected status 'published', got '%s'", updated.Status)
	}

	err = svc.PublishPost(post.ID)
	if err == nil {
		t.Error("expected error on already published post")
	}
}
