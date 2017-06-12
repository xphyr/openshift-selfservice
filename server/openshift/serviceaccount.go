package openshift

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"github.com/oscp/openshift-selfservice/server/common"
	"bytes"
	"log"
	"io/ioutil"
	"errors"
)

func newServiceAccountHandler(c *gin.Context) {
	project := c.PostForm("project")
	serviceaccount := c.PostForm("serviceaccount")
	username := common.GetUserName(c)

	if err := validateNewServiceAccount(username, project, serviceaccount); err != nil {
		c.HTML(http.StatusOK, newServiceAccountUrl, gin.H{
			"Error": err.Error(),
		})
		return
	}

	if err := createNewServiceAccount(username, project, serviceaccount); err != nil {
		c.HTML(http.StatusOK, newServiceAccountUrl, gin.H{
			"Error": err.Error(),
		})
	} else {

		c.HTML(http.StatusOK, newServiceAccountUrl, gin.H{
			"Success": "Der Service Account wurde angelegt",
		})
	}
}

func validateNewServiceAccount(username string, project string, serviceAccountName string) (error) {
	if (len(serviceAccountName) == 0) {
		return errors.New("Service Account muss angegeben werden")
	}

	// Validate permissions
	if err := checkAdminPermissions(username, project); err != nil {
		return err
	}

	return nil
}

func createNewServiceAccount(username string, project string, serviceaccount string) (error) {
	p := newObjectRequest("ServiceAccount", serviceaccount)

	client, req := getOseHttpClient("POST",
		"api/v1/namespaces/" + project + "/serviceaccounts",
		bytes.NewReader(p.Bytes()))

	resp, err := client.Do(req)
	defer resp.Body.Close()

	if (resp.StatusCode == http.StatusCreated) {
		log.Print(username + " created a new service account: " + serviceaccount + " on project " + project)
		return nil
	} else {
		errMsg, _ := ioutil.ReadAll(resp.Body)
		log.Println("Error creating new project:", err, resp.StatusCode, string(errMsg))
		return errors.New(genericApiError)
	}
}
