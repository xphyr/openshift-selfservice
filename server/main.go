package main

import (
	"log"
	"net/http"
	"os"

	"strings"

	"github.com/gin-gonic/gin"
	"github.com/oscp/openshift-selfservice/server/common"
	"github.com/oscp/openshift-selfservice/server/openshift"
)

func main() {
	// select directoy for translated web pages
	lang := strings.ToLower(os.Getenv("I18N_LANG"))
	if len(lang) == 0 {
		log.Fatal("Env variable 'I18N_LANG' must be specified")
	}

	router := gin.New()
	router.Use(gin.Recovery())
	router.LoadHTMLGlob("templates/" + lang + "/*")

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
