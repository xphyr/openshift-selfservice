package common

import (
	"gopkg.in/appleboy/gin-jwt.v2"
	"time"
	"github.com/gin-gonic/gin"
	"os"
	"log"
)

func GetAuthMiddleware() (*jwt.GinJWTMiddleware) {
	key := os.Getenv("SESSION_KEY")

	if (len(key) == 0) {
		log.Fatal("Env variable 'SESSION_KEY' must be specified")
	}

	return &jwt.GinJWTMiddleware{
		Realm:      "OSE_SSP",
		Key:        []byte(key),
		Timeout:    time.Hour,
		MaxRefresh: time.Hour,
		Authenticator: func(userId string, password string, c *gin.Context) (string, bool) {
			if (userId == "admin" && password == "admin") || (userId == "test" && password == "test") {
				return userId, true
			}

			return userId, false
		},
		Authorizator: func(userId string, c *gin.Context) bool {
			if userId == "admin" {
				return true
			}

			return false
		},
		Unauthorized: func(c *gin.Context, code int, message string) {
			c.JSON(code, gin.H{
				"code":    code,
				"message": message,
			})
		},
		TokenLookup: "header:Authorization",
		TokenHeadName: "Bearer",
		TimeFunc: time.Now,
	}
}
