package dao

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"sync"
	"time"

	"minifeed/internal/config"
	"minifeed/internal/model"

	"gorm.io/gorm"
)

const (
	hotPostsKey      = "hot:posts"
	hotPostsCacheTTL = 60 * time.Second
	hotPostsCacheTop = 100
	hotPostsEmptyKey = "hot:posts:empty"
)

var (
	hotCtx     = context.Background()
	hotBuildMu sync.Mutex
)

// applies a "double-delete" strategy for the hot posts cache
func InvalidateHotPostCache() {

	_ = config.Rdb.Del(hotCtx, hotPostsKey).Err()

	go func() {
		time.Sleep(100 * time.Millisecond)
		_ = config.Rdb.Del(hotCtx, hotPostsKey).Err()
	}()

}

// delete before write
func DelHotPostsCache() {
	_ = config.Rdb.Del(hotCtx, hotPostsKey).Err()
}

// delete after write
func DelHotPostsCacheAsync() {
	go func() {
		time.Sleep(100 * time.Millisecond)
		_ = config.Rdb.Del(hotCtx, hotPostsKey).Err()
	}()
}

// rebuild the hot posts cache periodically
func RefreshHotPostsCache(db *gorm.DB) error {
	_, err := buildHotPostsCache(db)
	return err
}

// loads hot posts from MySQL and writes their IDs to Redis
func buildHotPostsCache(db *gorm.DB) ([]model.Post, error) {
	var posts []model.Post
	if err := db.Order("like_count DESC").Limit(hotPostsCacheTop).Find(&posts).Error; err != nil {
		return nil, err
	}

	pipe := config.Rdb.TxPipeline()
	pipe.Del(hotCtx, hotPostsEmptyKey)

	if len(posts) == 0 {
		pipe.Del(hotCtx, hotPostsKey)
		pipe.Set(hotCtx, hotPostsEmptyKey, "1", 10*time.Second)

		if _, err := pipe.Exec(hotCtx); err != nil {
			return nil, err
		}
		return []model.Post{}, nil

	}

	pipe.Del(hotCtx, hotPostsKey)

	for _, p := range posts {
		pipe.RPush(hotCtx, hotPostsKey, fmt.Sprintf("%d", p.ID))
	}

	jitter := time.Duration(rand.Intn(30)) * time.Second
	pipe.Expire(hotCtx, hotPostsKey, hotPostsCacheTTL+jitter)

	if _, err := pipe.Exec(hotCtx); err != nil {
		return nil, err
	}

	return posts, nil
}

// reads hot posts
func GetHotPosts(db *gorm.DB, limit int) ([]model.Post, error) {
	if limit <= 0 {
		limit = 10
	}
	if limit > hotPostsCacheTop {
		limit = hotPostsCacheTop
	}

	empty, err := config.Rdb.Exists(hotCtx, hotPostsEmptyKey).Result()
	if err == nil && empty == 1 {
		return []model.Post{}, nil
	}

	idStrs, err := config.Rdb.LRange(hotCtx, hotPostsKey, 0, int64(limit-1)).Result()
	if err != nil {
		idStrs = nil
	}

	if len(idStrs) == 0 {
		hotBuildMu.Lock()
		defer hotBuildMu.Unlock()

		idStrs, _ = config.Rdb.LRange(hotCtx, hotPostsKey, 0, int64(limit-1)).Result()
		if len(idStrs) == 0 {
			empty, err := config.Rdb.Exists(hotCtx, hotPostsEmptyKey).Result()
			if err == nil && empty == 1 {
				return []model.Post{}, nil
			}

			if _, err := buildHotPostsCache(db); err != nil {
				var posts []model.Post
				if err := db.Order("like_count DESC").Order("id DESC").Limit(limit).Find(&posts).Error; err != nil {
					return nil, err
				}
				return posts, nil
			}

			idStrs, _ = config.Rdb.LRange(hotCtx, hotPostsKey, 0, int64(limit-1)).Result()
			if len(idStrs) == 0 {
				return []model.Post{}, nil
			}
		}

	}

	ids := make([]uint, 0, len(idStrs))
	for _, s := range idStrs {
		id64, err := strconv.ParseUint(s, 10, 64)
		if err != nil || id64 == 0 {
			continue
		}
		ids = append(ids, uint(id64))
	}
	if len(ids) == 0 {
		return []model.Post{}, nil
	}

	var posts []model.Post
	if err := db.Where("id IN ?", ids).Find(&posts).Error; err != nil {
		return nil, err
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

	return ordered, nil

}
