package service

import (
	"errors"
	"testing"

	"platform/internal/model"
	"platform/pkg/auth"
)

type mockUserRepository struct {
	users  map[string]*model.User
	byID   map[int]*model.User
	nextID int
}

func newMockUserRepository() *mockUserRepository {
	return &mockUserRepository{
		users:  make(map[string]*model.User),
		byID:   make(map[int]*model.User),
		nextID: 1,
	}
}

func (m *mockUserRepository) Create(user *model.User) error {
	if _, exists := m.users[user.Email]; exists {
		return errors.New("duplicate email")
	}
	user.ID = m.nextID
	m.nextID++
	m.users[user.Email] = user
	m.byID[user.ID] = user
	return nil
}

func (m *mockUserRepository) GetByEmail(email string) (*model.User, error) {
	if u, ok := m.users[email]; ok {
		return u, nil
	}
	return nil, nil
}

func (m *mockUserRepository) GetByID(id int) (*model.User, error) {
	if u, ok := m.byID[id]; ok {
		return u, nil
	}
	return nil, nil
}

func TestUserService_Register(t *testing.T) {
	repo := newMockUserRepository()
	svc := NewUserService(repo)

	req := &model.RegisterRequest{
		Email:    "test@example.com",
		Password: "password123",
	}
	user, err := svc.Register(req)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if user.Email != req.Email {
		t.Errorf("expected email %s, got %s", req.Email, user.Email)
	}
	if user.Password == "" || user.Password == req.Password {
		t.Error("password should be hashed")
	}

	_, err = svc.Register(req)
	if err != ErrUserAlreadyExists {
		t.Errorf("expected ErrUserAlreadyExists, got %v", err)
	}

	invalidReq := &model.RegisterRequest{
		Email:    "invalid",
		Password: "password123",
	}
	_, err = svc.Register(invalidReq)
	if err == nil {
		t.Error("expected validation error for invalid email")
	}

	shortPass := &model.RegisterRequest{
		Email:    "short@example.com",
		Password: "123",
	}
	_, err = svc.Register(shortPass)
	if err == nil {
		t.Error("expected validation error for short password")
	}
}

func TestUserService_Login(t *testing.T) {
	repo := newMockUserRepository()
	svc := NewUserService(repo)

	hashed, _ := auth.HashPassword("secret")
	user := &model.User{
		Email:    "login@example.com",
		Password: hashed,
	}
	repo.Create(user)

	req := &model.LoginRequest{
		Email:    "login@example.com",
		Password: "secret",
	}
	loggedUser, err := svc.Login(req)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if loggedUser.Email != req.Email {
		t.Errorf("expected %s, got %s", req.Email, loggedUser.Email)
	}

	req.Password = "wrong"
	_, err = svc.Login(req)
	if err != ErrInvalidCredentials {
		t.Errorf("expected ErrInvalidCredentials, got %v", err)
	}

	req.Email = "unknown@example.com"
	_, err = svc.Login(req)
	if err != ErrUserNotFound {
		t.Errorf("expected ErrUserNotFound, got %v", err)
	}
}
