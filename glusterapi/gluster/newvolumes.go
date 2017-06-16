package gluster

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"strconv"
	"strings"

	"github.com/oscp/cloud-selfservice-portal/glusterapi/models"
)

const (
	suffixWrongError      = "Invalid size. Size must be int followed by suffix (e.g. 100M). Allowed suffixes are 'G/M'. You sent: %v"
	commandExecutionError = "Error running command, see logs for details"
)

func createVolume(project string, size string) (string, error) {
	if len(size) == 0 || len(project) == 0 {
		return "", errors.New("Not all input values provided")
	}

	if err := validateSizeInput(size); err != nil {
		return "", err
	}

	pvNumber, err := getExistingLvForProject(project)
	if err != nil {
		return "", err
	}

	mountPoint := fmt.Sprintf("%v/%v/pv%v", BasePath, project, pvNumber)
	lvName := fmt.Sprintf("lv_%v_pv%v", project, pvNumber)

	// Create lvs on pool on all gluster servers
	if err := createLvOnAllServers(size, mountPoint, lvName); err != nil {
		return "", err
	}

	// Create gluster volume
	if err := createGlusterVolume(project, pvNumber, mountPoint); err != nil {
		return "", err
	}

	return fmt.Sprintf("%v_pv%v", project, pvNumber), nil
}

func createLvOnPool(size string, mountPoint string, lvName string) error {
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

		// Create brick folder
		fmt.Sprintf("mkdir %v/brick", mountPoint),

		// Handle Selinux
		fmt.Sprintf("semanage fcontext -a -t glusterd_brick_t %v/brick", mountPoint),
		fmt.Sprintf("restorecon -Rv %v/brick", mountPoint),
	}

	if err := executeCommandsLocally(commands); err != nil {
		return err
	}

	return nil
}

func validateSizeInput(size string) error {
	log.Println("Checking size of", size)

	if strings.HasSuffix(size, "M") {
		sizeInt, err := strconv.Atoi(strings.Replace(size, "M", "", 1))
		if err != nil {
			return fmt.Errorf(suffixWrongError, size)
		}

		if sizeInt <= MaxMB {
			return nil
		}

		return errors.New("Your size is to big for suffix 'M' use 'G' instead")
	}
	if strings.HasSuffix(size, "G") {
		sizeInt, err := strconv.Atoi(strings.Replace(size, "G", "", 1))
		if err != nil {
			return fmt.Errorf(suffixWrongError, size)
		}

		if sizeInt > MaxGB {
			return fmt.Errorf("Max allowed size exceeded. Max allowed is: %vG", MaxGB)
		}

		return nil
	}

	return fmt.Errorf(suffixWrongError, size)
}

func getExistingLvForProject(project string) (int, error) {
	out, err := exec.Command("bash", "-c", "lvs -o lv_name").Output()
	if err != nil {
		log.Printf("Could not count existing lvs for a project: %v. Error: %v", project, err.Error())
		return -1, errors.New(commandExecutionError)
	}

	return strings.Count(string(out), "lv_"+project) + 1, nil
}

func createLvOnAllServers(size string, mountPoint string, lvName string) error {
	// Create the lv on all other gluster servers
	if err := createLvOnOtherServers(size, mountPoint, lvName); err != nil {
		return err
	}

	// Create the lv locally
	if err := createLvOnPool(size, mountPoint, lvName); err != nil {
		return err
	}

	return nil
}

func createLvOnOtherServers(size string, mountPoint string, lvName string) error {
	remotes, err := getGlusterPeerServers()
	if err != nil {
		return err
	}

	p := models.CreateLVCommand{
		LvName:     lvName,
		MountPoint: mountPoint,
		Size:       size,
	}
	b := new(bytes.Buffer)

	if err = json.NewEncoder(b).Encode(p); err != nil {
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
		if err != nil || resp.StatusCode != http.StatusOK {
			if resp != nil {
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

func createGlusterVolume(project string, pvNumber int, mountPoint string) error {
	// Create a gluster volume
	// gluster volume create vol_ssd replica 2 devglusternode01:/gluster/ssd1/brick1 devglusternode02:/gluster/ssd1/brick1
	// gluster volume start vol_ssd
	volCmd := fmt.Sprintf("gluster volume create vol_%v_pv%v replica %v ", project, pvNumber, Replicas)

	// Add all remote servers
	servers, err := getGlusterPeerServers()
	if err != nil {
		return err
	}

	localIP, err := getLocalServersIP()
	if err != nil {
		return err
	}

	servers = append(servers, localIP)

	for _, r := range servers {
		volCmd += fmt.Sprintf("%v:%v/brick ", r, mountPoint)
	}

	commands := []string{
		volCmd,

		fmt.Sprintf("gluster volume start vol_%v_pv%v", project, pvNumber),
	}

	if err := executeCommandsLocally(commands); err != nil {
		return err
	}

	return nil
}
