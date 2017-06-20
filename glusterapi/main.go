package main

import (
	"flag"
	"log"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/oscp/cloud-selfservice-portal/glusterapi/gluster"
)

func init() {
	flag.IntVar(&gluster.Port, "port", 8080, "Specify the api-port")
	flag.IntVar(&gluster.MaxGB, "maxGB", 100, "Max GB a user can order per volume")
	flag.IntVar(&gluster.MaxMB, "maxMB", 1024, "Max MB a user can order per volume")
	flag.IntVar(&gluster.Replicas, "replicas", 2, "Define the replica count for new volumes")
	flag.StringVar(&gluster.PoolName, "poolName", "", "Specify which lvm pool should be used for orders")
	flag.StringVar(&gluster.VgName, "vgName", "", "Specify which vg is used for the pool")
	flag.StringVar(&gluster.BasePath, "basePath", "", "Specify basepath for gluster gluster")
	flag.StringVar(&gluster.Secret, "secret", "", "Specify the secret for communication on the /sec/ endpoints")
	flag.Parse()

	if len(gluster.BasePath) == 0 || len(gluster.PoolName) == 0 || len(gluster.VgName) == 0 || len(gluster.Secret) == 0 {
		log.Fatal("Must specify parameters 'poolName', 'basePath', 'vgName' and 'secret'")
	}
}

func main() {
	gin.SetMode(gin.ReleaseMode)

	r := gin.New()
	r.Use(gin.Recovery())

	// Public endpoint for volume monitoring
	r.GET("/volume/:pvname", gluster.VolumeInfoHandler)
	r.GET("/volume/:pvname/check", gluster.CheckVolumeHandler)

	// Secured endpoints with basic auth
	sec := r.Group("/sec", gin.BasicAuth(gin.Accounts{
		"GLUSTER_API": gluster.Secret,
	}))
	// /sec/volume 		= Create all the necessary things on all gluster servers for a new volume
	// /sec/volume/grow 	= Grows an existing volume on all the gluster servers
	// /sec/lv 		= Create LV on local server
	// /sec/lv/grow 	= Grows an existing LV on the local server
	sec.POST("/volume", gluster.CreateVolumeHandler)
	sec.POST("/lv", gluster.CreateLVHandler)
	sec.POST("/volume/grow", gluster.GrowVolumeHandler)
	sec.POST("/lv/grow", gluster.GrowLVHandler)

	log.Printf("Gluster api is running on: %v", gluster.Port)
	r.Run(":" + strconv.Itoa(gluster.Port))
}
