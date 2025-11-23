package service

import (
	"context"
	"fmt"
	"minifeed/internal/dao"
	"minifeed/internal/model"
	"strconv"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type PostService struct {
	db  *gorm.DB
	rdb *redis.Client
}

func NewPostService(db *gorm.DB, rdb *redis.Client) *PostService {
	return &PostService{
		db:  db,
		rdb: rdb,
	}
}

// create posts
func (s *PostService) CreatePost(userID uint, content, imageURL string) (*model.Post, error) {
	post := model.Post{
		UserID:   userID,
		Content:  content,
		ImageURL: imageURL,
	}

	if err := s.db.Create(&post).Error; err != nil {
		return nil, err
	}

	dao.AddPostToBloom(post.ID)

	dao.DelHotPostsCache()

	go s.pushPostInbox(post)

	return &post, nil

}

// public: newest first + cursor-based pagination
func (s *PostService) ListPublicPosts(limit int, cursor uint64) ([]model.Post, uint64, error) {
	if limit <= 10 || limit > 100 {
		limit = 10
	}

	var posts []model.Post
	query := s.db.Order("id DESC").Limit(limit)
	if cursor > 0 {
		query = query.Where("id < ?", cursor)
	}

	if err := query.Find(&posts).Error; err != nil {
		return nil, 0, err
	}

	var nextCursor uint64
	if len(posts) > 0 {
		nextCursor = uint64(posts[len(posts)-1].ID)
	}

	return posts, nextCursor, nil

}

// pull mode: posts from users I follow
func (s *PostService) ListFollowFeed(userID uint, limit int, cursor uint64) ([]model.Post, uint64, error) {
	if limit < 0 || limit > 100 {
		limit = 10
	}

	var rels []model.Follow
	if err := s.db.Where("user_id = ?", userID).Find(&rels).Error; err != nil {
		return nil, 0, err
	}
	if len(rels) == 0 {
		return []model.Post{}, 0, nil
	}

	ids := make([]uint, 0, len(rels))
	for _, r := range rels {
		ids = append(ids, r.FollowID)
	}

	var posts []model.Post
	query := s.db.Where("user_id IN ?", ids).Order("id DESC").Limit(limit)
	if cursor > 0 {
		query = query.Where("id < ?", cursor)
	}
	if err := query.Find(&posts).Error; err != nil {
		return nil, 0, err
	}

	var nextCursor uint64
	if len(posts) > 0 {
		nextCursor = uint64(posts[len(posts)-1].ID)
	}

	return posts, nextCursor, nil

}

// push mode: read one post from inbox ZSet
func (s *PostService) ListInboxFeed(userID uint, limit int, cursor string) ([]model.Post, string, error) {
	if limit <= 0 || limit > 100 {
		limit = 10
	}

	ctx := context.Background()
	inboxKey := fmt.Sprintf("inbox:%d", userID)

	max := "+inf"
	if cursor != "" {
		max = "(" + cursor
	}

	zs, err := s.rdb.ZRevRangeByScoreWithScores(ctx, inboxKey, &redis.ZRangeBy{
		Max:    max,
		Min:    "0",
		Offset: 0,
		Count:  int64(limit),
	}).Result()
	if err != nil {
		return nil, "", err
	}
	if len(zs) == 0 {
		return []model.Post{}, "", nil
	}

	ids := make([]uint, 0, len(zs))
	scores := make([]float64, 0, len(zs))
	for _, z := range zs {
		idStr := fmt.Sprint(z.Member)
		id64, err := strconv.ParseUint(idStr, 10, 64)
		if err != nil || id64 == 0 {
			continue
		}
		ids = append(ids, uint(id64))
		scores = append(scores, z.Score)
	}
	if len(ids) == 0 {
		return []model.Post{}, "", nil
	}

	var posts []model.Post
	if err := s.db.Where("id IN ?", ids).Find(&posts).Error; err != nil {
		return nil, "", err
	}

	m := make(map[uint]model.Post, len(posts))
	for _, p := range posts {
		m[p.ID] = p
	}
	ordered := make([]model.Post, 0, len(ids))
	for _, id := range ids {
		if p, ok := m[id]; ok {
			ordered = append(ordered, p)
		}
	}

	nextCursor := ""
	if len(scores) > 0 {
		nextCursor = fmt.Sprintf("%.0f", scores[len(scores)-1])
	}

	return ordered, nextCursor, nil

}

// like and unlike
func (s *PostService) ToggleLike(userID, postID uint) (bool, int64, error) {
	if !dao.PostMayExist(postID) {
		return false, 0, gorm.ErrRecordNotFound
	}

	var post model.Post
	if err := s.db.Select("id").Where("id = ?", postID).First(&post).Error; err != nil {
		return false, 0, err
	}

	ctx := context.Background()
	likeSetKey := fmt.Sprintf("like:%d", postID)
	likeCountKey := fmt.Sprintf("like_count:%d", postID)
	userIDStr := fmt.Sprintf("%d", userID)

	dao.DelHotPostsCache()

	isMember, err := s.rdb.SIsMember(ctx, likeSetKey, userIDStr).Result()
	if err != nil {
		return false, 0, err
	}

	var liked bool
	if !isMember {
		if err := s.rdb.SAdd(ctx, likeSetKey, userIDStr).Err(); err != nil {
			return false, 0, err
		}
		liked = true
	} else {
		if err := s.rdb.SRem(ctx, likeSetKey, userIDStr).Err(); err != nil {
			return false, 0, err
		}
		liked = false
	}

	count, err := s.rdb.SCard(ctx, likeSetKey).Result()
	if err != nil {
		return false, 0, err
	}

	if err := s.rdb.Set(ctx, likeCountKey, count, 0).Err(); err != nil {
		return liked, count, nil
	}

	dao.DelHotPostsCacheAsync()

	return liked, count, nil
}

func (s *PostService) ListHotPosts(limit int) ([]model.Post, error) {
	return dao.GetHotPosts(s.db, limit)
}

// push the new post to the author's and all followers' inboxes
func (s *PostService) pushPostInbox(post model.Post) {
	ctx := context.Background()

	var rels []model.Follow
	if err := s.db.Where("follow_id = ?", post.UserID).Find(&rels).Error; err != nil {
		return
	}

	userIDs := make([]uint, 0, len(rels)+1)
	userIDs = append(userIDs, post.UserID)
	for _, r := range rels {
		userIDs = append(userIDs, r.UserID)
	}

	score := float64(post.CreatedAt.Unix())

	for _, uid := range userIDs {
		key := fmt.Sprintf("inbox:%d", uid)
		_ = s.rdb.ZAdd(ctx, key, redis.Z{
			Score:  score,
			Member: post.ID,
		}).Err()
	}
}
