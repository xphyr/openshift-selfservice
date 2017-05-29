package openshift

import (
	"crypto/tls"
	"net/http"
	"github.com/gin-gonic/gin"
	"os"
	"log"
	"github.com/oscp/openshift-selfservice/server/models"
	"encoding/json"
	"io/ioutil"
	"strings"
	"io"
	"github.com/oscp/openshift-selfservice/server/common"
)

const GENERIC_API_ERROR = "Fehler beim Aufruf der OpenShift-API"

func RegisterRoutes(r *gin.RouterGroup) {
	r.GET("/openshift/editquotas", func(c *gin.Context) {
		c.HTML(http.StatusOK, "editquotas.html", gin.H{})
	})
	r.POST("/openshift/editquotas", editQuotasHandler)
}

func checkAdminPermissions(username string, project string) (bool, string) {
	client, req := getOseHttpClient("GET", "oapi/v1/namespaces/" + project + "/policybindings/:default", nil)
	resp, err := client.Do(req)
	defer resp.Body.Close()

	if (err != nil) {
		log.Println("Error from server: ", err.Error())
		return false, GENERIC_API_ERROR
	}

	if (resp.StatusCode == 404) {
		log.Println("Project was not found", project)
		return false, "Das Projekt existiert nicht"
	}

	// Remove the null values because of bug in OSE api
	policyBindings := models.PolicyBindingResponse{}
	bodyBytes, err := ioutil.ReadAll(resp.Body)
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
	for _, v := range policyBindings.RoleBindings {
		if (v.Name == "admin") {
			for _, u := range v.RoleBinding.UserNames {
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

func getOseAddress(end string) string {
	base := os.Getenv("OPENSHIFT_API")

	if (len(base) == 0) {
		log.Fatal("Env variable 'OPENSHIFT_API' must be specified")
	}

	return base + "/" + end
}

func getOseHttpClient(method string, endUrl string, body io.Reader) (*http.Client, *http.Request) {
	token := os.Getenv("OPENSHIFT_TOKEN")
	if (len(token) == 0) {
		log.Fatal("Env variable 'OPENSHIFT_TOKEN' must be specified")
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}

	req, _ := http.NewRequest(method, getOseAddress(endUrl), body)

	if (common.DebugMode()) {
		log.Print("Calling ", req.URL.String())
	}

	req.Header.Add("Authorization", "Bearer " + token)

	return client, req
}


