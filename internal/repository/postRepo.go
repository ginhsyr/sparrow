package repository

import (
	"Sparrow/internal/model"
	"fmt"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"strconv"
)

type PostRepository struct {
	DB *gorm.DB
}

type PostQueryOptions struct {
	IncludeContent bool
	IncludeEdits   bool
	EditsLimit     int
}

func NewPostRepository(db *gorm.DB) *PostRepository {
	return &PostRepository{DB: db}
}

func (r *PostRepository) GetPostByID(id string, opts PostQueryOptions) (*model.Post, error) {
	var post model.Post
	postID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid post id: %w", err)
	}

	query := r.DB.Session(&gorm.Session{NewDB: true}).
		Model(&model.Post{}).
		Preload("Publisher").
		Where("posts.post_id = ?", postID)

	if opts.IncludeContent {
		query = query.Preload("Content")
	}

	if opts.IncludeEdits {
		if opts.EditsLimit > 0 {
			query = query.Preload("PostEdits", func(db *gorm.DB) *gorm.DB {
				return db.Order("edited_at DESC").Limit(opts.EditsLimit)
			})
		} else {
			query = query.Preload("PostEdits")
		}
		query = query.Preload("PostEdits.Editor")
	}

	if err := query.Take(&post).Error; err != nil {
		return nil, err
	}
	return &post, nil
}

//	func (r *PostRepository) CreatePost(post *model.Post) error {
//		return r.DB.Create(post).Error
//	}
func (r *PostRepository) CreatePostContent(content *model.PostContent) error {
	return r.DB.Session(&gorm.Session{NewDB: true}).Create(content).Error
}

func (r *PostRepository) CreatePost(post *model.Post) error {
	return r.DB.Session(&gorm.Session{NewDB: true}).
		Model(&model.Post{}).
		Omit("Publisher", "Content", "PostEdits", "Likers").
		Create(post).Error
}

func (r *PostRepository) CreatePostWithContent(post *model.Post, content *model.PostContent) error {
	return r.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&model.Post{}).
			Omit("Publisher", "Content", "PostEdits", "Likers").
			Create(post).Error; err != nil {
			return err
		}

		content.PostID = post.PostID
		if err := tx.Create(content).Error; err != nil {
			return err
		}
		return nil
	})
}

func (r *PostRepository) PostLike(postLike *model.PostLike) (bool, error) {
	inserted := false
	err := r.DB.Session(&gorm.Session{NewDB: true}).Transaction(func(tx *gorm.DB) error {
		result := tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "post_id"}, {Name: "user_id"}},
			DoNothing: true,
		}).Create(postLike)
		if result.Error != nil {
			return result.Error
		}

		if result.RowsAffected == 0 {
			return nil
		}
		inserted = true

		update := tx.Model(&model.Post{}).
			Where("post_id = ?", postLike.PostID).
			UpdateColumn("like_count", gorm.Expr("like_count + 1"))
		if update.Error != nil {
			return update.Error
		}
		if update.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}
		return nil
	})
	return inserted, err
}

func (r *PostRepository) PostExists(postID int64) (bool, error) {
	var count int64
	if err := r.DB.Session(&gorm.Session{NewDB: true}).
		Model(&model.Post{}).
		Where("post_id = ?", postID).
		Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *PostRepository) UserExists(userID int64) (bool, error) {
	var count int64
	if err := r.DB.Session(&gorm.Session{NewDB: true}).
		Model(&model.User{}).
		Where("id = ?", userID).
		Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}
