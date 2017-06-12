package openshift

import (
	"net/http"
	"github.com/gin-gonic/gin"
	"os"
	"github.com/oscp/openshift-selfservice/server/common"
	"log"
	"github.com/Jeffail/gabs"
	"bytes"
	"io/ioutil"
	"errors"
)

func editQuotasHandler(c *gin.Context) {
	project := c.PostForm("project")
	cpu := c.PostForm("cpu")
	memory := c.PostForm("memory")
	username := common.GetUserName(c)

	if err := validateEditQuotas(username, project, cpu, memory); err != nil {
		c.HTML(http.StatusOK, editQuotasUrl, gin.H{
			"Error": err.Error(),
		})
		return
	}

	if err := updateQuotas(username, project, cpu, memory); err != nil {
		c.HTML(http.StatusOK, editQuotasUrl, gin.H{
			"Error": err.Error(),
		})
		return
	}

	c.HTML(http.StatusOK, editQuotasUrl, gin.H{
		"Success": "Die neuen Quotas wurden gespeichert",
	})
}

func validateEditQuotas(username string, project string, cpu string, memory string) (error) {
	maxCPU := os.Getenv("MAX_CPU")
	maxMemory := os.Getenv("MAX_MEMORY")

	if (len(maxCPU) == 0 || len(maxMemory) == 0) {
		log.Fatal("Env variables 'MAX_MEMORY' and 'MAX_CPU' must be specified")
	}

	// Validate user input
	if (len(project) == 0) {
		return errors.New("Projekt muss angegeben werden")
	}
	if err := common.ValidateIntInput(maxCPU, cpu); err != nil {
		return err
	}
	if err := common.ValidateIntInput(maxMemory, memory); err != nil {
		return err
	}

	// Validate permissions
	if err := checkAdminPermissions(username, project); err != nil {
		return err
	}

	return nil
}

func updateQuotas(username string, project string, cpu string, memory string) (error) {
	client, req := getOseHttpClient("GET", "api/v1/namespaces/" + project + "/resourcequotas", nil)
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

	firstQuota := json.S("items").Index(0)

	firstQuota.SetP(cpu, "spec.hard.cpu")
	firstQuota.SetP(memory + "Gi", "spec.hard.memory")

	client, req = getOseHttpClient("PUT",
		"api/v1/namespaces/" + project + "/resourcequotas/" + firstQuota.Path("metadata.name").Data().(string),
		bytes.NewReader(firstQuota.Bytes()))

	resp, err = client.Do(req)
	defer resp.Body.Close()

	if (resp.StatusCode == http.StatusOK) {
		log.Println("User " + username + " changed quotas for the project " + project + ". CPU: " + cpu, ", Mem: " + memory)
		return nil
	} else {
		errMsg, _ := ioutil.ReadAll(resp.Body)
		log.Println("Error updating resourceQuota:", err, resp.StatusCode, string(errMsg))

		return errors.New(genericApiError)
	}
}
