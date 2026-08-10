package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"platform/internal/model"
	"platform/internal/service"
	"platform/pkg/context"
	"platform/pkg/response"
)

type PostHandler struct {
	postService *service.PostService
}

func NewPostHandler(postService *service.PostService) *PostHandler {
	return &PostHandler{
		postService: postService,
	}
}

func (h *PostHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID := context.GetUserID(r)
	if userID == 0 {
		response.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req model.CreatePostRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	post, err := h.postService.Create(userID, &req)
	if err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	response.Success(w, http.StatusCreated, post)
}

func (h *PostHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/posts/")
	id, err := strconv.Atoi(path)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid post ID")
		return
	}

	post, err := h.postService.GetByID(id)
	if err != nil {
		if err == service.ErrPostNotFound {
			response.Error(w, http.StatusNotFound, err.Error())
		} else {
			response.Error(w, http.StatusInternalServerError, "Failed to fetch post")
		}
		return
	}

	// Черновики видны только автору
	if post.Status == "draft" {
		userID := context.GetUserID(r)
		if userID == 0 || userID != post.UserID {
			response.Error(w, http.StatusNotFound, "post not found")
			return
		}
	}

	response.Success(w, http.StatusOK, post)
}

func (h *PostHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID := context.GetUserID(r)
	if userID == 0 {
		response.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/api/posts/")
	postID, err := strconv.Atoi(path)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid post ID")
		return
	}

	var req model.UpdatePostRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	post, err := h.postService.Update(userID, postID, &req)
	if err != nil {
		switch err {
		case service.ErrPostNotFound:
			response.Error(w, http.StatusNotFound, err.Error())
		case service.ErrNotAuthor:
			response.Error(w, http.StatusForbidden, err.Error())
		default:
			response.Error(w, http.StatusInternalServerError, "Failed to update post")
		}
		return
	}

	response.Success(w, http.StatusOK, post)
}

func (h *PostHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID := context.GetUserID(r)
	if userID == 0 {
		response.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/api/posts/")
	postID, err := strconv.Atoi(path)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid post ID")
		return
	}

	err = h.postService.Delete(userID, postID)
	if err != nil {
		switch err {
		case service.ErrPostNotFound:
			response.Error(w, http.StatusNotFound, err.Error())
		case service.ErrNotAuthor:
			response.Error(w, http.StatusForbidden, err.Error())
		default:
			response.Error(w, http.StatusInternalServerError, "Failed to delete post")
		}
		return
	}

	response.Success(w, http.StatusNoContent, nil)
}

func (h *PostHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	limit := 10
	offset := 0

	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	if o := r.URL.Query().Get("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	posts, err := h.postService.GetAll(limit, offset)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to fetch posts")
		return
	}

	response.Success(w, http.StatusOK, posts)
}
