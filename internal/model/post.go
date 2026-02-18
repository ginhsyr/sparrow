package model

import "time"

type PostEdit struct {
	ID            int64     `gorm:"primaryKey;column:id;type:bigint;autoIncrement"`
	PostID        int64     `json:"postID" gorm:"column:post_id;index;type:bigint"`
	EditorID      int       `json:"editorID" gorm:"column:editor_id;type:integer"`
	Editor        User      `gorm:"foreignKey:EditorID;references:ID"`
	EditedContent string    `json:"editedContent" gorm:"type:text;column:edited_content"`
	EditedAt      time.Time `json:"editedAt" gorm:"column:edited_at;autoCreateTime"`
}

type Post struct {
    PostID      int64     `json:"postID" gorm:"primaryKey;column:post_id;type:bigint;autoIncrement;unique"`
    PublisherID int64     `json:"publisherID" gorm:"column:publisher_id"`
    Publisher   User      `gorm:"foreignKey:PublisherID;references:ID"`
    UpdatedAt   time.Time `json:"updatedAt" gorm:"autoUpdateTime"`
    CreatedAt   time.Time `json:"createdAt" gorm:"autoCreateTime"`
    IsEdited    bool      `json:"isEdited" gorm:"default:false"`

    Content    *PostContent `json:"content" gorm:"foreignKey:PostID;references:PostID"`
    PostEdits  []PostEdit   `json:"edits" gorm:"foreignKey:PostID;references:PostID"`
    LikeCount int64        `json:"likeCount" gorm:"column:like_count;type:bigint;default:0"`
    Likers    []PostLike   `json:"likers"    gorm:"foreignKey:PostID;references:PostID"`
}

type PostContent struct {
	PostID  int64  `json:"postID" gorm:"primaryKey;column:post_id"`
	Content string `json:"content" gorm:"column:content"`
}

type PostLike struct {
	PostID  int64     `json:"postID" gorm:"primaryKey;column:post_id;type:bigint;not null"`
	UserID  int64     `json:"userID" gorm:"primaryKey;column:user_id;type:bigint;not null"`
	LikedAt time.Time `json:"-" gorm:"column:liked_at;autoCreateTime"`
}

func (Post) TableName() string {
	return "posts"
}
