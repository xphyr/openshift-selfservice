package openshift

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/oscp/cloud-selfservice-portal/server/common"
)

func newVolumeHandler(c *gin.Context) {
	project := c.PostForm("project")
	size := c.PostForm("size")
	pvcName := c.PostForm("pvcname")
	username := common.GetUserName(c)

	if err := validateNewVolume(project, size, pvcName); err != nil {
		c.HTML(http.StatusOK, newVoluemURL, gin.H{
			"Error": err.Error(),
		})
		return
	}

	if err := createNewProject(project, username, billing, megaid); err != nil {
		c.HTML(http.StatusOK, newProjectURL, gin.H{
			"Success": "Das Projekt wurde erstellt",
		})
	} else {
		c.HTML(http.StatusOK, newProjectURL, gin.H{
			"Error": err.Error(),
		})
	}
}

func validateNewVolume(project string, size string, pvcName string) error {
	if len(project) == 0 || len(pvcName) == 0 || len(size) == 0 {
		return errors.New("Projektname / PVC-Name und Grösse müssen angegeben werden")
	}

	return nil
}

func createNewVolume() error {

}
