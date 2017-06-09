package gluster

import (
	"os/exec"
	"log"
	"errors"
)

func executeCommandsLocally(commands []string) (error) {
	for _, c := range commands {
		out, err := exec.Command("bash", "-c", c).Output()
		if (err != nil) {
			log.Println("Error executing command: ", c, err.Error(), out)
			return errors.New((commandExecutionError)
		}
	}

	return nil
}

