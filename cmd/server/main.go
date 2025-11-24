package main

import (
	"log"
	"os"

	"minifeed/internal/api"
	"minifeed/internal/config"
	"minifeed/internal/cron"
	"minifeed/internal/dao"
	"minifeed/internal/metrics"
	"minifeed/internal/middleware"
	"minifeed/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {

	mysqlDSN := os.Getenv("MYSQL_DSN")
	redisAddr := os.Getenv("REDIS_ADDR")
	jwtSecret := os.Getenv("JWT_SECRET")

	if mysqlDSN == "" || redisAddr == "" || jwtSecret == "" {
		log.Fatal("Missing required environment variables")
	}

	db := config.InitDB(mysqlDSN)
	rdb := config.InitRedis(redisAddr)

	if err := dao.InitPostBloom(db, 10000); err != nil {
		log.Printf("[warn] init post bloom failed: %v\n", err)
	}

	metrics.Init()

	cron.StartLikeSync(db)
	cron.StartHotPostsRefresh(db)

	userSvc := service.NewUserService(db)
	postSvc := service.NewPostService(db, rdb)
	followSvc := service.NewFollowService(db)

	r := gin.Default()
	r.Use(middleware.CORS(), middleware.RequestTiming(), middleware.PrometheusMiddleware())

	api.UserRoutes(r, userSvc)
	api.PostRoutes(r, postSvc)
	api.FollowRoutes(r, followSvc)

	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	r.Run(":8888")
}
