package service

import (
	"Sparrow/internal/model"
	"Sparrow/internal/repository"
	"errors"
	"gorm.io/gorm"
)

var (
	ErrInvalidPostID    = errors.New("invalid post id")
	ErrPostNotFound     = errors.New("post not found")
	ErrUserNotFound     = errors.New("user not found")
	ErrPostAlreadyLiked = errors.New("post already liked")
)

type PostService struct {
	repo *repository.PostRepository
}

type PostQueryOptions struct {
	IncludeContent bool
	IncludeEdits   bool
	EditsLimit     int
}

func NewPostService(repo *repository.PostRepository) *PostService {
	return &PostService{repo: repo}
}

func (s *PostService) GetPostByID(id string, opts PostQueryOptions) (*model.Post, error) {
	return s.repo.GetPostByID(id, repository.PostQueryOptions{
		IncludeContent: opts.IncludeContent,
		IncludeEdits:   opts.IncludeEdits,
		EditsLimit:     opts.EditsLimit,
	})
}

//	func (s *PostService) CreatePost(userID int64, content string) (*model.Post, error) {
//		contentS := model.PostContent{
//			Content: content,
//		}
//		post := &model.Post{
//			PublisherID: userID,
//			Content:     contentS,
//		}
//		err := s.repo.CreatePost(post)
//		return post, err
//	}
func (s *PostService) CreatePost(userID int64, content string) (*model.Post, error) {
	post := &model.Post{
		PublisherID: userID,
	}
	contentObj := &model.PostContent{
		Content: content,
	}
	if err := s.repo.CreatePostWithContent(post, contentObj); err != nil {
		return nil, err
	}

	post.Content = contentObj
	return post, nil
}
func (s *PostService) PostLike(postID, userID int64) (*model.PostLike, error) {
	if postID <= 0 {
		return nil, ErrInvalidPostID
	}

	postExists, err := s.repo.PostExists(postID)
	if err != nil {
		return nil, err
	}
	if !postExists {
		return nil, ErrPostNotFound
	}

	userExists, err := s.repo.UserExists(userID)
	if err != nil {
		return nil, err
	}
	if !userExists {
		return nil, ErrUserNotFound
	}

	postLike := &model.PostLike{
		PostID: postID,
		UserID: userID,
	}
	inserted, err := s.repo.PostLike(postLike)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrPostNotFound
		}
		return nil, err
	}
	if !inserted {
		return nil, ErrPostAlreadyLiked
	}
	return postLike, nil
}
