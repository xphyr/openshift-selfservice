package openshift

import (
	"bytes"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/Jeffail/gabs"
	"github.com/gin-gonic/gin"
	"github.com/oscp/cloud-selfservice-portal/server/common"
)

func editQuotasHandler(c *gin.Context) {
	project := c.PostForm("project")
	cpu := c.PostForm("cpu")
	memory := c.PostForm("memory")
	username := common.GetUserName(c)

	if err := validateEditQuotas(username, project, cpu, memory); err != nil {
		c.HTML(http.StatusOK, editQuotasURL, gin.H{
			"Error": err.Error(),
		})
		return
	}

	if err := updateQuotas(username, project, cpu, memory); err != nil {
		c.HTML(http.StatusOK, editQuotasURL, gin.H{
			"Error": err.Error(),
		})
		return
	}

	c.HTML(http.StatusOK, editQuotasURL, gin.H{
		"Success": "Die neuen Quotas wurden gespeichert",
	})
}

func validateEditQuotas(username string, project string, cpu string, memory string) error {
	maxCPU := os.Getenv("MAX_CPU")
	maxMemory := os.Getenv("MAX_MEMORY")

	if len(maxCPU) == 0 || len(maxMemory) == 0 {
		log.Fatal("Env variables 'MAX_MEMORY' and 'MAX_CPU' must be specified")
	}

	// Validate user input
	if len(project) == 0 {
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

func updateQuotas(username string, project string, cpu string, memory string) error {
	client, req := getOseHTTPClient("GET", "api/v1/namespaces/"+project+"/resourcequotas", nil)
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

	firstQuota := json.S("items").Index(0)

	firstQuota.SetP(cpu, "spec.hard.cpu")
	firstQuota.SetP(memory+"Gi", "spec.hard.memory")

	client, req = getOseHTTPClient("PUT",
		"api/v1/namespaces/"+project+"/resourcequotas/"+firstQuota.Path("metadata.name").Data().(string),
		bytes.NewReader(firstQuota.Bytes()))

	resp, err = client.Do(req)
	if err == nil && resp.StatusCode == http.StatusOK {
		defer resp.Body.Close()
		log.Println("User "+username+" changed quotas for the project "+project+". CPU: "+cpu, ", Mem: "+memory)
		return nil
	}

	errMsg, _ := ioutil.ReadAll(resp.Body)
	log.Println("Error updating resourceQuota:", err.Error(), resp.StatusCode, string(errMsg))

	return errors.New(genericAPIError)
}
