package main

import (
	"log"
	"flag"
	"github.com/gin-gonic/gin"
	"github.com/oscp/openshift-selfservice/glusterapi/volumes"
	"strconv"
)

var port int

func init() {
	log.Println("Starting gluster api")

	flag.IntVar(&port, "port", 8080, "Specify the api-port")
	flag.IntVar(&volumes.MaxGB, "maxGB", 100, "Max GB a user can order per volume")
	flag.IntVar(&volumes.MaxMB, "maxMB", 1024, "Max MB a user can order per volume")
	flag.StringVar(&volumes.PoolName, "poolName", "", "Specify which lvm pool should be used for orders")
	flag.StringVar(&volumes.VgName, "vgName", "", "Specify which vg is used for the pool")
	flag.StringVar(&volumes.BasePath, "basePath", "", "Specify basepath for gluster volumes")
	flag.Parse()

	if (len(volumes.BasePath) == 0 || len(volumes.PoolName) == 0 || len(volumes.VgName) == 0) {
		log.Fatal("Must specify parameters 'poolName' and 'basePath' and 'vgName'")
	}
}

func main() {
	gin.SetMode(gin.ReleaseMode)

	r := gin.New()
	r.Use(gin.Recovery())

	// TODO: Usage endpoint
	// TODO: Add auth-token to this
	r.GET("/", func(c *gin.Context) {
		isOk, msg := volumes.CreateVolume("ose-mon-a", "101M")

		c.JSON(200, gin.H{
			"isOk": isOk,
			"message": msg,
		})
	})
	r.Run(":" + strconv.Itoa(port))
}