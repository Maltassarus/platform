package repository

import "platform/internal/model"

type UserRepository interface {
	Create(user *model.User) error
	GetByEmail(email string) (*model.User, error)
	GetByID(id int) (*model.User, error)
}

type PostRepository interface {
	Create(post *model.Post) error
	GetByID(id int) (*model.Post, error)
	Update(post *model.Post) error
	Delete(id int) error
	GetAll(limit, offset int) ([]model.Post, error)
	GetPendingPosts() ([]model.Post, error)
	PublishPost(id int) error
}

type CommentRepository interface {
	Create(comment *model.Comment) error
	GetByPostID(postID int) ([]model.Comment, error)
}
