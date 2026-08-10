package service

import (
	"platform/internal/model"
	"platform/internal/repository"
	"platform/pkg/auth"
)

type UserService struct {
	userRepo repository.UserRepository
}

func NewUserService(userRepo repository.UserRepository) *UserService {
	return &UserService{
		userRepo: userRepo,
	}
}

func (s *UserService) Register(req *model.RegisterRequest) (*model.User, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	existingUser, err := s.userRepo.GetByEmail(req.Email)
	if err != nil {
		return nil, err
	}
	if existingUser != nil {
		return nil, ErrUserAlreadyExists
	}

	hashedPassword, err := auth.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	user := &model.User{
		Email:    req.Email,
		Password: hashedPassword,
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *UserService) Login(req *model.LoginRequest) (*model.User, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	user, err := s.userRepo.GetByEmail(req.Email)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}

	if err := auth.CheckPassword(req.Password, user.Password); err != nil {
		return nil, ErrInvalidCredentials
	}

	return user, nil
}
