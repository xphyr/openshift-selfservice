package openshift

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"github.com/oscp/openshift-selfservice/server/common"
	"bytes"
	"log"
	"io/ioutil"
)

func newServiceAccountHandler(c *gin.Context) {
	project := c.PostForm("project")
	serviceaccount := c.PostForm("serviceaccount")
	username := common.GetUserName(c)

	isOk, msg := validateNewServiceAccount(username, project, serviceaccount)
	if (!isOk) {
		c.HTML(http.StatusOK, newServiceAccountUrl, gin.H{
			"Error": msg,
		})
		return
	}

	isOk, msg = createNewServiceAccount(username, project, serviceaccount)

	c.HTML(http.StatusOK, newServiceAccountUrl, gin.H{
		"Success": msg,
	})
}

func validateNewServiceAccount(username string, project string, serviceAccountName string) (bool, string) {
	if (len(serviceAccountName) == 0) {
		return false, "Service Account muss angegeben werden"
	}

	// Validate permissions
	isOk, msg := checkAdminPermissions(username, project)
	if (!isOk) {
		return false, msg
	}

	return true, ""
}

func createNewServiceAccount(username string, project string, serviceaccount string) (bool, string) {
	p := newObjectRequest("ServiceAccount", serviceaccount)

	client, req := getOseHttpClient("POST",
		"api/v1/namespaces/" + project + "/serviceaccounts",
		bytes.NewReader(p.Bytes()))

	resp, err := client.Do(req)
	defer resp.Body.Close()

	if (resp.StatusCode == http.StatusCreated) {
		log.Print(username + " created a new service account: " + serviceaccount + " on project " + project)

		return true, "Service Account wurde angelegt"
	} else {
		errMsg, _ := ioutil.ReadAll(resp.Body)
		log.Println("Error creating new project:", err, resp.StatusCode, string(errMsg))

		return false, genericApiError
	}
}
