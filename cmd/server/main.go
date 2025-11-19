package main

import (
	"minifeed/internal/api"
	"minifeed/internal/config"
	"minifeed/internal/middleware"

	"github.com/gin-gonic/gin"
)

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Headers", "Content-Type,Authorization")
		c.Header("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

func main() {
	db := config.InitDB()

	config.InitRedis()

	r := gin.Default()

	r.Use(CORSMiddleware())

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
