package service

import (
	"errors"
	"minifeed/internal/model"

	"gorm.io/gorm"
)

var (
	ErrFollowSelf = errors.New("cannot follow yourself")
)

type FollowService struct {
	db *gorm.DB
}

func NewFollowService(db *gorm.DB) *FollowService {
	return &FollowService{db: db}
}

// follow
func (s *FollowService) Follow(userID, targetID uint) error {
	if userID == targetID {
		return ErrFollowSelf
	}

	f := model.Follow{
		UserID:   userID,
		FollowID: targetID,
	}

	return s.db.FirstOrCreate(&f, "user_id = ? AND follow_id = ?", userID, targetID).Error
}

// unfollow
func (s *FollowService) UnFollow(userID, targetID uint) error {
	return s.db.Where("user_id = ? AND follow_id = ?", userID, targetID).Delete(&model.Follow{}).Error
}

// who I follow
func (s *FollowService) ListFollowing(userID uint) ([]model.User, error) {
	var rels []model.Follow
	if err := s.db.Where("user_id = ?", userID).Find(&rels).Error; err != nil {
		return nil, err
	}
	if len(rels) == 0 {
		return []model.User{}, nil
	}

	ids := make([]uint, 0, len(rels))
	for _, r := range rels {
		ids = append(ids, r.FollowID)
	}

	var users []model.User
	if err := s.db.Where("id IN ?", ids).Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

// my followers
func (s *FollowService) ListFollowers(userID uint) ([]model.User, error) {
	var rels []model.Follow
	if err := s.db.Where("follow_id = ?", userID).Find(&rels).Error; err != nil {
		return nil, err
	}
	if len(rels) == 0 {
		return []model.User{}, nil
	}

	ids := make([]uint, 0, len(rels))
	for _, r := range rels {
		ids = append(ids, r.UserID)
	}

	var users []model.User
	if err := s.db.Where("id IN ?", ids).Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}
