package main

import (
	"net/http"
	"gopkg.in/appleboy/gin-jwt.v2"
	"time"
	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()

	authMiddleware := &jwt.GinJWTMiddleware{
		Realm:      "OSE_SSP",
		Key:        []byte("secret key"),
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
		// TokenLookup is a string in the form of "<source>:<name>" that is used
		// to extract token from the request.
		// Optional. Default value "header:Authorization".
		// Possible values:
		// - "header:<name>"
		// - "query:<name>"
		// - "cookie:<name>"
		TokenLookup: "header:Authorization",
		// TokenLookup: "query:token",
		// TokenLookup: "cookie:token",

		// TokenHeadName is a string in the header. Default value is "Bearer"
		TokenHeadName: "Bearer",

		// TimeFunc provides the current time. You can override it to use another time value. This is useful for testing or if your server uses a different time zone than your tokens.
		TimeFunc: time.Now,
	}

	router.POST("/login", authMiddleware.LoginHandler)

	auth := router.Group("/auth")
	auth.Use(authMiddleware.MiddlewareFunc())
	{
		auth.GET("/hello", echo)
		auth.GET("/refresh_token", authMiddleware.RefreshHandler)
	}

	router.Run()
}

func echo(c *gin.Context) {
	c.String(http.StatusOK, "Hello World")
}

//func ldapFunc() {
//	client := &ldap.LDAPClient{
//		Base:         "dc=example,dc=com",
//		Host:         "ldap.example.com",
//		Port:         389,
//		UseSSL:       false,
//		BindDN:       "uid=readonlysuer,ou=People,dc=example,dc=com",
//		BindPassword: "readonlypassword",
//		UserFilter:   "(uid=%s)",
//		GroupFilter: "(memberUid=%s)",
//		Attributes:   []string{"givenName", "sn", "mail", "uid"},
//	}
//	// It is the responsibility of the caller to close the connection
//	defer client.Close()
//
//	ok, user, err := client.Authenticate("username", "password")
//	if err != nil {
//		log.Fatalf("Error authenticating user %s: %+v", "username", err)
//	}
//	if !ok {
//		log.Fatalf("Authenticating failed for user %s", "username")
//	}
//	log.Printf("User: %+v", user)
//
//	groups, err := client.GetGroupsOfUser("username")
//	if err != nil {
//		log.Fatalf("Error getting groups for user %s: %+v", "username", err)
//	}
//	log.Printf("Groups: %+v", groups)
//}

