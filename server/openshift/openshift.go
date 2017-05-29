package openshift

import (
	"crypto/tls"
	"net/http"
	"github.com/gin-gonic/gin"
	"os"
	"log"
	"gopkg.in/appleboy/gin-jwt.v2"
	"strconv"
	"github.com/oscp/openshift-selfservice/server/models"
	"encoding/json"
	"io/ioutil"
	"strings"
)

const GENERIC_API_ERROR = "Fehler beim Aufruf der OpenShift-API"

func RegisterRoutes(r *gin.Engine) {
	r.GET("/openshift/editquotas", func(c *gin.Context) {
		c.HTML(http.StatusOK, "editquotas.html", gin.H{})
	})
	r.POST("/openshift/editquotas", editQuotasHandler)
}

func editQuotasHandler(c *gin.Context) {
	project := c.PostForm("project")
	cpu := c.PostForm("cpu")
	memory := c.PostForm("memory")

	isOk, msg := validateEditQuotas("u220374", project, cpu, memory)

	if (!isOk) {
		c.HTML(http.StatusOK, "editquotas.html", gin.H{
			"Error": msg,
		})
		return
	}

	log.Print("User has access to project, continue")

	c.HTML(http.StatusOK, "editquotas.html", gin.H{
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
	isOk, msg := validateIntInput(maxCPU, cpu)
	if (!isOk) {
		return false, msg
	}
	isOk, msg = validateIntInput(maxMemory, memory)
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

func validateIntInput(maxValue string, input string) (bool, string) {
	maxInt, err := strconv.Atoi(maxValue)
	if (err != nil) {
		log.Fatal("Could not parse 'MAX' value of", maxValue)
	}

	inputInt, err := strconv.Atoi(input)
	if (err != nil) {
		return false, "Bitte eine gültige Zahl eintragen"
	}

	if (inputInt > maxInt) {
		return false, "Du kannst maximal " + maxValue + " eintragen"
	}

	return true, ""
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

func checkAdminPermissions(username string, projectName string) (bool, string) {
	client, req := getHttpClient("GET", "oapi/v1/namespaces/" + projectName + "/policybindings/:default")
	resp, err := client.Do(req)
	defer resp.Body.Close()

	if (err != nil) {
		log.Println("Error from server: ", err.Error())
		return false, GENERIC_API_ERROR
	}

	if (resp.StatusCode == 404) {
		log.Println("Project was not found", projectName)
		return false, "Das Projekt existiert nicht"
	}

	// Remove the groupNames: null because of bug in k8s api
	policyBindings := models.PolicyBindingResponse{}
	bodyBytes, err :=ioutil.ReadAll(resp.Body)
	if (err != nil) {
		log.Println("error parsing body of response:", err)
		return false, GENERIC_API_ERROR
	}

	cBody := strings.Replace(strings.Replace(string(bodyBytes), "\"groupNames\":null,", "", -1),
		"\"userNames\":null,", "", -1)

	if err := json.Unmarshal([]byte(cBody), &policyBindings); err != nil {
		log.Println("error decoding json:", err, resp.StatusCode)
		return false, GENERIC_API_ERROR
	}

	// Check if user has admin-access
	hasAccess := false
	admins := ""
	for _,v := range policyBindings.RoleBindings {
		if (v.Name == "admin") {
			for _,u := range v.RoleBinding.UserNames {
				if (u == username) {
					hasAccess = true
				}

				if (len(admins) != 0) {
					admins += ", "
				}
				admins += u
			}
		}
	}

	if (hasAccess) {
		return true, ""
	} else {
		return false, "Du hast keine Admin Rechte auf dem Projekt. Bestehende Admins sind folgende Benutzer: " + admins
	}
}
