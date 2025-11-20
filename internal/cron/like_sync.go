package cron

import (
	"log"
	"time"

	"minifeed/internal/dao"

	"gorm.io/gorm"
)

// periodically sync Redis like_count back to MySQL
func StartLikeSync(db *gorm.DB) {
	ticker := time.NewTicker(10 * time.Second)

	go func() {
		for range ticker.C {
			keys, err := dao.GetAllLikeCountKeys()
			if err != nil {
				log.Println("[cron] failed to fetch like_count keys:", err)
				continue
			}

			for _, key := range keys {
				postID, err := dao.ExtractPostID(key)
				if err != nil || postID == 0 {
					continue
				}

				count, err := dao.GetLikeCountFromRedis(key)
				if err != nil {
					continue
				}

				if err := dao.UpdatePostLikeCount(db, postID, count); err != nil {
					log.Printf("[cron] failed to update MySQL (post_id=%d): %v\n", postID, err)
					continue
				}

				log.Printf("[cron] synced post_id = %d, like_id = %d\n", postID, count)
			}
		}
	}()
}
