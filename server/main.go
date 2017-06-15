package main

import (
	"net/http"

	"log"

	"github.com/gin-gonic/gin"
	"github.com/oscp/cloud-selfservice-portal/server/common"
	"github.com/oscp/cloud-selfservice-portal/server/openshift"
)

func main() {
	router := gin.New()
	router.Use(gin.Recovery())
	router.LoadHTMLGlob("templates/*")

	// Public routes
	authMiddleware := common.GetAuthMiddleware()
	router.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusTemporaryRedirect, "/auth/")
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
	}

	log.Println("Cloud SSP is running")
	router.Run()
}
