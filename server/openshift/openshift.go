package openshift

import (
	"crypto/tls"
	"net/http"
	"github.com/gin-gonic/gin"
	"os"
	"log"
	"encoding/json"
	"gopkg.in/appleboy/gin-jwt.v2"
	"github.com/oscp/openshift-selfservice/server/models"
)

func EditQuotasHandler(c *gin.Context) {
	var editQuotasCmd models.EditQuotasCommand
	c.BindJSON(&editQuotasCmd)

	reply := checkAdminPermissions(getUserName(c), editQuotasCmd.ProjectName)

	if (!reply.Status) {
		c.JSON(http.StatusInternalServerError, reply)
		return
	}

	log.Print("User has access to project, continue")
}

func getAddress(end string) string {
	base := os.Getenv("OPENSHIFT_API")

	if (len(base) == 0) {
		log.Fatal("Env variable 'OPENSHIFT_API' must be specified")
	}

	return base + "/" + end
}

func getHttpClient(method string, endUrl string) (*http.Client, *http.Request) {
	token := os.Getenv("OPENSHIFT_TOKEN")
	if (len(token) == 0) {
		log.Fatal("Env variable 'OPENSHIFT_TOKEN' must be specified")
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}

	req, _ := http.NewRequest(method, getAddress(endUrl), nil)
	log.Print("Calling ", req.URL.String())

	req.Header.Add("Authorization", "Bearer " + token)

	return client, req
}

func getUserName(c *gin.Context) string {
	jwtClaims := jwt.ExtractClaims(c)
	return jwtClaims["id"].(string)
}

func checkAdminPermissions(username string, projectName string) models.Reply {
	client, req := getHttpClient("GET", "oapi/v1/namespaces/" + projectName + "/policybindings/:default/")
	resp, err := client.Do(req)
	defer resp.Body.Close()

	genericError := "Fehler beim verarbeiten der OpenShift-API"

	if (err != nil) {
		log.Println("Error from server: ", err.Error())
		return models.Reply{
			Message: genericError,
			Status: false,
		}
	}

	if (resp.StatusCode == 404) {
		log.Println("Project was not found", projectName)
		return models.Reply{
			Message: "Das Projekt existiert nicht",
			Status: false,
		}
	}

	// Parse response
	var policyBindings models.PolicyBindingResponse
	err = json.NewDecoder(resp.Body).Decode(&policyBindings)
	if (err != nil) {
		log.Println("error decoding json:", err, resp.StatusCode)
		return models.Reply{
			Message: genericError,
			Status: false,
		}
	}

	log.Print(resp.StatusCode, policyBindings)

	return models.Reply{
		Status: true,
		Message: "",
	}
}
