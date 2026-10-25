package service

import (
	"advanced-blog-management-system/internal/apperrors"
	"advanced-blog-management-system/internal/model"
	"advanced-blog-management-system/internal/repository"
	"context"
	"fmt"
)

type CommentService struct {
	commentRepo repository.CommentRepository
	postRepo    repository.PostRepository
	userRepo    repository.UserRepository
}

func NewCommentService(commentRepo repository.CommentRepository, postRepo repository.PostRepository, userRepo repository.UserRepository) *CommentService {
	return &CommentService{
		commentRepo: commentRepo,
		postRepo:    postRepo,
		userRepo:    userRepo,
	}
}

func (s *CommentService) CreateComment(ctx context.Context, req *model.CommentCreateRequest, postID int, authorID int) (*model.Comment, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("validation: %w", err)
	}
	post, err := s.postRepo.GetByID(ctx, postID)
	if err != nil {
		return nil, err
	}
	if post == nil {
		return nil, apperrors.ErrPostNotFound
	}
	comment := &model.Comment{
		Content:  req.Content,
		PostID:   postID,
		AuthorID: authorID,
	}
	if err := s.commentRepo.Create(ctx, comment); err != nil {
		return nil, err
	}
	return comment, nil
}

func (s *CommentService) GetComment(ctx context.Context, id int) (*model.Comment, error) {
	c, err := s.commentRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if c == nil {
		return nil, apperrors.ErrCommentNotFound
	}
	return c, nil
}

func (s *CommentService) GetCommentsByPostID(ctx context.Context, postID int, limit, offset int) ([]*model.Comment, error) {
	exists, err := s.postRepo.Exists(ctx, postID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, apperrors.ErrPostNotFound
	}
	return s.commentRepo.GetByPostID(ctx, postID, limit, offset)
}

func (s *CommentService) GetCommentsCountByPostID(ctx context.Context, postID int) (int, error) {
	return s.commentRepo.GetCountByPostID(ctx, postID)
}

func (s *CommentService) UpdateComment(ctx context.Context, id int, req *model.CommentUpdateRequest, userID int) (*model.Comment, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("validation: %w", err)
	}
	comment, err := s.commentRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if comment == nil {
		return nil, apperrors.ErrCommentNotFound
	}
	if !comment.CanBeEditedBy(userID) {
		return nil, apperrors.ErrForbidden
	}
	comment.Content = req.Content
	if err := s.commentRepo.Update(ctx, comment); err != nil {
		return nil, err
	}
	return comment, nil
}

func (s *CommentService) DeleteComment(ctx context.Context, id int, userID int) error {
	comment, err := s.commentRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if comment == nil {
		return apperrors.ErrCommentNotFound
	}
	if !comment.CanBeDeletedBy(userID) {
		return apperrors.ErrForbidden
	}
	return s.commentRepo.Delete(ctx, id)
}
