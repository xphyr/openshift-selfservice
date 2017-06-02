package volumes

import (
	"strings"
	"strconv"
	"log"
	"fmt"
)

const (
	suffixWrongError = "Invalid size. Size must be int followed by suffix (e.g. 100M). Allowed suffixes are 'G/M'. You sent: "
)

func CreateVolume(project string, size string) (bool, string) {
	// Check if free space is ok
	isOk, msg := validateSizeInput(size)
	if (!isOk) {
		log.Print("Aborting...", msg)
		return isOk, msg
	}

	// Create lvs on pool on all gluster servers
	createLvOnPool(project, size)

	// Create gluster volume

	// Add snapshots

	return true, ""
}

func createLvOnPool(project string, size string) (bool, string) {
	// TODO: Get this from counting
	pvName := "pv1"

	mountPoint := fmt.Sprintf("%v/%v/%v", BasePath, project, pvName)
	lvName := fmt.Sprintf("lv_%v_%v", project, pvName)

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

	for _, c := range commands {
		log.Println(c)
	}

	// Execute this on all gluster servers

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
