package main

import (
	"log"
	"flag"
	"github.com/gin-gonic/gin"
	"strconv"
	"github.com/oscp/openshift-selfservice/glusterapi/gluster"
	"github.com/oscp/openshift-selfservice/glusterapi/models"
	"net/http"
)

const (
	wrongApiUsageError = "Wrong API usage. Your payload did not match the endpoint"
)

func init() {
	flag.IntVar(&gluster.Port, "port", 8080, "Specify the api-port")
	flag.IntVar(&gluster.MaxGB, "maxGB", 100, "Max GB a user can order per volume")
	flag.IntVar(&gluster.MaxMB, "maxMB", 1024, "Max MB a user can order per volume")
	flag.StringVar(&gluster.PoolName, "poolName", "", "Specify which lvm pool should be used for orders")
	flag.StringVar(&gluster.VgName, "vgName", "", "Specify which vg is used for the pool")
	flag.StringVar(&gluster.BasePath, "basePath", "", "Specify basepath for gluster gluster")
	flag.StringVar(&gluster.Secret, "secret", "", "Specify the secret for communication on the /sec/ endpoints")
	flag.Parse()

	if (len(gluster.BasePath) == 0 || len(gluster.PoolName) == 0 || len(gluster.VgName) == 0 || len(gluster.Secret) == 0) {
		log.Fatal("Must specify parameters 'poolName', 'basePath', 'vgName' and 'secret'")
	}
}

func main() {
	gin.SetMode(gin.DebugMode)

	r := gin.New()
	r.Use(gin.Recovery())

	// TODO: Usage endpoint

	sec := r.Group("/sec", gin.BasicAuth(gin.Accounts{
		"GLUSTER_API":    gluster.Secret,
	}))

	// /sec/volume = Create all the necessary things on all gluster servers for a new volume
	sec.POST("/volume", func(c *gin.Context) {
		if err := gluster.CreateVolume("ose-mon-a", "101M"); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": err.Error(),
			})
		} else {
			c.JSON(http.StatusOK, gin.H{
				"message": "Volume created",
			})
		}
	})

	// /sec/lv = Create LV on local server
	sec.POST("/lv", func(c *gin.Context) {
		var json models.CreateLVCommand
		if c.BindJSON(&json) == nil {
			if err := gluster.CreateLvOnPool(json.Size, json.MountPoint, json.LvName); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"message": err.Error(),
				})
			} else {
				c.JSON(http.StatusOK, gin.H{
					"message": "LV created",
				})
			}
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"message": wrongApiUsageError})
		}
	})

	r.Run(":" + strconv.Itoa(gluster.Port))
}