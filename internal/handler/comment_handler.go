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

type CommentHandler struct {
	commentService *service.CommentService
}

func NewCommentHandler(commentService *service.CommentService) *CommentHandler {
	return &CommentHandler{
		commentService: commentService,
	}
}

func (h *CommentHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID := context.GetUserID(r)
	if userID == 0 {
		response.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/api/posts/")
	parts := strings.Split(path, "/")
	postID, err := strconv.Atoi(parts[0])
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid post ID")
		return
	}

	var req model.CreateCommentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	comment, err := h.commentService.Create(userID, postID, &req)
	if err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	response.Success(w, http.StatusCreated, comment)
}

func (h *CommentHandler) GetByPostID(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/posts/")
	parts := strings.Split(path, "/")
	postID, err := strconv.Atoi(parts[0])
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid post ID")
		return
	}

	comments, err := h.commentService.GetByPostID(postID)
	if err != nil {
		response.Error(w, http.StatusNotFound, err.Error())
		return
	}

	response.Success(w, http.StatusOK, comments)
}
