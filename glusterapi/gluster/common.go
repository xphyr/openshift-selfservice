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

func getExistingLvForProject(project string) (int, error) {
	out, err := exec.Command("bash", "-c", "lvs").Output()
	if err != nil {
		log.Println("Could not count existing lvs for a project:", project, err.Error())
		return -1, err
	}

	return strings.Count(string(out), "lv" + project) + 1, nil
}