package openshift

import (
	"crypto/tls"
	"net/http"
	"github.com/gin-gonic/gin"
	"os"
	"log"
	"strings"
	"io"
	"github.com/oscp/openshift-selfservice/server/common"
	"github.com/Jeffail/gabs"
	"errors"
	"fmt"
)

const (
	editQuotasUrl = "editquotas.html"
	newProjectUrl = "newproject.html"
	newTestProjectUrl = "newtestproject.html"
	updateBillingUrl = "updatebilling.html"
	newServiceAccountUrl = "newserviceaccount.html"
	genericApiError = "Fehler beim Aufruf der OpenShift-API"
)

func RegisterRoutes(r *gin.RouterGroup) {
	// Quotas
	r.GET("/openshift/editquotas", func(c *gin.Context) {
		c.HTML(http.StatusOK, editQuotasUrl, gin.H{})
	})
	r.POST("/openshift/editquotas", editQuotasHandler)

	// NewProject
	r.GET("/openshift/newproject", func(c *gin.Context) {
		c.HTML(http.StatusOK, newProjectUrl, gin.H{})
	})
	r.POST("/openshift/newproject", newProjectHandler)

	// NewTestProject
	r.GET("/openshift/newtestproject", func(c *gin.Context) {
		c.HTML(http.StatusOK, newTestProjectUrl, gin.H{
			"User": common.GetUserName(c),
		})
	})
	r.POST("/openshift/newtestproject", newTestProjectHandler)

	// Update billing
	r.GET("/openshift/updatebilling", func(c *gin.Context) {
		c.HTML(http.StatusOK, updateBillingUrl, gin.H{})
	})
	r.POST("/openshift/updatebilling", updateBillingHandler)

	// NewServiceAccount
	r.GET("/openshift/newserviceaccount", func(c *gin.Context) {
		c.HTML(http.StatusOK, newServiceAccountUrl, gin.H{})
	})
	r.POST("/openshift/newserviceaccount", newServiceAccountHandler)
}

func checkAdminPermissions(username string, project string) (error) {
	policyBindings, err := getPolicyBindings(project)
	if (err != nil) {
		return err
	}

	// Check if user has admin-access
	hasAccess := false
	admins := ""
	children, err := policyBindings.S("roleBindings").Children()
	if (err != nil) {
		log.Println("Unable to parse roleBindings", err.Error())
		return errors.New(genericApiError)
	}
	for _, v := range children {
		if (v.Path("name").Data().(string) == "admin") {
			usernames, err := v.Path("roleBinding.userNames").Children()
			if (err != nil) {
				log.Println("Unable to parse roleBinding", err.Error())
				return errors.New(genericApiError)
			}
			for _, u := range usernames {
				if (strings.ToLower(u.Data().(string)) == strings.ToLower(username)) {
					hasAccess = true
				}

				if (len(admins) != 0) {
					admins += ", "
				}
				admins += u.Data().(string)
			}
		}
	}

	if (hasAccess) {
		return nil
	} else {
		return fmt.Errorf("Du hast keine Admin Rechte auf dem Projekt. Bestehende Admins sind folgende Benutzer: %v", admins)
	}
}

func getPolicyBindings(project string) (*gabs.Container, error) {
	client, req := getOseHttpClient("GET", "oapi/v1/namespaces/" + project + "/policybindings/:default", nil)
	resp, err := client.Do(req)
	defer resp.Body.Close()

	if (err != nil) {
		log.Println("Error from server: ", err.Error())
		return nil, errors.New(genericApiError)
	}

	if (resp.StatusCode == 404) {
		log.Println("Project was not found", project)
		return nil, errors.New("Das Projekt existiert nicht")
	}

	json, err := gabs.ParseJSONBuffer(resp.Body)
	if (err != nil) {
		log.Println("error parsing body of response:", err)
		return nil, errors.New(genericApiError)
	}

	return json, nil
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

func newObjectRequest(kind string, name string) *gabs.Container {
	json := gabs.New()

	json.Set(kind, "kind")
	json.Set("v1", "apiVersion")
	json.SetP(name, "metadata.name")

	return json
}