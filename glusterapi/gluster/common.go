package gluster

import (
	"fmt"
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
	cmd := fmt.Sprintf("lvs | grep lv_%v")

	out, err := exec.Command("bash", "-c", cmd).Output()
	if err != nil {
		log.Println("Could not count existing lvs for a project: ", project, err.Error())
		return -1, err
	}

	return len(strings.Split(string(out), "\n")), nil
}