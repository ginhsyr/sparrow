package model

import "time"

type User struct {
	ID               int        `json:"id" gorm:"primaryKey;column:id;type:integer;autoIncrement"`
	Nickname         string     `json:"nickname" gorm:"type:varchar(20)"`
	RealName         string     `json:"realName" gorm:"type:varchar(20)"`
	Email            string     `json:"email" gorm:"unique"`
	IsEmailConfirmed bool       `json:"isEmailConfirmed"`
	Password         string     `json:"-"`
	AvatarUrl        string     `json:"avatarUrl"`
	JoinAt           time.Time  `json:"joinAt" gorm:"autoCreateTime;type:timestamp"`
	PostCount        int        `json:"postCount" gorm:"type:integer"`
	CommentCount     int        `json:"commentCount" gorm:"type:integer"`
	Birthday         *time.Time `json:"birthday"`
	Role             RoleType   `json:"-" gorm:"type:smallint"`
}

type UserLogin struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type RoleType int

const (
	Users RoleType = iota
	Admin
	Visitors
)
