package service

import "errors"

var (
	ErrUserNotFound       = errors.New("user not found")
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrPostNotFound       = errors.New("post not found")
	ErrNotAuthor          = errors.New("you are not the author of this post")
)
