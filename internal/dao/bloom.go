package dao

import (
	"fmt"
	"minifeed/internal/model"
	"sync"

	"github.com/bits-and-blooms/bloom/v3"
	"gorm.io/gorm"
)

var (
	postBloom   *bloom.BloomFilter
	postBloomMu sync.RWMutex
)

func InitPostBloom(db *gorm.DB, nEstimates uint) error {
	postBloomMu.Lock()
	defer postBloomMu.Unlock()

	postBloom = bloom.NewWithEstimates(nEstimates, 0.001)

	var posts []model.Post
	if err := db.Select("id").Find(&posts).Error; err != nil {
		return err
	}

	for _, p := range posts {
		idStr := fmt.Sprintf("%d", p.ID)
		postBloom.AddString(idStr)
	}

	return nil

}

func AddPostToBloom(postID uint) {
	postBloomMu.RLock()
	defer postBloomMu.RUnlock()

	if postBloom == nil {
		return
	}

	postBloom.AddString(fmt.Sprintf("%d", postID))

}

func PostMayExist(postID uint) bool {
	postBloomMu.RLock()
	defer postBloomMu.RUnlock()

	if postBloom == nil {
		return true
	}
	return postBloom.TestString(fmt.Sprintf("%d", postID))

}
