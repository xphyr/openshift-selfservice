package gluster

import (
	"os/exec"
	"log"
	"strings"
)

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

func executeCommandsLocally(commands []string) (bool, string) {
	for _, c := range commands {
		out, err := exec.Command("bash", "-c", c).Output()
		if (err != nil) {
			log.Println("Error executing command: ", c, err.Error(), out)
			return false, commandExecutionError
		}
	}

	return true, ""
}

