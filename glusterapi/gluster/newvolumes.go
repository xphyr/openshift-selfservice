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
	"os/exec"
	"errors"
)

const (
	suffixWrongError = "Invalid size. Size must be int followed by suffix (e.g. 100M). Allowed suffixes are 'G/M'. You sent: %v"
	commandExecutionError = "Error running command, see logs for details"
)

func CreateVolume(project string, size string) (error) {
	if err := validateSizeInput(size); err != nil {
		return err
	}

	pvNumber, err := getExistingLvForProject(project)
	if (err != nil) {
		return err
	}

	mountPoint := fmt.Sprintf("%v/%v/pv%v", BasePath, project, pvNumber)
	lvName := fmt.Sprintf("lv_%v_pv%v", project, pvNumber)

	// Create lvs on pool on all gluster servers
	if err := createLvOnAllServers(size, mountPoint, lvName); err != nil {
		return err
	}

	// Create gluster volume
	if err := createGlusterVolume(project, size); err != nil {
		return err
	}

	return true, ""
}

func validateSizeInput(size string) (error) {
	log.Println("Checking size of", size)

	if (strings.HasSuffix(size, "M")) {
		sizeInt, err := strconv.Atoi(strings.Replace(size, "M", "", 1))
		if (err != nil) {
			return fmt.Errorf(suffixWrongError, size)
		}

		if (sizeInt <= MaxMB) {
			return nil
		} else {
			return errors.New("Your size is to big for suffix 'M' use 'G' instead")
		}
	}
	if (strings.HasSuffix(size, "G")) {
		sizeInt, err := strconv.Atoi(strings.Replace(size, "G", "", 1))
		if (err != nil) {
			return fmt.Errorf(suffixWrongError, size)
		}

		if (sizeInt > MaxGB) {
			return fmt.Errorf("Max allowed size exceeded. Max allowed is: %vG", MaxGB)
		}

		return nil
	}

	return fmt.Errorf(suffixWrongError, size)
}

func CreateLvOnPool(size string, mountPoint string, lvName string) (error) {
	if (len(size) == 0 || len(mountPoint) == 0 || len(lvName) == 0) {
		return false, "Not all input values provided"
	}

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

func getExistingLvForProject(project string) (int, error) {
	out, err := exec.Command("bash", "-c", "lvs").Output()
	if err != nil {
		return -1, fmt.Errorf("Could not count existing lvs for a project: %v. Error: %v", project, err.Error())
	}

	return strings.Count(string(out), "lv" + project) + 1, nil
}

func createLvOnAllServers(size string, mountPoint string, lvName string) (error) {
	// Create the lv on all other gluster servers
	if err := createLvOnOtherServers(size, mountPoint, lvName); err != nil {
		return err
	}

	// Create the lv locally
	if err := CreateLvOnPool(size, mountPoint, lvName); err != nil {
		return err
	}

	return nil
}

func createLvOnOtherServers(size string, mountPoint string, lvName string) (error) {
	remotes, err := getGlusterPeerServers()
	if (err != nil) {
		return errors.New(commandExecutionError)
	}

	p := models.CreateLVCommand{
		LvName: lvName,
		MountPoint: mountPoint,
		Size: size,
	}
	b := new(bytes.Buffer)

	if err = json.NewEncoder(b).Encode(p); err != nil) {
		log.Println("Error encoding json", err.Error())
		return errors.New(commandExecutionError)
	}

	// Execute the commands remote via API
	client := &http.Client{}
	for _, r := range remotes {
		log.Println("Going to create lv on remote:", r)

		req, _ := http.NewRequest("POST", fmt.Sprintf("http://%v:%v/sec/lv", r, Port), b)
		req.SetBasicAuth("GLUSTER_API", Secret)

		resp, err := client.Do(req)
		if (err != nil || resp.StatusCode != http.StatusOK) {
			if (resp != nil){
				log.Println("Remote did not respond with OK", resp.StatusCode)
			} else {
				log.Println("Connection to remote not possible", r, err.Error())
			}
			return errors.New(commandExecutionError)
		}
		resp.Body.Close()
	}

	return nil
}

func createGlusterVolume() (error) {
	// Create a gluster volume
	// gluster volume create vol_ssd replica 2 devglusternode01:/gluster/ssd1/brick1 devglusternode02:/gluster/ssd1/brick1
	// gluster volume start vol_ssd
	volCmd := fmt.Sprintf("gluster volume create %v replica %v ")

	// Append all servers here
	servers :=


	out, err := exec.Command("bash", "-c", "lvs").Output()
	if err != nil {
		log.Println("Could not count existing lvs for a project:", project, err.Error())
		return -1, err
	}


}