package handler

import (
	"Sparrow/internal/service"
	"Sparrow/internal/utils"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"net/http"
	"strconv"
)

type PostHandler struct {
	service *service.PostService
}

func NewPostHandler(service *service.PostService) *PostHandler {
	return &PostHandler{service: service}
}

func (h *PostHandler) GetPost(c *gin.Context) {
	idStr := c.Param("id")
	opts, err := buildPostQueryOptions(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	post, err := h.service.GetPostByID(idStr, opts)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "post not found"})
			return
		}
		utils.Log.Error("get post err", zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": "get post err"})
		return
	}
	c.JSON(http.StatusOK, post)
}

func buildPostQueryOptions(c *gin.Context) (service.PostQueryOptions, error) {
	opts := service.PostQueryOptions{
		IncludeContent: true,
		IncludeEdits:   false,
		EditsLimit:     20,
	}

	if raw, ok := c.GetQuery("includeContent"); ok {
		v, err := strconv.ParseBool(raw)
		if err != nil {
			return opts, fmt.Errorf("includeContent must be true or false")
		}
		opts.IncludeContent = v
	}

	if raw, ok := c.GetQuery("includeEdits"); ok {
		v, err := strconv.ParseBool(raw)
		if err != nil {
			return opts, fmt.Errorf("includeEdits must be true or false")
		}
		opts.IncludeEdits = v
	}

	if raw, ok := c.GetQuery("editsLimit"); ok {
		limit, err := strconv.Atoi(raw)
		if err != nil || limit <= 0 || limit > 200 {
			return opts, fmt.Errorf("editsLimit must be between 1 and 200")
		}
		opts.EditsLimit = limit
	}

	if !opts.IncludeEdits {
		opts.EditsLimit = 0
	}

	return opts, nil
}

func (h *PostHandler) CreatePost(c *gin.Context) {
	userID := getUserID(c)
	if userID == 0 {
		return
	}
	var req createPostRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Log.Error("bind json err", zap.Error(err))
		respondValidationError(c, err, req)
		return
	}
	post, err := h.service.CreatePost(userID, req.Content)
	if err != nil {
		utils.Log.Error("create post err", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create post"})
		return
	}
	c.JSON(http.StatusCreated, post)
}

func (h *PostHandler) PostLike(c *gin.Context) {
	userID := getUserID(c)
	if userID == 0 {
		return
	}
	postID, err := parsePostIDForLike(c)
	if err != nil {
		utils.Log.Error("parse like postID err", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid postID"})
		return
	}

	postLike, err := h.service.PostLike(postID, userID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidPostID):
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid postID"})
		case errors.Is(err, service.ErrPostNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "post not found"})
		case errors.Is(err, service.ErrUserNotFound):
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token user"})
		case errors.Is(err, service.ErrPostAlreadyLiked):
			c.JSON(http.StatusOK, gin.H{
				"postID":  postID,
				"userID":  userID,
				"liked":   false,
				"message": "already liked",
			})
		default:
			utils.Log.Error("like post err", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to like post"})
		}
		return
	}
	c.JSON(http.StatusCreated, gin.H{
		"postID": postLike.PostID,
		"userID": postLike.UserID,
		"liked":  true,
	})
}

func parsePostIDForLike(c *gin.Context) (int64, error) {
	return strconv.ParseInt(c.Param("id"), 10, 64)
}

func getUserID(c *gin.Context) int64 {
	userIDAny, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "userID not found"})
		return 0
	}

	userID, err := toInt64(userIDAny)
	if err != nil {
		utils.Log.Error("invalid userID", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid userID"})
		return 0
	}
	return userID
}

func toInt64(value any) (int64, error) {
	switch v := value.(type) {
	case string:
		return strconv.ParseInt(v, 10, 64)
	case int:
		return int64(v), nil
	case int64:
		return v, nil
	case float64:
		return int64(v), nil
	default:
		return 0, fmt.Errorf("unsupported type: %T", value)
	}
}
