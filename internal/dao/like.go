package dao

import (
	"context"
	"strconv"
	"strings"

	"minifeed/internal/config"
	"minifeed/internal/model"

	"gorm.io/gorm"
)

var ctx = context.Background()

// fetches all Redis keys matching like_count:*
func GetAllLikeCountKeys() ([]string, error) {
	var (
		cursor uint64
		keys   []string
	)

	for {
		k, nextCursor, err := config.Rdb.Scan(ctx, cursor, "like_count:*", 100).Result()
		if err != nil {
			return nil, err
		}
		keys = append(keys, k...)
		cursor = nextCursor

		if cursor == 0 {
			break
		}
	}
	return keys, nil
}

// reas like_count:{postID} value
func GetLikeCountFromRedis(key string) (uint, error) {
	countStr, err := config.Rdb.Get(ctx, key).Result()
	if err != nil {
		return 0, err
	}
	v, _ := strconv.Atoi(countStr)
	return uint(v), nil
}

// extracts postID from key "like_count:{postID}"
func ExtractPostID(key string) (uint, error) {
	parts := strings.Split(key, ":")
	if len(parts) != 2 {
		return 0, nil
	}
	id64, err := strconv.ParseUint(parts[1], 10, 64)
	return uint(id64), err
}

// updates MySQL like_count
func UpdatePostLikeCount(db *gorm.DB, postID uint, count uint) error {
	DelHotPostsCache()

	err := db.Model(&model.Post{}).Where("id = ?", postID).Update("like_count", count).Error
	if err != nil {
		return err
	}

	DelHotPostsCacheAsync()

	return nil
}
