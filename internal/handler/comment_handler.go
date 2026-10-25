package handler

import (
	"advanced-blog-management-system/internal/apperrors"
	"advanced-blog-management-system/internal/eventlog"
	"advanced-blog-management-system/internal/middleware"
	"advanced-blog-management-system/internal/model"
	"advanced-blog-management-system/internal/service"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type CommentHandler struct {
	commentService *service.CommentService
	eventLogger    *eventlog.Logger
}

func NewCommentHandler(commentService *service.CommentService, eventLogger *eventlog.Logger) *CommentHandler {
	return &CommentHandler{
		commentService: commentService,
		eventLogger:    eventLogger,
	}
}

func (h *CommentHandler) CreateComment(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r)
	if !ok {
		h.respondWithError(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	postID, err := strconv.Atoi(chi.URLParam(r, "postId"))
	if err != nil {
		h.respondWithError(w, "invalid post id", http.StatusBadRequest)
		return
	}
	var req model.CommentCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondWithError(w, "invalid request body", http.StatusBadRequest)
		return
	}
	comment, err := h.commentService.CreateComment(r.Context(), &req, postID, userID)
	if err != nil {
		switch {
		case errors.Is(err, apperrors.ErrPostNotFound):
			h.respondWithError(w, "post not found", http.StatusNotFound)
		default:
			h.respondWithError(w, err.Error(), http.StatusBadRequest)
		}
		return
	}

	if h.eventLogger != nil {
		h.eventLogger.Log(fmt.Sprintf("user %d created comment %d", userID, comment.ID))
	}

	h.respondWithJSON(w, http.StatusCreated, comment)
}

func (h *CommentHandler) GetCommentsByPostID(w http.ResponseWriter, r *http.Request) {
	postID, err := strconv.Atoi(chi.URLParam(r, "postId"))
	if err != nil {
		h.respondWithError(w, "invalid post id", http.StatusBadRequest)
		return
	}
	limit, offset := parseLimitOffset(r)
	comments, err := h.commentService.GetCommentsByPostID(r.Context(), postID, limit, offset)
	if err != nil {
		if errors.Is(err, apperrors.ErrPostNotFound) {
			h.respondWithError(w, "post not found", http.StatusNotFound)
			return
		}
		h.respondWithError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	total, _ := h.commentService.GetCommentsCountByPostID(r.Context(), postID)
	h.respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"comments": comments, "total": total, "limit": limit, "offset": offset,
	})
}

func (h *CommentHandler) UpdateComment(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r)
	if !ok {
		h.respondWithError(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		h.respondWithError(w, "invalid id", http.StatusBadRequest)
		return
	}
	var req model.CommentUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondWithError(w, "invalid request body", http.StatusBadRequest)
		return
	}
	comment, err := h.commentService.UpdateComment(r.Context(), id, &req, userID)
	if err != nil {
		switch {
		case errors.Is(err, apperrors.ErrCommentNotFound):
			h.respondWithError(w, "comment not found", http.StatusNotFound)
		case errors.Is(err, apperrors.ErrForbidden):
			h.respondWithError(w, "forbidden", http.StatusForbidden)
		default:
			h.respondWithError(w, err.Error(), http.StatusBadRequest)
		}
		return
	}
	h.respondWithJSON(w, http.StatusOK, comment)
}

func (h *CommentHandler) DeleteComment(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r)
	if !ok {
		h.respondWithError(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		h.respondWithError(w, "invalid id", http.StatusBadRequest)
		return
	}
	if err := h.commentService.DeleteComment(r.Context(), id, userID); err != nil {
		switch {
		case errors.Is(err, apperrors.ErrCommentNotFound):
			h.respondWithError(w, "comment not found", http.StatusNotFound)
		case errors.Is(err, apperrors.ErrForbidden):
			h.respondWithError(w, "forbidden", http.StatusForbidden)
		default:
			h.respondWithError(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *CommentHandler) respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(payload)
}

func (h *CommentHandler) respondWithError(w http.ResponseWriter, message string, code int) {
	type ErrorResponse struct {
		Error string `json:"error"`
	}
	h.respondWithJSON(w, code, ErrorResponse{Error: message})
}
