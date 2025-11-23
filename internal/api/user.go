package api

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"minifeed/internal/middleware"
	"minifeed/internal/service"
)

type Response struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data,omitempty"`
}

func OK(c *gin.Context, data any) {
	c.JSON(http.StatusOK, Response{
		Code: 0,
		Msg:  "success",
		Data: data,
	})
}

func Fail(c *gin.Context, code int, msg string) {
	c.JSON(http.StatusOK, Response{
		Code: code,
		Msg:  msg,
	})
}

func UserRoutes(r *gin.Engine, userSvc *service.UserService) {

	//==================== register ======================
	r.POST("/user/register", func(c *gin.Context) {
		var req struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			Fail(c, 1001, "invalid request!")
			return
		}
		//empty field check
		if req.Username == "" || req.Password == "" {
			Fail(c, 1002, "username or password is empty!")
			return
		}

		u, err := userSvc.Register(req.Username, req.Password)
		if err != nil {
			if errors.Is(err, service.ErrUserExists) {
				Fail(c, 1004, "username alerady exists")
				return
			}
			Fail(c, 1003, "db error")
			return
		}

		OK(c, gin.H{
			"user_id":  u.ID,
			"username": u.Username,
		})
	})

	//==================== login =======================
	r.POST("/user/login", func(c *gin.Context) {
		var req struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			Fail(c, 2001, "invalid request!")
			return
		}
		if req.Username == "" || req.Password == "" {
			Fail(c, 2002, "username or password is empty!")
			return
		}

		u, token, err := userSvc.Login(req.Username, req.Password)
		if err != nil {
			if errors.Is(err, service.ErrUserNotFound) {
				Fail(c, 2003, "user not found")
				return
			}
			if errors.Is(err, service.ErrWrongPassword) {
				Fail(c, 2005, "wrong password")
				return
			}
			Fail(c, 2004, "db error")
			return
		}

		OK(c, gin.H{
			"user_id":  u.ID,
			"username": u.Username,
			"msg":      "login succeeded",
			"token":    token,
		})
	})

	//================================== User APIs (Require Authentication) =============================
	authGroup := r.Group("/api", middleware.Auth())

	authGroup.GET("/users/search", func(c *gin.Context) {
		keyword := c.Query("keyword")
		if keyword == "" {
			Fail(c, 4001, "keyword is empty")
			return
		}

		users, err := userSvc.SearchByUsername(keyword, 20)
		if err != nil {
			Fail(c, 4002, "db error")
			return
		}

		OK(c, gin.H{
			"list": users,
		})
	})

	authGroup.GET("/me", func(c *gin.Context) {
		userID, _ := c.Get("user_id")

		c.JSON(200, gin.H{
			"msg":     "Hello user",
			"user_id": userID,
		})
	})

}
