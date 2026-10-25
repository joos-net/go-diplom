package repository

import (
	"advanced-blog-management-system/internal/model"
	"context"
	"database/sql"
	"fmt"
)

type postRepository struct {
	db *sql.DB
}

// NewPostRepository создает новый репозиторий постов
func NewPostRepository(db *sql.DB) PostRepository {
	return &postRepository{db: db}
}

func (r *postRepository) Create(ctx context.Context, post *model.Post) error {
	query := `INSERT INTO posts (title, content, author_id, status, publish_at)
	          VALUES ($1,$2,$3,$4,$5) RETURNING id, created_at, updated_at`
	return r.db.QueryRowContext(ctx, query,
		post.Title, post.Content, post.AuthorID, post.Status, post.PublishAt).
		Scan(&post.ID, &post.CreatedAt, &post.UpdatedAt)
}

func (r *postRepository) GetByID(ctx context.Context, id int) (*model.Post, error) {
	query := `SELECT id, title, content, author_id, status, publish_at, created_at, updated_at
	          FROM posts WHERE id=$1`
	p := &model.Post{}
	err := r.db.QueryRowContext(ctx, query, id).
		Scan(&p.ID, &p.Title, &p.Content, &p.AuthorID, &p.Status, &p.PublishAt, &p.CreatedAt, &p.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("GetByID: %w", err)
	}
	return p, nil
}

func (r *postRepository) GetAll(ctx context.Context, limit, offset int) ([]*model.Post, error) {
	query := `SELECT id, title, content, author_id, status, publish_at, created_at, updated_at
	          FROM posts ORDER BY created_at DESC LIMIT $1 OFFSET $2`
	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []*model.Post
	for rows.Next() {
		p := &model.Post{}
		if err := rows.Scan(&p.ID, &p.Title, &p.Content, &p.AuthorID, &p.Status,
			&p.PublishAt, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		posts = append(posts, p)
	}
	return posts, rows.Err()
}

func (r *postRepository) GetTotalCount(ctx context.Context) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM posts`).Scan(&count)
	return count, err
}

func (r *postRepository) Update(ctx context.Context, post *model.Post) error {
	query := `UPDATE posts SET title=$1, content=$2, status=$3, publish_at=$4, updated_at=NOW()
	          WHERE id=$5`
	_, err := r.db.ExecContext(ctx, query,
		post.Title, post.Content, post.Status, post.PublishAt, post.ID)
	return err
}

func (r *postRepository) Delete(ctx context.Context, id int) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM posts WHERE id=$1`, id)
	return err
}

func (r *postRepository) Exists(ctx context.Context, id int) (bool, error) {
	var exists bool
	err := r.db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM posts WHERE id=$1)`, id).Scan(&exists)
	return exists, err
}

func (r *postRepository) GetByAuthorID(ctx context.Context, authorID int, limit, offset int) ([]*model.Post, error) {
	query := `SELECT id, title, content, author_id, status, publish_at, created_at, updated_at
	          FROM posts WHERE author_id=$1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`
	rows, err := r.db.QueryContext(ctx, query, authorID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []*model.Post
	for rows.Next() {
		p := &model.Post{}
		if err := rows.Scan(&p.ID, &p.Title, &p.Content, &p.AuthorID, &p.Status,
			&p.PublishAt, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		posts = append(posts, p)
	}
	return posts, rows.Err()
}

func (r *postRepository) GetTotalCountByAuthorID(ctx context.Context, authorID int) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM posts WHERE author_id=$1`, authorID).Scan(&count)
	return count, err
}

func (r *postRepository) GetScheduledPosts(ctx context.Context) ([]*model.Post, error) {
	query := `SELECT id, title, content, author_id, status, publish_at, created_at, updated_at
	          FROM posts WHERE status='draft' AND publish_at IS NOT NULL AND publish_at <= NOW()`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var posts []*model.Post
	for rows.Next() {
		p := &model.Post{}
		if err := rows.Scan(&p.ID, &p.Title, &p.Content, &p.AuthorID, &p.Status,
			&p.PublishAt, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		posts = append(posts, p)
	}
	return posts, rows.Err()
}

func (r *postRepository) PublishPost(ctx context.Context, id int) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE posts SET status='published', publish_at=NULL, updated_at=NOW() WHERE id=$1`, id)
	return err
}
