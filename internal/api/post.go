package api

import (
	"errors"
	"strconv"

	"minifeed/internal/middleware"
	"minifeed/internal/service"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// push the newly created post into each follower's index
/*func pushPostInbox(post model.Post, db *gorm.DB) {
	ctx := context.Background()
	//query all followers
	var rels []model.Follow
	if err := db.Where("follow_id = ?", post.UserID).Find(&rels).Error; err != nil {
		return
	}
	//collect all user IDs who should receive this post: all followers + the author himself
	userIDs := make([]uint, 0, len(rels)+1)
	userIDs = append(userIDs, post.UserID)
	for _, r := range rels {
		userIDs = append(userIDs, r.UserID)
	}
	//use post creation timestamp as score for sorting in the ZSet
	score := float64(post.CreatedAt.Unix())
	//push the post into each user's inbox in Redis
	for _, uid := range userIDs {
		key := fmt.Sprintf("inbox:%d", uid)
		_ = config.Rdb.ZAdd(ctx, key, redis.Z{
			Score:  score,
			Member: post.ID,
		}).Err()
	}

}*/

func PostRoutes(r *gin.Engine, svc *service.PostService) {
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

			post, err := svc.CreatePost(userID, req.Content, req.ImageURL)
			if err != nil {
				Fail(c, 5001, "db error")
				return
			}

			OK(c, gin.H{
				"post_id":    post.ID,
				"user_id":    post.UserID,
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

			liked, likeCount, err := svc.ToggleLike(userID, postID)
			if err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					Fail(c, 6004, "post not found")
					return
				}
				Fail(c, 6005, "internal error")
				return
			}

			OK(c, gin.H{
				"post_id":    postID,
				"liked":      liked,
				"like_count": likeCount,
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
		var cursor uint64
		if cursorStr != "" {
			if cVal, err := strconv.ParseUint(cursorStr, 10, 64); err == nil && cVal > 0 {
				cursor = cVal
			}
		}

		posts, nextCursor, err := svc.ListPublicPosts(limit, cursor)
		if err != nil {
			Fail(c, 5002, "db error")
			return
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
		var cursor uint64
		if cursorStr != "" {
			if cVal, err := strconv.ParseUint(cursorStr, 10, 64); err == nil && cVal > 0 {
				cursor = cVal
			}
		}

		posts, nextCursor, err := svc.ListFollowFeed(userID, limit, cursor)
		if err != nil {
			Fail(c, 3044, "db error")
			return
		}

		OK(c, gin.H{
			"list":        posts,
			"next_cursor": nextCursor,
		})

	})

	//============================== reaad feed data from the user's inbox in push mode =======================
	authGroup.GET("/feed/push", func(c *gin.Context) {

		//get current user ID from context
		uidVal, ok := c.Get("user_id")
		if !ok {
			Fail(c, 3051, "no user in context")
			return
		}
		userID, ok := uidVal.(uint)
		if !ok {
			Fail(c, 3052, "invalid user id")
			return
		}

		//pagination parameters
		limitStr := c.DefaultQuery("limit", "10")
		limit, err := strconv.Atoi(limitStr)
		if err != nil || limit <= 0 || limit > 100 {
			limit = 100
		}
		cursor := c.Query("cursor")

		posts, nextCursor, err := svc.ListInboxFeed(userID, limit, cursor)
		if err != nil {
			Fail(c, 3053, "db or cache error")
			return
		}

		//return feed list + next cursor
		OK(c, gin.H{
			"list":        posts,
			"next_cursor": nextCursor,
		})

	})

	//=================================== hot posts feed (by like_count, cached in Redis) ================================
	authGroup.GET("/feed/hot", func(c *gin.Context) {

		limitStr := c.DefaultQuery("limit", "10")
		limit, err := strconv.Atoi(limitStr)
		if err != nil || limit <= 0 || limit > 50 {
			limit = 10
		}

		posts, err := svc.ListHotPosts(limit)
		if err != nil {
			Fail(c, 5003, "db or cache error")
			return
		}

		OK(c, gin.H{
			"lists": posts,
		})

	})

}
