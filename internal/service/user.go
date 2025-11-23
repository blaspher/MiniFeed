package service

import (
	"errors"

	"minifeed/internal/model"
	jwtUtil "minifeed/pkg/jwt"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var (
	ErrUserExists    = errors.New("username already exists")
	ErrUserNotFound  = errors.New("user not found")
	ErrWrongPassword = errors.New("wrong password")
)

type UserService struct {
	db *gorm.DB
}

func NewUserService(db *gorm.DB) *UserService {
	return &UserService{db: db}
}

// Register
func (s *UserService) Register(username, password string) (*model.User, error) {
	var count int64
	if err := s.db.Model(&model.User{}).Where("username = ?", username).Count(&count).Error; err != nil {
		return nil, err
	}
	if count > 0 {
		return nil, ErrUserExists
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	u := &model.User{
		Username: username,
		Password: string(hashed),
	}
	if err := s.db.Create(u).Error; err != nil {
		return nil, err
	}
	return u, nil

}

// login
func (s *UserService) Login(username, password string) (*model.User, string, error) {
	var u model.User
	if err := s.db.Where("username = ?", username).First(&u).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, "", ErrUserNotFound
		}
		return nil, "", err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password)); err != nil {
		return nil, "", ErrWrongPassword
	}

	token, err := jwtUtil.GenerateToken(u.ID)
	if err != nil {
		return nil, "", err
	}

	return &u, token, nil

}

// search by username
func (s *UserService) SearchByUsername(keyword string, limit int) ([]model.User, error) {
	if limit <= 0 || limit > 20 {
		limit = 20
	}

	var users []model.User
	if err := s.db.Limit(limit).Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil

}
