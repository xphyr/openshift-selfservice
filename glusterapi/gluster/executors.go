package gluster

import (
	"os/exec"
	"log"
	"errors"
)

func executeCommandsLocally(commands []string) (error) {
	log.Println("Got new commands to execute:")
	for _, c := range commands {
		out, err := exec.Command("bash", "-c", c).Output()
		if (err != nil) {
			log.Println("Error executing command: ", c, err.Error(), string(out))
			return errors.New(commandExecutionError)
		}
		log.Printf("Cmd: %v | StdOut: %v", c, string(out))
	}

	return nil
}

