package main

import (
	"log"

	"minifeed/internal/api"
	"minifeed/internal/config"
	"minifeed/internal/cron"
	"minifeed/internal/dao"
	"minifeed/internal/metrics"
	"minifeed/internal/middleware"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	db := config.InitDB()

	config.InitRedis()

	if err := dao.InitPostBloom(db, 10000); err != nil {
		log.Printf("[warn] init post bloom failed: %v\n", err)
	}

	metrics.Init()

	cron.StartLikeSync(db)
	cron.StartHotPostsRefresh(db)

	r := gin.Default()

	r.Use(middleware.CORS(), middleware.RequestTiming(), middleware.PrometheusMiddleware())

	api.UserRoutes(r, db)
	api.PostRoutes(r, db)
	api.FollowRoutes(r, db)

	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	r.Run(":8888")
}
