package service

import (
	"advanced-blog-management-system/internal/apperrors"
	"advanced-blog-management-system/internal/model"
	"advanced-blog-management-system/internal/repository"
	"context"
	"fmt"
	"time"
)

type PostService struct {
	postRepo    repository.PostRepository
	userRepo    repository.UserRepository
	commentRepo repository.CommentRepository
}

func NewPostService(postRepo repository.PostRepository, userRepo repository.UserRepository, commentRepo repository.CommentRepository) *PostService {
	return &PostService{
		postRepo:    postRepo,
		userRepo:    userRepo,
		commentRepo: commentRepo,
	}
}

func (s *PostService) CreatePost(ctx context.Context, req *model.PostCreateRequest, authorID int) (*model.Post, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("validation: %w", err)
	}
	status := model.PostStatusPublished
	if req.PublishAt != nil && req.PublishAt.After(time.Now()) {
		status = model.PostStatusDraft
	}
	post := &model.Post{
		Title:     req.Title,
		Content:   req.Content,
		AuthorID:  authorID,
		Status:    status,
		PublishAt: req.PublishAt,
	}
	if err := s.postRepo.Create(ctx, post); err != nil {
		return nil, err
	}
	return post, nil
}

func (s *PostService) GetPost(ctx context.Context, id int) (*model.Post, error) {
	post, err := s.postRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if post == nil {
		return nil, apperrors.ErrPostNotFound
	}
	return post, nil
}

func (s *PostService) GetAllPosts(ctx context.Context, limit, offset int) ([]*model.Post, error) {
	return s.postRepo.GetAll(ctx, limit, offset)
}

func (s *PostService) GetPostsCount(ctx context.Context) (int, error) {
	return s.postRepo.GetTotalCount(ctx)
}

func (s *PostService) UpdatePost(ctx context.Context, id int, req *model.PostUpdateRequest, userID int) (*model.Post, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("validation: %w", err)
	}
	post, err := s.postRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if post == nil {
		return nil, apperrors.ErrPostNotFound
	}
	if !post.CanBeEditedBy(userID) {
		return nil, apperrors.ErrForbidden
	}
	post.Title = req.Title
	post.Content = req.Content
	if req.Status != "" {
		post.Status = req.Status
	}
	if req.PublishAt != nil {
		post.PublishAt = req.PublishAt
	}
	if err := s.postRepo.Update(ctx, post); err != nil {
		return nil, err
	}
	return post, nil
}

func (s *PostService) DeletePost(ctx context.Context, id int, userID int) error {
	post, err := s.postRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if post == nil {
		return apperrors.ErrPostNotFound
	}
	if !post.CanBeDeletedBy(userID) {
		return apperrors.ErrForbidden
	}
	return s.postRepo.Delete(ctx, id)
}

func (s *PostService) GetPostsByAuthor(ctx context.Context, authorID int, limit, offset int) ([]*model.Post, error) {
	return s.postRepo.GetByAuthorID(ctx, authorID, limit, offset)
}

func (s *PostService) GetPostsCountByAuthor(ctx context.Context, authorID int) (int, error) {
	return s.postRepo.GetTotalCountByAuthorID(ctx, authorID)
}
