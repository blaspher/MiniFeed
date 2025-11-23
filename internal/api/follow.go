package api

import (
	"errors"
	"minifeed/internal/middleware"
	"minifeed/internal/service"

	"strconv"

	"github.com/gin-gonic/gin"
)

func FollowRoutes(r *gin.Engine, followSvc *service.FollowService) {
	authGroup := r.Group("/api", middleware.Auth())

	//=================== follow an user ===================
	authGroup.POST("/follow/:id", func(c *gin.Context) {
		uidVal, ok := c.Get("user_id")
		if !ok {
			Fail(c, 3001, "no user in context")
			return
		}
		userID, ok := uidVal.(uint)
		if !ok {
			Fail(c, 3002, "invalid user id")
			return
		}

		//followed user_id
		targetIDStr := c.Param("id")
		targetID64, err := strconv.ParseUint(targetIDStr, 10, 64)
		if err != nil || targetID64 == 0 {
			Fail(c, 3003, "invalid target id")
			return
		}
		targetID := uint(targetID64)

		if err := followSvc.Follow(userID, targetID); err != nil {
			if errors.Is(err, service.ErrFollowSelf) {
				Fail(c, 3004, "cannot follow yourself")
				return
			}
			Fail(c, 3005, "db error")
			return
		}

		OK(c, gin.H{
			"msg":       "follow succeeded",
			"user_id":   userID,
			"follow_id": targetID,
		})

	})

	//======================= unfollow an user ======================
	authGroup.POST("/unfollow/:id", func(c *gin.Context) {
		uidVal, ok := c.Get("user_id")
		if !ok {
			Fail(c, 3011, "no user in context")
			return
		}
		userID, ok := uidVal.(uint)
		if !ok {
			Fail(c, 3012, "invalid user id")
			return
		}

		targetIDStr := c.Param("id")
		targetID64, err := strconv.ParseUint(targetIDStr, 10, 64)
		if err != nil || targetID64 == 0 {
			Fail(c, 3013, "invalid target id")
			return
		}
		targetID := uint(targetID64)

		if err := followSvc.UnFollow(userID, targetID); err != nil {
			Fail(c, 3014, "db.Error")
			return
		}

		OK(c, gin.H{
			"msg":       "unfollow succeeded",
			"user_id":   userID,
			"follow_id": targetID,
		})

	})

	//========================== following list ===================
	authGroup.GET("/following", func(c *gin.Context) {
		uidVal, ok := c.Get("user_id")
		if !ok {
			Fail(c, 3021, "no user in context")
			return
		}
		userID, ok := uidVal.(uint)
		if !ok {
			Fail(c, 3022, "invalid user id")
			return
		}

		users, err := followSvc.ListFollowing(userID)
		if err != nil {
			Fail(c, 3023, "db error")
			return
		}

		OK(c, gin.H{
			"list": users,
		})

	})

	//======================= My Followers =============================
	authGroup.GET("/followers", func(c *gin.Context) {
		uidVal, ok := c.Get("user_id")
		if !ok {
			Fail(c, 3031, "no user in context")
			return
		}
		userID, ok := uidVal.(uint)
		if !ok {
			Fail(c, 3032, "invalid user id")
			return
		}

		users, err := followSvc.ListFollowers(userID)
		if err != nil {
			Fail(c, 3033, "db error")
		}

		OK(c, gin.H{
			"list": users,
		})

	})
}
