package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/oscp/openshift-selfservice/server/common"
	"github.com/oscp/openshift-selfservice/server/openshift"
)

func main() {
	router := gin.New()
	router.Use(gin.Recovery())
	router.LoadHTMLGlob("templates/*")

	// Public routes
	authMiddleware := common.GetAuthMiddleware()
	router.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusPermanentRedirect, "/auth/")
	})
	router.GET("/login", func(c *gin.Context) {
		c.HTML(http.StatusOK, "login.html", gin.H{})
	})
	router.POST("/login", func(c *gin.Context) {
		common.CookieLoginHandler(authMiddleware, c)
	})
	router.GET("/logout", func(c *gin.Context) {
		c.Abort()
		c.SetCookie("token", "", -1, "", "", false, true)
		c.Redirect(http.StatusTemporaryRedirect, "/login")
	})

	// Protected routes
	auth := router.Group("/auth/")
	auth.Use(authMiddleware.MiddlewareFunc())
	{
		// Index page
		auth.GET("/", func(c *gin.Context) {
			c.HTML(http.StatusOK, "index.html", gin.H{})
		})

		// Openshift routes
		openshift.RegisterRoutes(auth)

		// Thirdparty routes
		// ...
	}

	router.Run()
}



