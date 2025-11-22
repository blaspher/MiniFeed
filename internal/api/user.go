package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"minifeed/internal/middleware"
	"minifeed/internal/model"
	jwtUtil "minifeed/pkg/jwt"
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

func UserRoutes(r *gin.Engine, db *gorm.DB) {

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
		//deplicated name check
		var count int64
		if err := db.Model(&model.User{}).Where("username = ?", req.Username).Count(&count).Error; err != nil {
			Fail(c, 1003, "db error!")
			return
		}
		if count > 0 {
			Fail(c, 1004, "username already exists!")
			return
		}

		//bcrypt encrypt password
		hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			Fail(c, 1005, "encrypt password failed!")
			return
		}

		//create user
		u := model.User{
			Username: req.Username,
			Password: string(hashed),
		}
		if err := db.Create(&u).Error; err != nil {
			Fail(c, 1006, "creat user failed!")
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

		//query user
		var u model.User
		if err := db.Where("username = ?", req.Username).First(&u).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				Fail(c, 2003, "user not found!")
				return
			}
			Fail(c, 2004, "db error!")
			return
		}

		//check password
		if err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(req.Password)); err != nil {
			Fail(c, 2005, "wrong password!")
			return
		}

		token, err := jwtUtil.GenerateToken(u.ID)
		if err != nil {
			Fail(c, 2006, "generate Token failed!")
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
		var users []model.User
		if err := db.Where("username LIKE ?", "%"+keyword+"%").Limit(20).Find(&users).Error; err != nil {
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
