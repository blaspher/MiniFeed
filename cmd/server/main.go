package main

import (
	"minifeed/internal/api"
	"minifeed/internal/config"
	"minifeed/internal/cron"
	"minifeed/internal/middleware"

	"github.com/gin-gonic/gin"
)

func main() {
	db := config.InitDB()

	config.InitRedis()

	cron.StartLikeSync(db)

	r := gin.Default()

	r.Use(middleware.CORS())

	api.UserRoutes(r, db)
	api.PostRoutes(r, db)
	api.FollowRoutes(r, db)

	authGroup := r.Group("/api", middleware.Auth())
	{
		authGroup.GET("/me", func(c *gin.Context) {
			userID, _ := c.Get("user_id")

			c.JSON(200, gin.H{
				"msg":     "Hello user",
				"user_id": userID,
			})
		})
	}

	r.Run(":8888")
}
