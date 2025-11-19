package api

import (
	"minifeed/internal/middleware"
	"minifeed/internal/model"

	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func FollowRoutes(r *gin.Engine, db *gorm.DB) {
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

		//self-following is not allowed
		if targetID == userID {
			Fail(c, 3004, "cannot follow yourself")
			return
		}

		f := model.Follow{
			UserID:   userID,
			FollowID: targetID,
		}
		if err := db.FirstOrCreate(&f, "user_id = ? AND follow_id = ?", userID, targetID).Error; err != nil {
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

		if err := db.Where("user_id = ? AND follow_id = ?", userID, targetID).Delete(&model.Follow{}).Error; err != nil {
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

		var rels []model.Follow
		if err := db.Where("user_id = ?", userID).Find(&rels).Error; err != nil {
			Fail(c, 3023, "db error")
			return
		}
		if len(rels) == 0 {
			OK(c, gin.H{"list": []model.User{}})
			return
		}

		ids := make([]uint, 0, len(rels))
		for _, r := range rels {
			ids = append(ids, r.FollowID)
		}

		var users []model.User
		if err := db.Where("id IN ?", ids).Find(&users).Error; err != nil {
			Fail(c, 3024, "db error")
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

		var rels []model.Follow
		if err := db.Where("follow_id = ?", userID).Find(&rels).Error; err != nil {
			Fail(c, 3033, "db error")
			return
		}
		if len(rels) == 0 {
			OK(c, gin.H{"list": []model.User{}})
			return
		}

		ids := make([]uint, 0, len(rels))
		for _, r := range rels {
			ids = append(ids, r.UserID)
		}

		var users []model.User
		if err := db.Where("id IN ?", ids).Find(&users).Error; err != nil {
			Fail(c, 3034, "db error")
			return
		}

		OK(c, gin.H{
			"list": users,
		})

	})
}
