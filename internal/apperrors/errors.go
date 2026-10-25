package apperrors

import "errors"

var (
	ErrUnauthorized       = errors.New("unauthorized")
	ErrForbidden          = errors.New("forbidden")
	ErrUserNotFound       = errors.New("user not found")
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrPostNotFound       = errors.New("post not found")
	ErrCommentNotFound    = errors.New("comment not found")
	ErrInvalidPostID      = errors.New("invalid post id")
)
