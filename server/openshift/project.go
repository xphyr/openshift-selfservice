package openshift

import (
	"github.com/gin-gonic/gin"
	"github.com/oscp/openshift-selfservice/server/common"
	"net/http"
	"encoding/json"
	"bytes"
	"log"
	"io/ioutil"
	"strings"
	"github.com/oscp/openshift-selfservice/server/models"
	"github.com/Jeffail/gabs"
)

func newProjectHandler(c *gin.Context) {
	project := c.PostForm("project")
	billing := c.PostForm("billing")
	megaid := c.PostForm("megaid")
	username := common.GetUserName(c)

	isOk, msg := validateNewProject(project, billing, false)
	if (!isOk) {
		c.HTML(http.StatusOK, newProjectUrl, gin.H{
			"Error": msg,
		})
		return
	}

	isOk, msg = createNewProject(project, username, billing, megaid)
	if (!isOk) {
		c.HTML(http.StatusOK, newProjectUrl, gin.H{
			"Error": msg,
		})
	} else {
		c.HTML(http.StatusOK, newProjectUrl, gin.H{
			"Success": msg,
		})
	}
}

func newTestProjectHandler(c *gin.Context) {
	project := c.PostForm("project")
	username := common.GetUserName(c)

	// Special values for a test project
	billing := "keine-verrechnung"
	project = username + "-" + project

	isOk, msg := validateNewProject(project, billing, true)
	if (!isOk) {
		c.HTML(http.StatusOK, newTestProjectUrl, gin.H{
			"Error": msg,
			"User": common.GetUserName(c),
		})
		return
	}
	isOk, msg = createNewProject(project, username, billing, "")
	if (!isOk) {
		c.HTML(http.StatusOK, newTestProjectUrl, gin.H{
			"Error": msg,
			"User": common.GetUserName(c),
		})
	} else {
		c.HTML(http.StatusOK, newTestProjectUrl, gin.H{
			"Success": msg,
			"User": common.GetUserName(c),
		})
	}
}

func updateBillingHandler(c *gin.Context) {
	username := common.GetUserName(c)
	project := c.PostForm("project")
	billing := c.PostForm("billing")

	isOk, msg := validateBillingInformation(project, billing, username)
	if (!isOk) {
		c.HTML(http.StatusOK, updateBillingUrl, gin.H{
			"Error": msg,
		})
		return
	}

	isOk, msg = createOrUpdateMetadata(project, billing, "", username)
	if (!isOk) {
		c.HTML(http.StatusOK, updateBillingUrl, gin.H{
			"Error": msg,
		})
	} else {
		c.HTML(http.StatusOK, updateBillingUrl, gin.H{
			"Success": "Die neuen Daten wurden gespeichert",
		})
	}
}

func validateNewProject(project string, billing string, isTestproject bool) (bool, string) {
	if (len(project) == 0) {
		return false, "Projektname muss angegeben werden"
	}

	if (!isTestproject && len(billing) == 0) {
		return false, "Kontierungsnummer muss angegeben werden"
	}

	return true, ""
}

func validateBillingInformation(project string, billing string, username string) (bool, string) {
	if (len(project) == 0) {
		return false, "Projektname muss angegeben werden"
	}

	if (len(billing) == 0) {
		return false, "Kontierungsnummer muss angegeben werden"
	}

	// Validate permissions
	isOk, msg := checkAdminPermissions(username, project)
	if (!isOk) {
		return false, msg
	}

	return true, ""
}

func createNewProject(project string, username string, billing string, megaid string) (bool, string) {
	p := models.NewObjectRequest{
		APIVersion: "v1",
		Kind: "ProjectRequest",
		Metadata: models.Metadata{Name: project, },
	}

	e, err := json.Marshal(p)
	if (err != nil) {
		log.Println("error encoding json:", err)
		return false, genericApiError
	}

	client, req := getOseHttpClient("POST",
		"oapi/v1/projectrequests",
		bytes.NewReader(e))

	resp, err := client.Do(req)
	defer resp.Body.Close()

	if (resp.StatusCode == http.StatusCreated) {
		log.Print(username + " created a new project: " + project)

	isOk, msg := changeProjectPermission(project, username)
	if (!isOk) {
		return isOk, msg
	}

	isOk, msg = createOrUpdateMetadata(project, billing, megaid, username)
	if (!isOk) {
		return isOk, msg
	} else {
		return true, "Das neue Projekt wurde erstellt"
	}
	} else {
		if (resp.StatusCode == http.StatusConflict) {
			return false, "Das Projekt existiert bereits"
		}

		errMsg, _ := ioutil.ReadAll(resp.Body)
		log.Println("Error creating new project:", err, resp.StatusCode, string(errMsg))

		return false, genericApiError
	}
}

func changeProjectPermission(project string, username string) (bool, string) {
	// Get existing policybindings
	policyBindings, msg := getPolicyBindings(project)

	if (policyBindings == nil) {
		return false, msg
	}

	children, err := policyBindings.S("roleBindings").Children()
	if (err != nil) {
		log.Println("Unable to parse roleBindings", err.Error())
		return false, genericApiError
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
		return false, genericApiError
	}

	if (resp.StatusCode == http.StatusOK) {
		log.Print(username + " is now admin of " + project)
		return true, ""
	} else {
		errMsg, _ := ioutil.ReadAll(resp.Body)
		log.Println("Error updating project permissions:", err, resp.StatusCode, string(errMsg))
		return false, genericApiError
	}
}

func createOrUpdateMetadata(project string, billing string, megaid string, username string) (bool, string) {
	client, req := getOseHttpClient("GET", "api/v1/namespaces/" + project, nil)
	resp, err := client.Do(req)
	defer resp.Body.Close()

	if (err != nil) {
		log.Println("Error from server: ", err.Error())
		return false, genericApiError
	}

	json, err := gabs.ParseJSONBuffer(resp.Body)
	if err != nil {
		log.Println("error decoding json:", err, resp.StatusCode)
		return false, genericApiError
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
		return true, ""
	} else {
		errMsg, _ := ioutil.ReadAll(resp.Body)
		log.Println("Error updating project config:", err, resp.StatusCode, string(errMsg))

		return false, genericApiError
	}
}
