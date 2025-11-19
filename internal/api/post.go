package api

import (
	"context"
	"fmt"
	"strconv"

	"minifeed/internal/config"
	"minifeed/internal/middleware"
	"minifeed/internal/model"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func syncLikeCountToDB(postID uint, count int, db *gorm.DB) {
	_ = db.Model(&model.Post{}).Where("id = ?", postID).Update("like_count", count).Error
}

func PostRoutes(r *gin.Engine, db *gorm.DB) {
	//=============privacy:post a status update, need to login================
	authGroup := r.Group("/api", middleware.Auth())
	{
		//post a status update
		authGroup.POST("/post", func(c *gin.Context) {
			var req struct {
				Content  string `json:"content"`
				ImageURL string `json:"image_url"`
			}
			if err := c.ShouldBindJSON(&req); err != nil || req.Content == "" {
				Fail(c, 4001, "invalid content")
				return
			}

			//retrieve user_id from JWT
			uidVal, ok := c.Get("user_id")
			if !ok {
				Fail(c, 4002, "no user in context")
				return
			}
			userID, ok := uidVal.(uint)
			if !ok {
				Fail(c, 4003, "invalid user id")
				return
			}
			post := model.Post{
				UserID:   userID,
				Content:  req.Content,
				ImageURL: req.ImageURL,
			}
			if err := db.Create(&post).Error; err != nil {
				Fail(c, 5001, "db error")
				return
			}
			OK(c, gin.H{
				"post_id":    post.ID,
				"user_ud":    post.UserID,
				"content":    post.Content,
				"image_url":  post.ImageURL,
				"created_at": post.CreatedAt,
			})
		})

		//like and unlike
		authGroup.POST("/post/:id/like", func(c *gin.Context) {
			//parse current user
			uidVal, ok := c.Get("user_id")
			if !ok {
				Fail(c, 6001, "no user in context")
				return
			}
			userID, ok := uidVal.(uint)
			if !ok {
				Fail(c, 6002, "invalid user id")
				return
			}
			//parse post ID
			postIDStr := c.Param("id")
			postID64, err := strconv.ParseUint(postIDStr, 10, 64)
			if err != nil || postID64 == 0 {
				Fail(c, 6003, "invalid post id")
				return
			}
			postID := uint(postID64)
			//verify post existence
			var post model.Post
			if err := db.Select("id").Where("id = ?", postID).First(&post).Error; err != nil {
				if err == gorm.ErrRecordNotFound {
					Fail(c, 6004, "post not found")
					return
				}
				Fail(c, 6005, "db error")
				return
			}

			ctx := context.Background()
			likeSetKey := fmt.Sprintf("like:%d", postID)
			likeCountKey := fmt.Sprintf("like_count:%d", postID)
			userIDStr := fmt.Sprintf("%d", userID)
			//check if user has liked the post
			isMember, err := config.Rdb.SIsMember(ctx, likeSetKey, userIDStr).Result()
			if err != nil {
				Fail(c, 6006, "redis Error")
				return
			}
			var liked bool
			if !isMember {
				if err := config.Rdb.SAdd(ctx, likeSetKey, userIDStr).Err(); err != nil {
					Fail(c, 6007, "redis error")
					return
				}
				liked = true
			} else {
				if err := config.Rdb.SRem(ctx, likeSetKey, userIDStr).Err(); err != nil {
					Fail(c, 6008, "redis error")
					return
				}
				liked = false
			}

			//recalculate total like count from the Redis set(authoritative source)
			count, err := config.Rdb.SCard(ctx, likeSetKey).Result()
			if err != nil {
				Fail(c, 6009, "redis error")
			}
			//update like count cache for fast reads
			if err := config.Rdb.Set(ctx, likeCountKey, count, 0).Err(); err != nil {
				Fail(c, 6010, "redis error")
				return
			}

			go syncLikeCountToDB(postID, int(count), db)

			OK(c, gin.H{
				"post_id":    postID,
				"liked":      liked,
				"like_count": count,
			})

		})

	}

	//=============== public: newest first + cursor-based pagination =======================
	r.GET("/posts", func(c *gin.Context) {
		//limit: number per page
		limitStr := c.DefaultQuery("limit", "10")
		limit, err := strconv.Atoi(limitStr)
		if err != nil || limit <= 0 || limit > 100 {
			limit = 10
		}

		//cursor: last post_id from previous page
		cursorStr := c.Query("cursor")

		var posts []model.Post
		query := db.Order("id DESC").Limit(limit)

		if cursorStr != "" {
			if cursor, err := strconv.ParseUint(cursorStr, 10, 64); err == nil && cursor > 0 {
				query = query.Where("id < ?", cursor)
			}
		}

		if err := query.Find(&posts).Error; err != nil {
			Fail(c, 5002, "db error")
			return
		}

		//cursor: last id from current page
		var nextCursor uint
		if len(posts) > 0 {
			nextCursor = posts[len(posts)-1].ID
		}

		OK(c, gin.H{
			"list":        posts,
			"next_cursor": nextCursor,
		})
	})

	//======================= posts from people I follow ===============================
	authGroup.GET("/feed/pull", func(c *gin.Context) {

		//get user_id
		uidVal, ok := c.Get("user_id")
		if !ok {
			Fail(c, 3041, "no user in context")
			return
		}
		userID, ok := uidVal.(uint)
		if !ok {
			Fail(c, 3042, "invalid user id")
			return
		}

		//parse pagination params:limit/cursor
		limitStr := c.DefaultQuery("limit", "10")
		limit, err := strconv.Atoi(limitStr)
		if err != nil || limit < 0 || limit > 100 {
			limit = 10
		}
		cursorStr := c.Query("cursor") //last post_id from previous page

		//check who I follow
		var rels []model.Follow
		if err := db.Where("user_id = ?", userID).Find(&rels).Error; err != nil {
			Fail(c, 3043, "db error")
			return
		}
		if len(rels) == 0 {
			OK(c, gin.H{
				"list":        []model.Post{},
				"next_cursor": 0,
			})
			return
		}

		//
		ids := make([]uint, 0, len(rels)+1)
		for _, r := range rels {
			ids = append(ids, r.FollowID)
		}

		//=========================query their posts from 'post' table(id DES + cursor-based pagination)=================
		var posts []model.Post
		query := db.Where("user_id IN ?", ids).Order("id DESC").Limit(limit)
		if cursorStr != "" {
			if cursor, err := strconv.ParseUint(cursorStr, 10, 64); err == nil && cursor > 0 {
				query = query.Where("id < ?", cursor)
			}
		}
		if err := query.Find(&posts).Error; err != nil {
			Fail(c, 3044, "db error")
			return
		}

		//=========================== calculate next page cursor (the last id from current page)=============================
		var nextCursor uint
		if len(posts) > 0 {
			nextCursor = posts[len(posts)-1].ID
		}

		OK(c, gin.H{
			"list":        posts,
			"next_cursor": nextCursor,
		})

	})

}
