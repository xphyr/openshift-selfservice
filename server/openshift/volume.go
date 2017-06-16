package openshift

import (
	"errors"
	"net/http"

	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"strings"

	"encoding/json"

	"github.com/Jeffail/gabs"
	"github.com/gin-gonic/gin"
	"github.com/oscp/cloud-selfservice-portal/glusterapi/models"
	"github.com/oscp/cloud-selfservice-portal/server/common"
	"os"
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

	// Todo validate me better

	// Validate permissions
	if err := checkAdminPermissions(username, project); err != nil {
		return err
	}

	return nil
}

func createNewVolume(project string, username string, size string, pvcName string, mode string) error {
	pvName, err := createGlusterVolume(project, size, username);
	if (err != nil) {
		return err
	}

	if err := createOpenShiftPV(size, pvName, mode, username); err != nil {
		return err
	}

	if err := createOpenShiftPVC(project, size, pvcName, mode, username); err != nil {
		return err
	}

	// Create Gluster Service & Endpoints in user project
	if err := createOpenShiftGlusterService(project, username); err != nil {
		return err
	}
	if err := createOpenShiftGlusterEndpoint(project, username); err != nil {
		return err
	}

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
	}

	errMsg, _ := ioutil.ReadAll(resp.Body)
	log.Println("Error creating gluster volume:", err, resp.StatusCode, string(errMsg))

	return "", fmt.Errorf("Fehlerhafte Antwort vom Gluster-API: %v", string(errMsg))
}

func createOpenShiftPV(size string, pvName string, mode string, username string) error {
	p := newObjectRequest("PersistentVolume", strings.Replace(pvName, "_", "-", -1))

	p.SetP(size, "spec.capacity.storage")
	p.SetP("glusterfs-cluster", "spec.glusterfs.endpoints")
	p.SetP(fmt.Sprintf("vol_%v", pvName), "spec.glusterfs.path")
	p.SetP(false, "spec.glusterfs.readOnly")
	p.SetP("Retain", "spec.persistentVolumeReclaimPolicy")

	p.ArrayP("spec.accessModes")
	p.ArrayAppend(mode, "spec", "accessModes")

	client, req := getOseHTTPClient("POST",
		"api/v1/persistentvolumes",
		bytes.NewReader(p.Bytes()))

	resp, err := client.Do(req)
	if err == nil {
		defer resp.Body.Close()
	}

	if resp.StatusCode == http.StatusCreated {
		log.Printf("Created the pv %v based on the request of %v", pvName, username)
		return nil
	}

	errMsg, _ := ioutil.ReadAll(resp.Body)
	log.Println("Error creating new PV:", err, resp.StatusCode, string(errMsg))

	return errors.New(genericAPIError)
}

func createOpenShiftPVC(project string, size string, pvcName string, mode string, username string) error {
	p := newObjectRequest("PersistentVolumeClaim", strings.Replace(pvcName, "_", "-", -1))

	p.SetP(size, "spec.resources.requests.storage")
	p.ArrayP("spec.accessModes")
	p.ArrayAppend(mode, "spec", "accessModes")

	client, req := getOseHTTPClient("POST",
		"api/v1/namespaces/"+project+"/persistentvolumeclaims",
		bytes.NewReader(p.Bytes()))

	resp, err := client.Do(req)
	if err == nil {
		defer resp.Body.Close()
	}

	if resp.StatusCode == http.StatusCreated {
		log.Printf("Created the pvc %v based on the request of %v", pvcName, username)
		return nil
	}

	errMsg, _ := ioutil.ReadAll(resp.Body)
	log.Println("Error creating new PVC:", err, resp.StatusCode, string(errMsg))

	return errors.New(genericAPIError)
}

func createOpenShiftGlusterService(project string, username string) error {
	p := newObjectRequest("Service", "glusterfs-cluster")

	port := gabs.New()
	port.Set(1, "port")

	p.ArrayP("spec.ports")
	p.ArrayAppendP(port.Data(), "spec.ports")

	client, req := getOseHTTPClient("POST",
		"api/v1/namespaces/"+project+"/services",
		bytes.NewReader(p.Bytes()))

	resp, err := client.Do(req)
	if err == nil {
		defer resp.Body.Close()
	}

	if resp.StatusCode == http.StatusCreated {
		log.Printf("Created the gluster service based on the request of %v", username)
		return nil
	}

	if (resp.StatusCode == http.StatusConflict) {
		log.Println("Gluster service already existed, skipping")
		return nil
	}

	errMsg, _ := ioutil.ReadAll(resp.Body)
	log.Println("Error creating gluster service:", err, resp.StatusCode, string(errMsg))

	return errors.New(genericAPIError)
}

func createOpenShiftGlusterEndpoint(project string, username string) error {
	p, err := getGlusterEndpointsContainer()
	if (err != nil) {
		return err
	}

	client, req := getOseHTTPClient("POST",
		"api/v1/namespaces/"+project+"/endpoints",
		bytes.NewReader(p.Bytes()))

	resp, err := client.Do(req)
	if err == nil {
		defer resp.Body.Close()
	}

	if resp.StatusCode == http.StatusCreated {
		log.Printf("Created the gluster endpoints based on the request of %v", username)
		return nil
	}

	if (resp.StatusCode == http.StatusConflict) {
		log.Println("Gluster endpoints already existed, skipping")
		return nil
	}

	errMsg, _ := ioutil.ReadAll(resp.Body)
	log.Println("Error creating gluster endpoints:", err, resp.StatusCode, string(errMsg))

	return errors.New(genericAPIError)
}

func getGlusterEndpointsContainer() (*gabs.Container, error) {
	p := newObjectRequest("Endpoints", "glusterfs-cluster")
	p.Array("subsets")

	// Add gluster endpoints
	glusterIPs := os.Getenv("GLUSTER_IPS")
	if (len(glusterIPs) == 0) {
		log.Println("Wrong configuration. Missing env variable 'GLUSTER_IPS'")
		return nil, errors.New(genericAPIError)
	}

	addresses := gabs.New()
	addresses.Array("addresses")
	addresses.Array("ports")
	for _, ip := range strings.Split(glusterIPs, ",") {
		address := gabs.New()
		address.Set(ip, "ip")

		addresses.ArrayAppend(address.Data(), "addresses")
	}

	port := gabs.New()
	port.Set(1, "port")
	addresses.ArrayAppend(port.Data(), "ports")

	p.ArrayAppend(addresses.Data(), "subsets")

	return p, nil
}
