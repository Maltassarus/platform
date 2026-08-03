package repository

import (
	"database/sql"
	"errors"
	"time"

	"platform/internal/model"

	"github.com/jmoiron/sqlx"
)

type commentRepository struct {
	db *sqlx.DB
}

func NewCommentRepository(db *sqlx.DB) CommentRepository {
	return &commentRepository{db: db}
}

func (r *commentRepository) Create(comment *model.Comment) error {
	query := `
		INSERT INTO comments (post_id, user_id, content, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`
	now := time.Now()
	err := r.db.QueryRow(query, comment.PostID, comment.UserID, comment.Content, now, now).Scan(&comment.ID)
	if err != nil {
		return err
	}
	comment.CreatedAt = now
	comment.UpdatedAt = now
	return nil
}

func (r *commentRepository) GetByPostID(postID int) ([]model.Comment, error) {
	var comments []model.Comment
	query := `
		SELECT id, post_id, user_id, content, created_at, updated_at 
		FROM comments 
		WHERE post_id = $1
		ORDER BY created_at DESC
	`
	err := r.db.Select(&comments, query, postID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}
	return comments, nil
}
