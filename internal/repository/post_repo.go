package repository

import (
	"database/sql"
	"errors"
	"time"

	"platform/internal/model"

	"github.com/jmoiron/sqlx"
)

type postRepository struct {
	db *sqlx.DB
}

func NewPostRepository(db *sqlx.DB) PostRepository {
	return &postRepository{db: db}
}

func (r *postRepository) Create(post *model.Post) error {
	query := `
		INSERT INTO posts (user_id, title, content, status, publish_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
	`
	now := time.Now()
	err := r.db.QueryRow(query, post.UserID, post.Title, post.Content, post.Status, post.PublishAt, now, now).Scan(&post.ID)
	if err != nil {
		return err
	}
	post.CreatedAt = now
	post.UpdatedAt = now
	return nil
}

func (r *postRepository) GetByID(id int) (*model.Post, error) {
	var post model.Post
	query := `SELECT id, user_id, title, content, status, publish_at, created_at, updated_at FROM posts WHERE id = $1`
	err := r.db.Get(&post, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &post, nil
}

func (r *postRepository) Update(post *model.Post) error {
	query := `
		UPDATE posts 
		SET title = $1, content = $2, status = $3, updated_at = $4
		WHERE id = $5
	`
	now := time.Now()
	_, err := r.db.Exec(query, post.Title, post.Content, post.Status, now, post.ID)
	if err != nil {
		return err
	}
	post.UpdatedAt = now
	return nil
}

func (r *postRepository) Delete(id int) error {
	query := `DELETE FROM posts WHERE id = $1`
	_, err := r.db.Exec(query, id)
	return err
}

func (r *postRepository) GetAll(limit, offset int) ([]model.Post, error) {
	var posts []model.Post
	query := `
		SELECT id, user_id, title, content, status, publish_at, created_at, updated_at 
		FROM posts 
		WHERE status = 'published'
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`
	err := r.db.Select(&posts, query, limit, offset)
	return posts, err
}

func (r *postRepository) GetPendingPosts() ([]model.Post, error) {
	var posts []model.Post
	query := `
		SELECT id, user_id, title, content, status, publish_at, created_at, updated_at 
		FROM posts 
		WHERE status = 'draft' 
		AND publish_at IS NOT NULL 
		AND publish_at <= NOW()
		ORDER BY publish_at ASC
	`
	err := r.db.Select(&posts, query)
	return posts, err
}

func (r *postRepository) PublishPost(id int) error {
	query := `
		UPDATE posts 
		SET status = 'published', updated_at = NOW()
		WHERE id = $1 AND status = 'draft'
	`
	result, err := r.db.Exec(query, id)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return errors.New("post not found or already published")
	}
	return nil
}
