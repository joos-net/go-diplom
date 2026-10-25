package repository

import (
	"advanced-blog-management-system/internal/model"
	"context"
	"database/sql"
	"fmt"
)

type commentRepository struct {
	db *sql.DB
}

// NewCommentRepository создает новый репозиторий комментариев
func NewCommentRepository(db *sql.DB) CommentRepository {
	return &commentRepository{db: db}
}

func (r *commentRepository) Create(ctx context.Context, comment *model.Comment) error {
	query := `INSERT INTO comments (content, post_id, author_id) VALUES ($1,$2,$3)
	          RETURNING id, created_at, updated_at`
	return r.db.QueryRowContext(ctx, query,
		comment.Content, comment.PostID, comment.AuthorID).
		Scan(&comment.ID, &comment.CreatedAt, &comment.UpdatedAt)
}

func (r *commentRepository) GetByID(ctx context.Context, id int) (*model.Comment, error) {
	query := `SELECT id, content, post_id, author_id, created_at, updated_at
	          FROM comments WHERE id=$1`
	c := &model.Comment{}
	err := r.db.QueryRowContext(ctx, query, id).
		Scan(&c.ID, &c.Content, &c.PostID, &c.AuthorID, &c.CreatedAt, &c.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("GetByID: %w", err)
	}
	return c, nil
}

func (r *commentRepository) GetByPostID(ctx context.Context, postID int, limit, offset int) ([]*model.Comment, error) {
	query := `SELECT id, content, post_id, author_id, created_at, updated_at
	          FROM comments WHERE post_id=$1 ORDER BY created_at ASC LIMIT $2 OFFSET $3`
	rows, err := r.db.QueryContext(ctx, query, postID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var comments []*model.Comment
	for rows.Next() {
		c := &model.Comment{}
		if err := rows.Scan(&c.ID, &c.Content, &c.PostID, &c.AuthorID,
			&c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}
		comments = append(comments, c)
	}
	return comments, rows.Err()
}

func (r *commentRepository) GetCountByPostID(ctx context.Context, postID int) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM comments WHERE post_id=$1`, postID).Scan(&count)
	return count, err
}

func (r *commentRepository) Update(ctx context.Context, comment *model.Comment) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE comments SET content=$1, updated_at=NOW() WHERE id=$2`,
		comment.Content, comment.ID)
	return err
}

func (r *commentRepository) Delete(ctx context.Context, id int) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM comments WHERE id=$1`, id)
	return err
}
