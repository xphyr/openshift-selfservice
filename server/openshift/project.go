package openshift

import (
	"github.com/gin-gonic/gin"
	"github.com/oscp/openshift-selfservice/server/common"
	"net/http"
	"bytes"
	"log"
	"io/ioutil"
	"strings"
	"github.com/Jeffail/gabs"
	"errors"
)

func newProjectHandler(c *gin.Context) {
	project := c.PostForm("project")
	billing := c.PostForm("billing")
	megaid := c.PostForm("megaid")
	username := common.GetUserName(c)

	if err := validateNewProject(project, billing, false); err != nil {
		c.HTML(http.StatusOK, newProjectUrl, gin.H{
			"Error": err.Error(),
		})
		return
	}

	if err := createNewProject(project, username, billing, megaid); err != nil {
		c.HTML(http.StatusOK, newProjectUrl, gin.H{
			"Success": "Das Projekt wurde erstellt",
		})
	} else {
		c.HTML(http.StatusOK, newProjectUrl, gin.H{
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
		c.HTML(http.StatusOK, newTestProjectUrl, gin.H{
			"Error": err.Error(),
			"User": common.GetUserName(c),
		})
		return
	}
	if err := createNewProject(project, username, billing, ""); err != nil {
		c.HTML(http.StatusOK, newTestProjectUrl, gin.H{
			"Error": err.Error(),
			"User": common.GetUserName(c),
		})
	} else {
		c.HTML(http.StatusOK, newTestProjectUrl, gin.H{
			"Success": "Das Test-Projekt wurde erstellt",
			"User": common.GetUserName(c),
		})
	}
}

func updateBillingHandler(c *gin.Context) {
	username := common.GetUserName(c)
	project := c.PostForm("project")
	billing := c.PostForm("billing")

	if err := validateBillingInformation(project, billing, username); err != nil {
		c.HTML(http.StatusOK, updateBillingUrl, gin.H{
			"Error": err.Error(),
		})
		return
	}

	if err := createOrUpdateMetadata(project, billing, "", username); err != nil {
		c.HTML(http.StatusOK, updateBillingUrl, gin.H{
			"Error": err.Error(),
		})
	} else {
		c.HTML(http.StatusOK, updateBillingUrl, gin.H{
			"Success": "Die neuen Daten wurden gespeichert",
		})
	}
}

func validateNewProject(project string, billing string, isTestproject bool) (error) {
	if (len(project) == 0) {
		return errors.New("Projektname muss angegeben werden")
	}

	if (!isTestproject && len(billing) == 0) {
		return errors.New("Kontierungsnummer muss angegeben werden")
	}

	return nil
}

func validateBillingInformation(project string, billing string, username string) (error) {
	if (len(project) == 0) {
		return errors.New("Projektname muss angegeben werden")
	}

	if (len(billing) == 0) {
		return errors.New("Kontierungsnummer muss angegeben werden")
	}

	// Validate permissions
	if err := checkAdminPermissions(username, project); err != nil {
		return err
	}

	return nil
}

func createNewProject(project string, username string, billing string, megaid string) (error) {
	p := newObjectRequest("ProjectRequest", project)

	client, req := getOseHttpClient("POST",
		"oapi/v1/projectrequests",
		bytes.NewReader(p.Bytes()))

	resp, err := client.Do(req)
	defer resp.Body.Close()

	if (resp.StatusCode == http.StatusCreated) {
		log.Print(username + " created a new project: " + project)

		if err := changeProjectPermission(project, username); err != nil {
			return err
		}

		if err := createOrUpdateMetadata(project, billing, megaid, username); err != nil {
			return err
		} else {
			return nil
		}
	} else {
		if (resp.StatusCode == http.StatusConflict) {
			return errors.New("Das Projekt existiert bereits")
		}

		errMsg, _ := ioutil.ReadAll(resp.Body)
		log.Println("Error creating new project:", err, resp.StatusCode, string(errMsg))

		return errors.New(genericApiError)
	}
}

func changeProjectPermission(project string, username string) (error) {
	// Get existing policybindings
	policyBindings, err := getPolicyBindings(project)

	if (policyBindings == nil) {
		return err
	}

	children, err := policyBindings.S("roleBindings").Children()
	if (err != nil) {
		log.Println("Unable to parse roleBindings", err.Error())
		return errors.New(genericApiError)
	}
	for _, v := range children {
		if (v.Path("name").Data().(string) == "admin") {
			v.ArrayAppend(strings.ToLower(username), "roleBinding", "userNames")
			v.ArrayAppend(strings.ToUpper(username), "roleBinding", "userNames")
		}
	}

	// Update the policyBindings on the api
	client, req := getOseHttpClient("PUT",
		"oapi/v1/namespaces/" + project + "/policybindings/:default",
		bytes.NewReader(policyBindings.Bytes()))

	resp, err := client.Do(req)
	defer resp.Body.Close()

	if (err != nil) {
		log.Println("Error from server: ", err.Error())
		return errors.New(genericApiError)
	}

	if (resp.StatusCode == http.StatusOK) {
		log.Print(username + " is now admin of " + project)
		return nil
	} else {
		errMsg, _ := ioutil.ReadAll(resp.Body)
		log.Println("Error updating project permissions:", err, resp.StatusCode, string(errMsg))
		return errors.New(genericApiError)
	}
}

func createOrUpdateMetadata(project string, billing string, megaid string, username string) (error) {
	client, req := getOseHttpClient("GET", "api/v1/namespaces/" + project, nil)
	resp, err := client.Do(req)
	defer resp.Body.Close()

	if (err != nil) {
		log.Println("Error from server: ", err.Error())
		return errors.New(genericApiError)
	}

	json, err := gabs.ParseJSONBuffer(resp.Body)
	if err != nil {
		log.Println("error decoding json:", err, resp.StatusCode)
		return errors.New(genericApiError)
	}

	annotations := json.Path("metadata.annotations")
	annotations.Set(billing, "openshift.io/kontierung-element")
	annotations.Set(username, "openshift.io/requester")

	if (len(megaid) > 0) {
		annotations.Set(megaid, "openshift.io/MEGAID")
	}

	client, req = getOseHttpClient("PUT",
		"api/v1/namespaces/" + project,
		bytes.NewReader(json.Bytes()))

	resp, err = client.Do(req)
	defer resp.Body.Close()

	if (resp.StatusCode == http.StatusOK) {
		log.Println("User " + username + " changed changed config of project project " + project + ". Kontierungsnummer: " + billing, ", MegaID: " + megaid)
		return nil
	} else {
		errMsg, _ := ioutil.ReadAll(resp.Body)
		log.Println("Error updating project config:", err, resp.StatusCode, string(errMsg))

		return errors.New(genericApiError)
	}
}
