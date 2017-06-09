package gluster

import (
	"os/exec"
	"log"
	"strings"
)

var MaxGB int
var MaxMB int
var Port int
var PoolName string
var VgName string
var BasePath string
var Secret string

func getGlusterPeerServers() ([]string, error) {
	out, err := exec.Command("bash", "-c", "gluster peer status | grep Hostname").Output()
	if (err != nil) {
		log.Println("Error getting other gluster servers", err.Error())
		return []string{}, err
	}

	lines := strings.Split(string(out), "\n")
	servers := []string{}
	for _, l := range lines {
		if (len(l) > 0) {
			servers = append(servers, strings.Replace(l, "Hostname: ", "", -1))
		}
	}

	return servers, nil
}