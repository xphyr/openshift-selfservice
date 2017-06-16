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

	"os"
	"strconv"

	"github.com/Jeffail/gabs"
	"github.com/gin-gonic/gin"
	"github.com/oscp/cloud-selfservice-portal/glusterapi/models"
	"github.com/oscp/cloud-selfservice-portal/server/common"
)

const (
	wrongSizeError = "Ungültige Grösse. Format muss Zahl gefolgt von M/G sein (z.B. 100M). Maximale erlaubte Grössen sind: M: %v, G: %v"
)

func newVolumeHandler(c *gin.Context) {
	project := c.PostForm("project")
	size := strings.ToUpper(c.PostForm("size"))
	pvcName := c.PostForm("pvcname")
	mode := c.PostForm("mode")
	username := common.GetUserName(c)

	if err := validateNewVolume(project, size, pvcName, mode, username); err != nil {
		c.HTML(http.StatusOK, newVolumeURL, gin.H{
			"Error": err.Error(),
		})
		return
	}

	if err := createNewVolume(project, username, size, pvcName, mode); err != nil {
		c.HTML(http.StatusOK, newVolumeURL, gin.H{
			"Error": err.Error(),
		})
	} else {
		c.HTML(http.StatusOK, newVolumeURL, gin.H{
			"Success": "Das Volume wurde erstellt. Deinem Projekt wurde das PVC, und der Gluster Service & Endpunkte hinzugefügt.",
		})
	}
}

func fixVolumeHandler(c *gin.Context) {
	project := c.PostForm("project")
	username := common.GetUserName(c)

	if err := validateFixVolume(project, username); err != nil {
		c.HTML(http.StatusOK, fixVolumeURL, gin.H{
			"Error": err.Error(),
		})
		return
	}

	if err := recreateGlusterObjects(project, username); err != nil {
		c.HTML(http.StatusOK, fixVolumeURL, gin.H{
			"Error": err.Error(),
		})
	} else {
		c.HTML(http.StatusOK, fixVolumeURL, gin.H{
			"Success": "Die Gluster-Objekte wurden in deinem Projekt erzeugt.",
		})
	}
}

func validateNewVolume(project string, size string, pvcName string, mode string, username string) error {
	// Required fields
	if len(project) == 0 || len(pvcName) == 0 || len(size) == 0 || len(mode) == 0 {
		return errors.New("Es müssen alle Felder ausgefüllt werden")
	}

	// Size limits
	maxMB, maxGB := getMaxSizes()
	sizeOk := false
	if strings.HasSuffix(size, "M") {
		sizeInt, err := strconv.Atoi(strings.Replace(size, "M", "", 1))
		if err != nil {
			return fmt.Errorf(wrongSizeError, maxMB, maxGB)
		}

		if sizeInt <= maxMB {
			sizeOk = true
		} else {
			return errors.New("Deine Angaben sind zu gross für 'M'. Bitte gib die Grösse als Ganzzahl in 'G' an")
		}
	}
	if strings.HasSuffix(size, "G") {
		sizeInt, err := strconv.Atoi(strings.Replace(size, "G", "", 1))
		if err != nil {
			return fmt.Errorf(wrongSizeError, maxMB, maxGB)
		}

		if sizeInt <= maxGB {
			sizeOk = true
		}
	}

	if !sizeOk {
		return fmt.Errorf(wrongSizeError, maxMB, maxGB)
	}

	// Permissions on project
	if err := checkAdminPermissions(username, project); err != nil {
		return err
	}

	return nil
}

func validateFixVolume(project string, username string) error {
	if len(project) == 0 {
		return errors.New("Projekt muss angegeben werden")
	}

	// Permissions on project
	if err := checkAdminPermissions(username, project); err != nil {
		return err
	}

	return nil
}

func getMaxSizes() (int, int) {
	maxMB := os.Getenv("MAX_MB")
	maxGB := os.Getenv("MAX_GB")

	maxMBInt, errMB := strconv.Atoi(maxMB)
	maxGBInt, errGB := strconv.Atoi(maxGB)

	if errMB != nil || errGB != nil || maxMBInt <= 0 || maxGBInt <= 0 {
		log.Fatal("Env variables 'MAX_MB' and 'MAX_GB' must be specified and a valid integer")
	}
	return maxMBInt, maxGBInt
}

func createNewVolume(project string, username string, size string, pvcName string, mode string) error {
	pvName, err := createGlusterVolume(project, size, username)
	if err != nil {
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

		respJson, err := gabs.ParseJSONBuffer(resp.Body)
		if err != nil {
			log.Println("Error parsing respJson from gluster-api response", err.Error())
			return "", errors.New(genericAPIError)
		}

		// Add gl_ to pvName because of conflicting PVs on other storage technology
		return fmt.Sprintf("gl_%v", respJson.Path("message").Data().(string)), nil
	}

	errMsg, _ := ioutil.ReadAll(resp.Body)
	log.Println("Error creating gluster volume:", err, resp.StatusCode, string(errMsg))

	return "", fmt.Errorf("Fehlerhafte Antwort vom Gluster-API: %v", string(errMsg))
}

func createOpenShiftPV(size string, pvName string, mode string, username string) error {
	p := newObjectRequest("PersistentVolume", strings.Replace(pvName, "_", "-", -1))

	p.SetP(size, "spec.capacity.storage")
	p.SetP("glusterfs-cluster", "spec.glusterfs.endpoints")
	p.SetP(strings.Replace(pvName, "gl_", "vol_", 1), "spec.glusterfs.path")
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

func recreateGlusterObjects(project string, username string) error {
	if err := createOpenShiftGlusterService(project, username); err != nil {
		return err
	}

	if err := createOpenShiftGlusterEndpoint(project, username); err != nil {
		return err
	}

	return nil
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

	if resp.StatusCode == http.StatusConflict {
		log.Println("Gluster service already existed, skipping")
		return nil
	}

	errMsg, _ := ioutil.ReadAll(resp.Body)
	log.Println("Error creating gluster service:", err, resp.StatusCode, string(errMsg))

	return errors.New(genericAPIError)
}

func createOpenShiftGlusterEndpoint(project string, username string) error {
	p, err := getGlusterEndpointsContainer()
	if err != nil {
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

	if resp.StatusCode == http.StatusConflict {
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
	if len(glusterIPs) == 0 {
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
