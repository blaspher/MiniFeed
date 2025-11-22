package cron

import (
	"log"
	"minifeed/internal/dao"
	"time"

	"gorm.io/gorm"
)

// refresh hot posts cache periodically
func StartHotPostsRefresh(db *gorm.DB) {
	if err := dao.RefreshHotPostsCache(db); err != nil {
		log.Printf("[cron] refresh hot posts cache failed: %v\n", err)
	} else {
		log.Println("[cron] hot posts cache refreshed")
	}

	ticker := time.NewTicker(1 * time.Minute)

	go func() {
		for range ticker.C {
			if err := dao.RefreshHotPostsCache(db); err != nil {
				log.Printf("[cron] refresh hot posts cache failed: %v\n", err)
			} else {
				log.Println("[cron] hot posts cache refreshed")
			}
		}
	}()

}
