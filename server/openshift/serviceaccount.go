package openshift

import (
	"bytes"
	"errors"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/oscp/cloud-selfservice-portal/server/common"
)

func newServiceAccountHandler(c *gin.Context) {
	project := c.PostForm("project")
	serviceaccount := c.PostForm("serviceaccount")
	username := common.GetUserName(c)

	if err := validateNewServiceAccount(username, project, serviceaccount); err != nil {
		c.HTML(http.StatusOK, newServiceAccountURL, gin.H{
			"Error": err.Error(),
		})
		return
	}

	if err := createNewServiceAccount(username, project, serviceaccount); err != nil {
		c.HTML(http.StatusOK, newServiceAccountURL, gin.H{
			"Error": err.Error(),
		})
	} else {

		c.HTML(http.StatusOK, newServiceAccountURL, gin.H{
			"Success": "Der Service Account wurde angelegt",
		})
	}
}

func validateNewServiceAccount(username string, project string, serviceAccountName string) error {
	if len(serviceAccountName) == 0 {
		return errors.New("Service Account muss angegeben werden")
	}

	// Validate permissions
	if err := checkAdminPermissions(username, project); err != nil {
		return err
	}

	return nil
}

func createNewServiceAccount(username string, project string, serviceaccount string) error {
	p := newObjectRequest("ServiceAccount", serviceaccount)

	client, req := getOseHTTPClient("POST",
		"api/v1/namespaces/"+project+"/serviceaccounts",
		bytes.NewReader(p.Bytes()))

	resp, err := client.Do(req)

	if resp.StatusCode == http.StatusCreated {
		resp.Body.Close()
		log.Print(username + " created a new service account: " + serviceaccount + " on project " + project)
		return nil
	}

	errMsg, _ := ioutil.ReadAll(resp.Body)
	log.Println("Error creating new project:", err, resp.StatusCode, string(errMsg))
	return errors.New(genericAPIError)
}
