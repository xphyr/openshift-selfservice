package gluster

import (
	"errors"
	"log"
	"net"
	"os/exec"
	"strings"
)

var MaxGB int
var MaxMB int
var Replicas int
var Port int
var PoolName string
var VgName string
var BasePath string
var Secret string

func getGlusterPeerServers() ([]string, error) {
	out, err := exec.Command("bash", "-c", "gluster peer status | grep Hostname").Output()
	if err != nil {
		log.Println("Error getting other gluster servers", err.Error())
		return []string{}, errors.New(commandExecutionError)
	}

	lines := strings.Split(string(out), "\n")
	servers := []string{}
	for _, l := range lines {
		if len(l) > 0 {
			servers = append(servers, strings.Replace(l, "Hostname: ", "", -1))
		}
	}

	return servers, nil
}

func getLocalServersIP() (string, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		log.Println("Error getting local servers ipv4 address", err.Error())
		return "", errors.New(commandExecutionError)
	}
	for _, i := range ifaces {
		addrs, err := i.Addrs()
		if err != nil {
			log.Println("Error getting local servers ipv4 address", err.Error())
			return "", errors.New(commandExecutionError)
		}

		for _, addr := range addrs {
			if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
				if ipnet.IP.To4() != nil {
					return ipnet.IP.String(), nil
				}
			}
		}
	}

	log.Println("IPv4 address of local server not found")
	return "", errors.New(commandExecutionError)
}
