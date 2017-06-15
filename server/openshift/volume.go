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

	"github.com/Jeffail/gabs"
	"github.com/gin-gonic/gin"
	"github.com/oscp/cloud-selfservice-portal/glusterapi/models"
	"github.com/oscp/cloud-selfservice-portal/server/common"
)

func newVolumeHandler(c *gin.Context) {
	project := c.PostForm("project")
	size := strings.ToUpper(c.PostForm("size"))
	pvcName := c.PostForm("pvcname")
	mode := c.PostForm("mode")
	username := common.GetUserName(c)

	project = "test"
	pvcName = "meinpvc"
	size = "100M"

	if err := validateNewVolume(project, size, pvcName, mode, username); err != nil {
		c.HTML(http.StatusOK, newVoluemeURL, gin.H{
			"Error": err.Error(),
		})
		return
	}

	if err := createNewVolume(project, username, size, pvcName, mode); err != nil {
		c.HTML(http.StatusOK, newVoluemeURL, gin.H{
			"Error": err.Error(),
		})
	} else {
		c.HTML(http.StatusOK, newVoluemeURL, gin.H{
			"Success": "Das Volume wurde erstellt. Deinem Projekt wurde das PVC, und der Gluster Service & Endpunkte hinzugefügt.",
		})
	}
}

func validateNewVolume(project string, size string, pvcName string, mode string, username string) error {
	if len(project) == 0 || len(pvcName) == 0 || len(size) == 0 || len(mode) == 0 {
		return errors.New("Es müssen alle Felder ausgefüllt werden")
	}

	// Validate permissions
	if err := checkAdminPermissions(username, project); err != nil {
		return err
	}

	return nil
}

func createNewVolume(project string, username string, size string, pvcName string, mode string) error {
	//if pvName, err := createGlusterVolume(project, size, username); err != nil {
	//	return err
	//}

	pvName := "test_pv5"
	if err := createOpenShiftPVandPVC(project, size, pvName, pvcName, mode, username); err != nil {
		return err
	}

	// Create Gluster Service & Endpoints in user project

	return nil
}

func createGlusterVolume(project string, size string, username string) (string, error) {
	cmd := models.CreateVolumeCommand{
		Project: project,
		Size:    size,
	}

	b := new(bytes.Buffer)
	if err := json.NewEncoder(b).Encode(cmd); err != nil {
		log.Println(err.Error())
		return "", errors.New(genericAPIError)
	}

	client, req := getGlusterHTTPClient("sec/volume", b)

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		log.Printf("%v created a gluster volume. Project: %v, size: %v", username, project, size)

		json, err := gabs.ParseJSONBuffer(resp.Body)
		if err != nil {
			log.Println("Error parsing json from gluster-api response", err.Error())
			return "", errors.New(genericAPIError)
		}

		return json.Path("message").Data().(string), nil
	} else {
		errMsg, _ := ioutil.ReadAll(resp.Body)
		log.Println("Error creating gluster volume:", err, resp.StatusCode, string(errMsg))

		return "", fmt.Errorf("Fehlerhafte Antwort vom Gluster-API: %v", string(errMsg))
	}
}

func createOpenShiftPVandPVC(project string, size string, pvName string, pvcName string, mode string, username string) error {
	p := newObjectRequest("PersistentVolume", strings.Replace(pvName, "_", "-", -1))

	p.SetP(size, "spec.capacity.storage")
	p.SetP("glusterfs-cluster", "spec.glusterfs.endpoints")
	p.SetP(fmt.Sprintf("vol_%v", pvName), "spec.glusterfs.path")
	p.SetP(false, "spec.glusterfs.readOnly")
	p.SetP("Retain", "spec.persistentVolumeReclaimPolicy")

	p.ArrayP("spec.accessModes")
	p.ArrayAppend(mode, "spec", "accessModes")

	log.Println(p.String())

	client, req := getOseHTTPClient("POST",
		"api/v1/persistentvolumes",
		bytes.NewReader(p.Bytes()))

	resp, err := client.Do(req)
	if err == nil {
		defer resp.Body.Close()
	}

	if resp.StatusCode == http.StatusCreated {
		log.Printf("Created the pv %v because of the request of %v", pvName, username)

		// Create pvc now

		return nil
	}

	errMsg, _ := ioutil.ReadAll(resp.Body)
	log.Println("Error creating new PV:", err, resp.StatusCode, string(errMsg))

	return errors.New(genericAPIError)
}
