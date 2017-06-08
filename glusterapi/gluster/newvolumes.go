package gluster

import (
	"strings"
	"strconv"
	"log"
	"fmt"
	"net/http"
	"bytes"
	"github.com/oscp/openshift-selfservice/glusterapi/models"
	"encoding/json"
	"io/ioutil"
)

const (
	suffixWrongError = "Invalid size. Size must be int followed by suffix (e.g. 100M). Allowed suffixes are 'G/M'. You sent: "
	commandExecutionError = "Error running command, see logs for details"
)

func CreateVolume(project string, size string) (bool, string) {
	isOk, msg := validateSizeInput(size)
	if (!isOk) {
		log.Print("Aborting...", msg)
		return isOk, msg
	}

	// Create lvs on pool on all gluster servers
	ok, msg := createLvOnAllServers(project, size)
	if (!ok) {
		return ok, msg
	}

	// Create gluster volume

	return true, ""
}

func validateSizeInput(size string) (bool, string) {
	log.Println("Checking size of", size)

	if (strings.HasSuffix(size, "M")) {
		sizeInt, err := strconv.Atoi(strings.Replace(size, "M", "", 1))
		if (err != nil) {
			return false, suffixWrongError + size
		}

		if (sizeInt <= MaxMB) {
			return true, ""
		} else {
			return false, "Your size is to big for suffix 'M' use 'G' instead"
		}
	}
	if (strings.HasSuffix(size, "G")) {
		sizeInt, err := strconv.Atoi(strings.Replace(size, "G", "", 1))
		if (err != nil) {
			return false, suffixWrongError + size
		}

		if (sizeInt > MaxGB) {
			return false, fmt.Sprintf("Max allowed size exceeded. Max allowed is: %vG", MaxGB)
		}

		return true, ""
	}

	return false, suffixWrongError + size
}

func createLvOnAllServers(project string, size string) (bool, string) {
	pvNumber, err := getExistingLvForProject(project)
	if (err != nil) {
		return false, commandExecutionError
	}

	mountPoint := fmt.Sprintf("%v/%v/%v", BasePath, project, pvNumber)
	lvName := fmt.Sprintf("lv_%v_%v", project, pvNumber)

	// Create the lv on all other gluster servers
	ok, msg := createLvOnOtherServers(size, mountPoint, lvName)
	if (!ok) {
		return ok, msg
	}

	// Create the lv locally
	ok, msg = CreateLvOnPool(size, mountPoint, lvName)
	if (!ok) {
		return ok, msg
	}

	return true, ""
}

func createLvOnOtherServers(size string, mountPoint string, lvName string) (bool, string) {
	remotes, err := getGlusterPeerServers()
	if (err != nil) {
		return false, commandExecutionError
	}

	p := models.CreateLVCommand{
		LvName: lvName,
		MountPoint: mountPoint,
		Size: size,
	}
	b := new(bytes.Buffer)
	err = json.NewEncoder(b).Encode(p)
	if (err != nil) {
		log.Println("Error encoding json", err.Error())
		return false, commandExecutionError
	}

	// Execute the commands remote via API
	client := &http.Client{}
	for _, r := range remotes {
		log.Println("Going to create lv on remote:", r)

		req, _ := http.NewRequest("POST", fmt.Sprintf("http://%v:%v/sec/lv", r, Port), b)
		req.SetBasicAuth("GLUSTER_API", Secret)

		resp, err := client.Do(req)
		defer resp.Body.Close()
		if (err != nil || resp.StatusCode != http.StatusOK) {
			body, _ := ioutil.ReadAll(resp.Body)
			log.Println("Remote did not respond with OK", resp.StatusCode, body)
			return false, commandExecutionError
		}
	}

	return true, ""
}

func CreateLvOnPool(size string, mountPoint string, lvName string) (bool, string) {
	commands := []string{
		// Create a directory
		fmt.Sprintf("mkdir -p %v", mountPoint),

		// Create a lv
		fmt.Sprintf("lvcreate -V %v -T %v/%v -n %v", size, VgName, PoolName, lvName),

		// Create file system
		fmt.Sprintf("mkfs.xfs -i size=512 -n size=8192 /dev/%v/%v", VgName, lvName),

		// Fstab
		fmt.Sprintf("echo \"/dev/%v/%v %v xfs rw,inode64,noatime,nouuid 1 2\" | tee -a /etc/fstab > /dev/null ",
			VgName,
			lvName,
			mountPoint),

		// Mount
		fmt.Sprintf("mount -o rw,inode64,noatime,nouuid /dev/%v/%v %v", VgName, lvName, mountPoint),

		// Handle Selinux
		fmt.Sprintf("semanage fcontext -a -t glusterd_brick_t %v", mountPoint),
		fmt.Sprintf("restorecon -Rv %v", mountPoint),
	}

	log.Println("Going to execute the following commands: ")
	for _, c := range commands {
		log.Println(c)
	}

	//ok, msg := executeCommandsLocally(commands)
	//if (!ok) {
	//	return ok, msg
	//}

	return true, ""
}
