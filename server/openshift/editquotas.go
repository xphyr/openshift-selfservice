package openshift

import (
	"encoding/json"
	"bytes"
	"net/http"
	"github.com/gin-gonic/gin"
	"os"
	"github.com/oscp/openshift-selfservice/server/common"
	"log"
	"github.com/oscp/openshift-selfservice/server/models"
	"io/ioutil"
)

func editQuotasHandler(c *gin.Context) {
	project := c.PostForm("project")
	cpu := c.PostForm("cpu")
	memory := c.PostForm("memory")
	username := common.GetUserName(c)

	isOk, msg := validateEditQuotas(username, project, cpu, memory)

	if (!isOk) {
		c.HTML(http.StatusOK, editQuotasUrl, gin.H{
			"Error": msg,
		})
		return
	}

	isOk, msg = updateQuotas(username, project, cpu, memory)

	c.HTML(http.StatusOK, editQuotasUrl, gin.H{
		"Success": "Die neuen Quotas wurden gespeichert",
	})
}

func validateEditQuotas(username string, project string, cpu string, memory string) (bool, string) {
	maxCPU := os.Getenv("MAX_CPU")
	maxMemory := os.Getenv("MAX_MEMORY")

	if (len(maxCPU) == 0 || len(maxMemory) == 0) {
		log.Fatal("Env variables 'MAX_MEMORY' and 'MAX_CPU' must be specified")
	}

	// Validate user input
	if (len(project) == 0) {
		return false, "Projekt muss angegeben werden"
	}
	isOk, msg := common.ValidateIntInput(maxCPU, cpu)
	if (!isOk) {
		return false, msg
	}
	isOk, msg = common.ValidateIntInput(maxMemory, memory)
	if (!isOk) {
		return false, msg
	}

	// Validate permissions
	isOk, msg = checkAdminPermissions(username, project)
	if (!isOk) {
		return false, msg
	}

	return true, ""
}

func updateQuotas(username string, project string, cpu string, memory string) (bool, string) {
	client, req := getOseHttpClient("GET", "api/v1/namespaces/" + project + "/resourcequotas", nil)
	resp, err := client.Do(req)
	defer resp.Body.Close()

	if (err != nil) {
		log.Println("Error from server: ", err.Error())
		return false, genericApiError
	}

	var existingQuotas models.ResourceQuotaResponse
	if err := json.NewDecoder(resp.Body).Decode(&existingQuotas); err != nil {
		log.Println("error decoding json:", err, resp.StatusCode)
		return false, genericApiError
	}

	existingQuotas.Items[0].Spec.Hard.CPU = cpu
	existingQuotas.Items[0].Spec.Hard.Memory = memory + "Gi"

	e, err := json.Marshal(existingQuotas.Items[0])
	if (err != nil) {
		log.Println("error encoding json:", err)
		return false, genericApiError
	}

	client, req = getOseHttpClient("PUT",
		"api/v1/namespaces/" + project + "/resourcequotas/" + existingQuotas.Items[0].Metadata.Name,
		bytes.NewReader(e))

	resp, err = client.Do(req)
	defer resp.Body.Close()

	if (resp.StatusCode == http.StatusOK) {
		log.Println("User " + username + " changed quotas for the project " + project + ". CPU: " + cpu, ", Mem: " + memory)
		return true, ""
	} else {
		errMsg, _ := ioutil.ReadAll(resp.Body)
		log.Println("Error updating resourceQuota:", err, resp.StatusCode, string(errMsg))

		return false, genericApiError
	}
}
