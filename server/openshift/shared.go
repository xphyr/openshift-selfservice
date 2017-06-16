package openshift

import (
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/Jeffail/gabs"
	"github.com/gin-gonic/gin"
	"github.com/oscp/cloud-selfservice-portal/server/common"
)

const (
	editQuotasURL        = "editquotas.html"
	newProjectURL        = "newproject.html"
	newTestProjectURL    = "newtestproject.html"
	updateBillingURL     = "updatebilling.html"
	newServiceAccountURL = "newserviceaccount.html"
	newVolumeURL         = "newvolume.html"
	fixVolumeURL         = "fixvolume.html"
	growVolumeURL        = "growvolume.html"
	genericAPIError      = "Fehler beim Aufruf der OpenShift-API"
)

// RegisterRoutes registers the routes for OpenShift
func RegisterRoutes(r *gin.RouterGroup) {
	// Quotas
	r.GET("/openshift/editquotas", func(c *gin.Context) {
		c.HTML(http.StatusOK, editQuotasURL, gin.H{})
	})
	r.POST("/openshift/editquotas", editQuotasHandler)

	// NewProject
	r.GET("/openshift/newproject", func(c *gin.Context) {
		c.HTML(http.StatusOK, newProjectURL, gin.H{})
	})
	r.POST("/openshift/newproject", newProjectHandler)

	// NewTestProject
	r.GET("/openshift/newtestproject", func(c *gin.Context) {
		c.HTML(http.StatusOK, newTestProjectURL, gin.H{
			"User": common.GetUserName(c),
		})
	})
	r.POST("/openshift/newtestproject", newTestProjectHandler)

	// Update billing
	r.GET("/openshift/updatebilling", func(c *gin.Context) {
		c.HTML(http.StatusOK, updateBillingURL, gin.H{})
	})
	r.POST("/openshift/updatebilling", updateBillingHandler)

	// NewServiceAccount
	r.GET("/openshift/newserviceaccount", func(c *gin.Context) {
		c.HTML(http.StatusOK, newServiceAccountURL, gin.H{})
	})
	r.POST("/openshift/newserviceaccount", newServiceAccountHandler)

	// NewVolume
	r.GET("/openshift/newvolume", func(c *gin.Context) {
		c.HTML(http.StatusOK, newVolumeURL, gin.H{})
	})
	r.POST("/openshift/newvolume", newVolumeHandler)

	// FixVolume
	r.GET("/openshift/fixvolume", func(c *gin.Context) {
		c.HTML(http.StatusOK, fixVolumeURL, gin.H{})
	})
	r.POST("/openshift/fixvolume", fixVolumeHandler)

	// growVolume
	r.GET("/openshift/growvolume", func(c *gin.Context) {
		c.HTML(http.StatusOK, growVolumeURL, gin.H{})
	})
	r.POST("/openshift/growvolume", growVolumeHandler)
}

func checkAdminPermissions(username string, project string) error {
	policyBindings, err := getPolicyBindings(project)
	if err != nil {
		return err
	}

	// Check if user has admin-access
	hasAccess := false
	admins := ""
	children, err := policyBindings.S("roleBindings").Children()
	if err != nil {
		log.Println("Unable to parse roleBindings", err.Error())
		return errors.New(genericAPIError)
	}
	for _, v := range children {
		if v.Path("name").Data().(string) == "admin" {
			usernames, err := v.Path("roleBinding.userNames").Children()
			if err != nil {
				log.Println("Unable to parse roleBinding", err.Error())
				return errors.New(genericAPIError)
			}
			for _, u := range usernames {
				if strings.ToLower(u.Data().(string)) == strings.ToLower(username) {
					hasAccess = true
				}

				if len(admins) != 0 {
					admins += ", "
				}
				admins += u.Data().(string)
			}
		}
	}

	if hasAccess {
		return nil
	}

	return fmt.Errorf("Du hast keine Admin Rechte auf dem Projekt. Bestehende Admins sind folgende Benutzer: %v", admins)
}

func getPolicyBindings(project string) (*gabs.Container, error) {
	client, req := getOseHTTPClient("GET", "oapi/v1/namespaces/"+project+"/policybindings/:default", nil)
	resp, err := client.Do(req)

	if err != nil {
		log.Println("Error from server: ", err.Error())
		return nil, errors.New(genericAPIError)
	}

	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		log.Println("Project was not found", project)
		return nil, errors.New("Das Projekt existiert nicht")
	}

	json, err := gabs.ParseJSONBuffer(resp.Body)
	if err != nil {
		log.Println("error parsing body of response:", err)
		return nil, errors.New(genericAPIError)
	}

	return json, nil
}

func getOseAddress(end string) string {
	base := os.Getenv("OPENSHIFT_API")

	if len(base) == 0 {
		log.Fatal("Env variable 'OPENSHIFT_API' must be specified")
	}

	return base + "/" + end
}

func getOseHTTPClient(method string, endURL string, body io.Reader) (*http.Client, *http.Request) {
	token := os.Getenv("OPENSHIFT_TOKEN")
	if len(token) == 0 {
		log.Fatal("Env variable 'OPENSHIFT_TOKEN' must be specified")
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}

	req, _ := http.NewRequest(method, getOseAddress(endURL), body)

	if common.DebugMode() {
		log.Print("Calling ", req.URL.String())
	}

	req.Header.Add("Authorization", "Bearer "+token)

	return client, req
}

func getGlusterHTTPClient(url string, body io.Reader) (*http.Client, *http.Request) {
	glusterURL := os.Getenv("GLUSTER_API_URL")
	glusterSecret := os.Getenv("GLUSTER_SECRET")

	if len(glusterURL) == 0 || len(glusterSecret) == 0 {
		log.Fatal("Env variables 'GLUSTER_API_URL' and 'GLUSTER_SECRET' must be specified")
	}

	client := &http.Client{}
	req, _ := http.NewRequest("POST", fmt.Sprintf("%v/%v", glusterURL, url), body)

	if common.DebugMode() {
		log.Printf("Calling %v", req.URL.String())
	}

	req.SetBasicAuth("GLUSTER_API", glusterSecret)

	return client, req
}

func newObjectRequest(kind string, name string) *gabs.Container {
	json := gabs.New()

	json.Set(kind, "kind")
	json.Set("v1", "apiVersion")
	json.SetP(name, "metadata.name")

	return json
}
