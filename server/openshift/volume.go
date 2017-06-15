package openshift

import (
	"errors"
	"net/http"

	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/oscp/cloud-selfservice-portal/glusterapi/models"
	"github.com/oscp/cloud-selfservice-portal/server/common"
)

func newVolumeHandler(c *gin.Context) {
	project := c.PostForm("project")
	size := strings.ToUpper(c.PostForm("size"))
	pvcName := c.PostForm("pvcname")
	username := common.GetUserName(c)

	project = "test"
	pvcName = "meinpvc"
	size = "1000G"

	if err := validateNewVolume(project, size, pvcName); err != nil {
		c.HTML(http.StatusOK, newVoluemeURL, gin.H{
			"Error": err.Error(),
		})
		return
	}

	if err := createNewVolume(project, username, size, pvcName); err != nil {
		c.HTML(http.StatusOK, newVoluemeURL, gin.H{
			"Error": err.Error(),
		})
	} else {
		c.HTML(http.StatusOK, newVoluemeURL, gin.H{
			"Success": "Das Volume wurde erstellt. Deinem Projekt wurde das PVC, und der Gluster Service & Endpunkte hinzugefügt.",
		})
	}
}

func validateNewVolume(project string, size string, pvcName string) error {
	if len(project) == 0 || len(pvcName) == 0 || len(size) == 0 {
		return errors.New("Projektname / PVC-Name und Grösse müssen angegeben werden")
	}

	return nil
}

func createNewVolume(project string, username string, size string, pvcName string) error {
	if err := createGlusterVolume(project, size, username); err != nil {
		return err
	}

	return nil
}

func createGlusterVolume(project string, size string, username string) error {
	cmd := models.CreateVolumeCommand{
		Project: project,
		Size:    size,
	}

	b := new(bytes.Buffer)
	if err := json.NewEncoder(b).Encode(cmd); err != nil {
		log.Println(err.Error())
		return errors.New(genericAPIError)
	}

	client, req := getGlusterHTTPClient("sec/volume", b)

	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		log.Printf("%v created a gluster volume. Project: %v, size: %v", username, project, size)
		return nil
	} else {
		errMsg, _ := ioutil.ReadAll(resp.Body)
		log.Println("Error creating gluster volume:", err, resp.StatusCode, string(errMsg))

		return fmt.Errorf("Fehlerhafte Antwort vom Gluster-API: %v", string(errMsg))
	}
}
