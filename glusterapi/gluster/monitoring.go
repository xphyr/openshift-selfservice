package gluster

import (
	"github.com/oscp/openshift-selfservice/glusterapi/models"
	"regexp"
	"strconv"
	"fmt"
	"os/exec"
	"log"
	"errors"
)

func getVolumeUsage(pvName string) (*models.VolInfo, error) {
	cmd := fmt.Sprintf("df --output=size,used,source | grep %v", pvName)

	out, err := exec.Command("bash", "-c", cmd).Output()
	if err != nil {
		msg := "Could not parse usage size: " + err.Error()
		log.Println(msg)
		return nil, errors.New(msg)
	}

	volInfo, err := parseOutput(string(out))
	if err != nil {
		return nil, err
	}
	return volInfo, nil
}

func parseOutput(stdOut string) (*models.VolInfo, error) {
	// Examples
	// 5472   118048 /dev/mapper/vg_slow-lv_test_pv5
	num := regexp.MustCompile("(\\d+)")
	nums := num.FindAllString(stdOut, -1)

	size, err := strconv.Atoi(nums[0])
	if (err != nil) {
		log.Println("Unable to parse size value of df output", stdOut)
		return nil, errors.New(commandExecutionError)
	}

	used, err := strconv.Atoi(nums[1])
	if (err != nil) {
		log.Println("Unable to parse used value of df output", stdOut)
		return nil, errors.New(commandExecutionError)
	}

	return &models.VolInfo{
		TotalKiloBytes: size,
		UsedKiloBytes: used,
	}, nil
}

func checkVolumeUsage(pvName string, treshold string) (error) {
	t, err := strconv.ParseFloat(treshold, 64)
	if (err != nil) {
		return errors.New("Wrong treshold. Is not a valid integer")
	}

	volInfo, err := getVolumeUsage(pvName)
	if (err != nil) {
		return err
	}

	usedPercentage := 100 / float64(volInfo.TotalKiloBytes) * float64(volInfo.UsedKiloBytes)
	if (usedPercentage > t) {
		return fmt.Errorf("Error used %v is bigger than treshold: %v", usedPercentage, t)
	}

	return nil
}
