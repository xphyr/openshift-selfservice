package openshift

import (
	"bytes"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/Jeffail/gabs"
	"github.com/gin-gonic/gin"
	"github.com/oscp/cloud-selfservice-portal/server/common"
)

func newProjectHandler(c *gin.Context) {
	project := c.PostForm("project")
	billing := c.PostForm("billing")
	megaid := c.PostForm("megaid")
	username := common.GetUserName(c)

	if err := validateNewProject(project, billing, false); err != nil {
		c.HTML(http.StatusOK, newProjectURL, gin.H{
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

func newTestProjectHandler(c *gin.Context) {
	project := c.PostForm("project")
	username := common.GetUserName(c)

	// Special values for a test project
	billing := "keine-verrechnung"
	project = username + "-" + project

	if err := validateNewProject(project, billing, true); err != nil {
		c.HTML(http.StatusOK, newTestProjectURL, gin.H{
			"Error": err.Error(),
			"User":  common.GetUserName(c),
		})
		return
	}
	if err := createNewProject(project, username, billing, ""); err != nil {
		c.HTML(http.StatusOK, newTestProjectURL, gin.H{
			"Error": err.Error(),
			"User":  common.GetUserName(c),
		})
	} else {
		c.HTML(http.StatusOK, newTestProjectURL, gin.H{
			"Success": "Das Test-Projekt wurde erstellt",
			"User":    common.GetUserName(c),
		})
	}
}

func updateBillingHandler(c *gin.Context) {
	username := common.GetUserName(c)
	project := c.PostForm("project")
	billing := c.PostForm("billing")

	if err := validateBillingInformation(project, billing, username); err != nil {
		c.HTML(http.StatusOK, updateBillingURL, gin.H{
			"Error": err.Error(),
		})
		return
	}

	if err := createOrUpdateMetadata(project, billing, "", username); err != nil {
		c.HTML(http.StatusOK, updateBillingURL, gin.H{
			"Error": err.Error(),
		})
	} else {
		c.HTML(http.StatusOK, updateBillingURL, gin.H{
			"Success": "Die neuen Daten wurden gespeichert",
		})
	}
}

func validateNewProject(project string, billing string, isTestproject bool) error {
	if len(project) == 0 {
		return errors.New("Projektname muss angegeben werden")
	}

	if !isTestproject && len(billing) == 0 {
		return errors.New("Kontierungsnummer muss angegeben werden")
	}

	return nil
}

func validateBillingInformation(project string, billing string, username string) error {
	if len(project) == 0 {
		return errors.New("Projektname muss angegeben werden")
	}

	if len(billing) == 0 {
		return errors.New("Kontierungsnummer muss angegeben werden")
	}

	// Validate permissions
	if err := checkAdminPermissions(username, project); err != nil {
		return err
	}

	return nil
}

func createNewProject(project string, username string, billing string, megaid string) error {
	p := newObjectRequest("ProjectRequest", project)

	client, req := getOseHTTPClient("POST",
		"oapi/v1/projectrequests",
		bytes.NewReader(p.Bytes()))

	resp, err := client.Do(req)
	if err == nil {
		defer resp.Body.Close()
	}

	if resp.StatusCode == http.StatusCreated {
		log.Print(username + " created a new project: " + project)

		if err := changeProjectPermission(project, username); err != nil {
			return err
		}

		if err := createOrUpdateMetadata(project, billing, megaid, username); err != nil {
			return err
		}
		return nil
	}
	if resp.StatusCode == http.StatusConflict {
		return errors.New("Das Projekt existiert bereits")
	}

	errMsg, _ := ioutil.ReadAll(resp.Body)
	log.Println("Error creating new project:", err, resp.StatusCode, string(errMsg))

	return errors.New(genericAPIError)
}

func changeProjectPermission(project string, username string) error {
	// Get existing policybindings
	policyBindings, err := getPolicyBindings(project)

	if policyBindings == nil {
		return err
	}

	children, err := policyBindings.S("roleBindings").Children()
	if err != nil {
		log.Println("Unable to parse roleBindings", err.Error())
		return errors.New(genericAPIError)
	}
	for _, v := range children {
		if v.Path("name").Data().(string) == "admin" {
			v.ArrayAppend(strings.ToLower(username), "roleBinding", "userNames")
			v.ArrayAppend(strings.ToUpper(username), "roleBinding", "userNames")
		}
	}

	// Update the policyBindings on the api
	client, req := getOseHTTPClient("PUT",
		"oapi/v1/namespaces/"+project+"/policybindings/:default",
		bytes.NewReader(policyBindings.Bytes()))

	resp, err := client.Do(req)
	if err != nil {
		log.Println("Error from server: ", err.Error())
		return errors.New(genericAPIError)
	}

	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		log.Print(username + " is now admin of " + project)
		return nil
	}

	errMsg, _ := ioutil.ReadAll(resp.Body)
	log.Println("Error updating project permissions:", err, resp.StatusCode, string(errMsg))
	return errors.New(genericAPIError)
}

func createOrUpdateMetadata(project string, billing string, megaid string, username string) error {
	client, req := getOseHTTPClient("GET", "api/v1/namespaces/"+project, nil)
	resp, err := client.Do(req)
	if err != nil {
		log.Println("Error from server: ", err.Error())
		return errors.New(genericAPIError)
	}

	defer resp.Body.Close()

	json, err := gabs.ParseJSONBuffer(resp.Body)
	if err != nil {
		log.Println("error decoding json:", err, resp.StatusCode)
		return errors.New(genericAPIError)
	}

	annotations := json.Path("metadata.annotations")
	annotations.Set(billing, "openshift.io/kontierung-element")
	annotations.Set(username, "openshift.io/requester")

	if len(megaid) > 0 {
		annotations.Set(megaid, "openshift.io/MEGAID")
	}

	client, req = getOseHTTPClient("PUT",
		"api/v1/namespaces/"+project,
		bytes.NewReader(json.Bytes()))

	resp, err = client.Do(req)

	if resp.StatusCode == http.StatusOK {
		resp.Body.Close()
		log.Println("User "+username+" changed changed config of project project "+project+". Kontierungsnummer: "+billing, ", MegaID: "+megaid)
		return nil
	}

	errMsg, _ := ioutil.ReadAll(resp.Body)
	log.Println("Error updating project config:", err, resp.StatusCode, string(errMsg))

	return errors.New(genericAPIError)
}
