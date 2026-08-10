package scheduler

import (
	"testing"
	"time"

	"platform/internal/model"
	"platform/internal/service"
)

type mockPostRepositoryForScheduler struct {
	posts   map[int]*model.Post
	nextID  int
	pending []model.Post
}

func newMockPostRepositoryForScheduler() *mockPostRepositoryForScheduler {
	return &mockPostRepositoryForScheduler{
		posts:  make(map[int]*model.Post),
		nextID: 1,
	}
}

func (m *mockPostRepositoryForScheduler) Create(post *model.Post) error {
	post.ID = m.nextID
	m.nextID++
	m.posts[post.ID] = post
	return nil
}

func (m *mockPostRepositoryForScheduler) GetByID(id int) (*model.Post, error) {
	if p, ok := m.posts[id]; ok {
		copy := *p
		return &copy, nil
	}
	return nil, nil
}

func (m *mockPostRepositoryForScheduler) Update(post *model.Post) error {
	if _, ok := m.posts[post.ID]; !ok {
		return nil
	}
	m.posts[post.ID] = post
	return nil
}

func (m *mockPostRepositoryForScheduler) Delete(id int) error {
	delete(m.posts, id)
	return nil
}

func (m *mockPostRepositoryForScheduler) GetAll(limit, offset int) ([]model.Post, error) {
	var res []model.Post
	for _, p := range m.posts {
		if p.Status == "published" {
			res = append(res, *p)
		}
	}
	return res, nil
}

func (m *mockPostRepositoryForScheduler) GetPendingPosts() ([]model.Post, error) {
	return m.pending, nil
}

func (m *mockPostRepositoryForScheduler) PublishPost(id int) error {
	if p, ok := m.posts[id]; ok {
		if p.Status == "draft" {
			p.Status = "published"
			return nil
		}
		return nil
	}
	return nil
}

func (m *mockPostRepositoryForScheduler) setPendingPosts(posts []model.Post) {
	m.pending = posts
}

func TestScheduler_PublishPendingPost(t *testing.T) {
	repo := newMockPostRepositoryForScheduler()
	postService := service.NewPostService(repo)

	pastTime := time.Now().Add(-time.Hour)
	post := &model.Post{
		UserID:    1,
		Title:     "Тестовый отложенный пост",
		Content:   "Пост должен быть опубликован планировщиком",
		Status:    "draft",
		PublishAt: &pastTime,
	}
	repo.Create(post)

	repo.setPendingPosts([]model.Post{*post})

	got, err := postService.GetByID(post.ID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if got.Status != "draft" {
		t.Errorf("expected status 'draft', got '%s'", got.Status)
	}

	err = postService.PublishPost(post.ID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	published, err := postService.GetByID(post.ID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if published.Status != "published" {
		t.Errorf("expected status 'published', got '%s'", published.Status)
	}

	t.Log("Отложенный пост опубликован")
}

func TestScheduler_GetPendingPosts(t *testing.T) {
	repo := newMockPostRepositoryForScheduler()
	postService := service.NewPostService(repo)

	now := time.Now()

	pastTime := now.Add(-time.Hour)
	post1 := &model.Post{
		UserID:    1,
		Title:     "Готов к публикации",
		Content:   "Пост готов к публикации",
		Status:    "draft",
		PublishAt: &pastTime,
	}
	repo.Create(post1)

	futureTime := now.Add(time.Hour)
	post2 := &model.Post{
		UserID:    1,
		Title:     "Ещё не готов",
		Content:   "Пост ещё не готов к публикации",
		Status:    "draft",
		PublishAt: &futureTime,
	}
	repo.Create(post2)

	post3 := &model.Post{
		UserID:  1,
		Title:   "Уже опубликован",
		Content: "Пост уже опубликован",
		Status:  "published",
	}
	repo.Create(post3)

	repo.setPendingPosts([]model.Post{*post1})

	pending, err := postService.GetPendingPosts()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(pending) != 1 {
		t.Errorf("expected 1 pending post, got %d", len(pending))
	}

	if len(pending) > 0 && pending[0].ID != post1.ID {
		t.Errorf("expected post ID %d, got %d", post1.ID, pending[0].ID)
	}

	err = postService.PublishPost(post1.ID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	updated, _ := postService.GetByID(post1.ID)
	if updated.Status != "published" {
		t.Errorf("expected status 'published', got '%s'", updated.Status)
	}

	t.Log("Тест GetPendingPosts пройден")
}

func TestScheduler_DoNotPublishFuturePosts(t *testing.T) {
	repo := newMockPostRepositoryForScheduler()
	postService := service.NewPostService(repo)

	futureTime := time.Now().Add(24 * time.Hour)
	post := &model.Post{
		UserID:    1,
		Title:     "Будущий пост",
		Content:   "Пост ещё не должен публиковаться",
		Status:    "draft",
		PublishAt: &futureTime,
	}
	repo.Create(post)

	repo.setPendingPosts([]model.Post{})

	pending, err := postService.GetPendingPosts()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(pending) != 0 {
		t.Errorf("expected 0 pending posts, got %d", len(pending))
	}

	got, err := postService.GetByID(post.ID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if got.Status != "draft" {
		t.Errorf("expected status 'draft', got '%s'", got.Status)
	}

	t.Log("Пост с будущим временем не опубликован")
}
