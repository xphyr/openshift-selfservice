package common

import (
	"time"
	"github.com/gin-gonic/gin"
	"os"
	"log"
	"net/http"
	"gopkg.in/appleboy/gin-jwt.v2"
	jwt3 "gopkg.in/dgrijalva/jwt-go.v3"
)

func GetAuthMiddleware() (*jwt.GinJWTMiddleware) {
	key := os.Getenv("SESSION_KEY")

	if (len(key) == 0) {
		log.Fatal("Env variable 'SESSION_KEY' must be specified")
	}

	return &jwt.GinJWTMiddleware{
		Realm:      "CLOUD_SSP",
		Key:        []byte(key),
		Timeout:    time.Hour,
		MaxRefresh: time.Hour,
		Authenticator: func(userId string, password string, c *gin.Context) (string, bool) {
			if (len(userId) > 0) {
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
			c.Abort()
			if (message == "Cookie token empty") {
				message = ""
			}
			c.HTML(http.StatusOK, "login.html", gin.H{
				"Error": message,
			})
		},
		TokenLookup: "cookie:token",
		TimeFunc: time.Now,
	}
}

func CookieLoginHandler(mw *jwt.GinJWTMiddleware, c *gin.Context) {
	// Initial middleware default setting.
	mw.MiddlewareInit()

	username := c.PostForm("username")
	password := c.PostForm("password")

	if (len(username) == 0 || len(password) == 0) {
		mw.Unauthorized(c, http.StatusBadRequest, "Benutzername / Passwort nicht angegeben")
		return
	}

	if mw.Authenticator == nil {
		mw.Unauthorized(c, http.StatusInternalServerError, "Internes Problem")
		return
	}

	userID, ok := mw.Authenticator(username, password, c)

	if !ok {
		mw.Unauthorized(c, http.StatusUnauthorized, "Benutzername / Passwort nicht korrekt")
		return
	}

	// Create the token
	token := jwt3.New(jwt3.GetSigningMethod(mw.SigningAlgorithm))
	claims := token.Claims.(jwt3.MapClaims)

	if mw.PayloadFunc != nil {
		for key, value := range mw.PayloadFunc(username) {
			claims[key] = value
		}
	}

	if userID == "" {
		userID = username
	}

	expire := mw.TimeFunc().Add(mw.Timeout)
	claims["id"] = userID
	claims["exp"] = expire.Unix()
	claims["orig_iat"] = mw.TimeFunc().Unix()

	tokenString, err := token.SignedString(mw.Key)

	if err != nil {
		mw.Unauthorized(c, http.StatusUnauthorized, "Token konnte nicht erstellt werden")
		return
	}

	c.SetCookie("token", tokenString, 0, "", "", false, true)

	c.Redirect(http.StatusFound, "/auth/")
}
