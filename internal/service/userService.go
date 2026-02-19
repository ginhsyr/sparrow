package service

import (
	"Sparrow/internal/model"
	"Sparrow/internal/repository"
	"Sparrow/internal/utils"
	"errors"
	"github.com/jackc/pgx/v5/pgconn"
	"time"
)

var ErrEmailAlreadyRegistered = errors.New("email already registered")

type UserService struct {
	repo *repository.UserRepository
}

func NewUserService(repo *repository.UserRepository) *UserService {
	return &UserService{repo: repo}
}

func (s *UserService) GetUser(id string) (*model.User, error) {
	return s.repo.GetUser(id)
}

func (s *UserService) GetUserByEmail(email string) (*model.User, error) {
	return s.repo.GetUserByEmail(email)
}

func (s *UserService) UserRegister(nickname, realName, email, password string, birthday *time.Time) (*model.User, error) {
	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		return nil, err
	}

	user := &model.User{
		Nickname:         nickname,
		RealName:         realName,
		Email:            email,
		IsEmailConfirmed: false,
		Password:         hashedPassword,
		AvatarUrl:        "",
		PostCount:        0,
		CommentCount:     0,
		Birthday:         birthday,
		Role:             model.Users,
	}
	err = s.repo.UserRegister(user)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, ErrEmailAlreadyRegistered
		}
		return nil, err
	}
	return user, err
}

func (s *UserService) UpgradePasswordHash(userID int64, rawPassword string) error {
	hashedPassword, err := utils.HashPassword(rawPassword)
	if err != nil {
		return err
	}
	return s.repo.UpdatePassword(userID, hashedPassword)
}
