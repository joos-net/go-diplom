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

type PostHandler struct {
	postService *service.PostService
	eventLogger *eventlog.Logger
}

func NewPostHandler(postService *service.PostService, eventLogger *eventlog.Logger) *PostHandler {
	return &PostHandler{
		postService: postService,
		eventLogger: eventLogger,
	}
}

func (h *PostHandler) CreatePost(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r)
	if !ok {
		h.respondWithError(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	var req model.PostCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondWithError(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if err := req.Validate(); err != nil {
		h.respondWithError(w, err.Error(), http.StatusBadRequest)
		return
	}
	post, err := h.postService.CreatePost(r.Context(), &req, userID)
	if err != nil {
		h.respondWithError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if h.eventLogger != nil {
		h.eventLogger.Log(fmt.Sprintf("user %d created post %d", userID, post.ID))
	}

	h.respondWithJSON(w, http.StatusCreated, post)
}

func (h *PostHandler) GetPost(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		h.respondWithError(w, "invalid id", http.StatusBadRequest)
		return
	}
	post, err := h.postService.GetPost(r.Context(), id)
	if err != nil {
		if errors.Is(err, apperrors.ErrPostNotFound) {
			h.respondWithError(w, "post not found", http.StatusNotFound)
			return
		}
		h.respondWithError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	h.respondWithJSON(w, http.StatusOK, post)
}

func (h *PostHandler) GetAllPosts(w http.ResponseWriter, r *http.Request) {
	limit, offset := parseLimitOffset(r)
	posts, err := h.postService.GetAllPosts(r.Context(), limit, offset)
	if err != nil {
		h.respondWithError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	total, _ := h.postService.GetPostsCount(r.Context())
	h.respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"posts": posts, "total": total, "limit": limit, "offset": offset,
	})
}

func (h *PostHandler) UpdatePost(w http.ResponseWriter, r *http.Request) {
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
	var req model.PostUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondWithError(w, "invalid request body", http.StatusBadRequest)
		return
	}
	post, err := h.postService.UpdatePost(r.Context(), id, &req, userID)
	if err != nil {
		switch {
		case errors.Is(err, apperrors.ErrPostNotFound):
			h.respondWithError(w, "post not found", http.StatusNotFound)
		case errors.Is(err, apperrors.ErrForbidden):
			h.respondWithError(w, "forbidden", http.StatusForbidden)
		default:
			h.respondWithError(w, err.Error(), http.StatusBadRequest)
		}
		return
	}
	h.respondWithJSON(w, http.StatusOK, post)
}

func (h *PostHandler) DeletePost(w http.ResponseWriter, r *http.Request) {
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
	if err := h.postService.DeletePost(r.Context(), id, userID); err != nil {
		switch {
		case errors.Is(err, apperrors.ErrPostNotFound):
			h.respondWithError(w, "post not found", http.StatusNotFound)
		case errors.Is(err, apperrors.ErrForbidden):
			h.respondWithError(w, "forbidden", http.StatusForbidden)
		default:
			h.respondWithError(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *PostHandler) GetPostsByAuthor(w http.ResponseWriter, r *http.Request) {
	authorID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		h.respondWithError(w, "invalid author id", http.StatusBadRequest)
		return
	}
	limit, offset := parseLimitOffset(r)
	posts, err := h.postService.GetPostsByAuthor(r.Context(), authorID, limit, offset)
	if err != nil {
		h.respondWithError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	total, _ := h.postService.GetPostsCountByAuthor(r.Context(), authorID)
	h.respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"posts": posts, "total": total, "limit": limit, "offset": offset,
	})
}

func parseLimitOffset(r *http.Request) (int, int) {
	limit, offset := 10, 0
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 100 {
			limit = n
		}
	}
	if v := r.URL.Query().Get("offset"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			offset = n
		}
	}
	return limit, offset
}

func (h *PostHandler) respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(payload)
}

func (h *PostHandler) respondWithError(w http.ResponseWriter, message string, code int) {
	type ErrorResponse struct {
		Error string `json:"error"`
	}
	h.respondWithJSON(w, code, ErrorResponse{Error: message})
}
