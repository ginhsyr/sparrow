package repository

import (
	"Sparrow/internal/model"
	"gorm.io/gorm"
)

type UserRepository struct {
	DB *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{DB: db}
}

func (r *UserRepository) GetUser(id string) (*model.User, error) {
	var user model.User
	if err := r.DB.Session(&gorm.Session{NewDB: true}).First(&user, id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) GetUserByEmail(email string) (*model.User, error) {
	var user model.User
	if err := r.DB.Session(&gorm.Session{NewDB: true}).Where("email = ?", email).Take(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) UserRegister(user *model.User) error {
	return r.DB.Session(&gorm.Session{NewDB: true}).Create(user).Error
}

func (r *UserRepository) UpdatePassword(userID int, passwordHash string) error {
	return r.DB.Session(&gorm.Session{NewDB: true}).
		Model(&model.User{}).
		Where("id = ?", userID).
		Update("password", passwordHash).Error
}
