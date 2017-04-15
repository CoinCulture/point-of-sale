package main

import (
	"os/exec"
)

func activateBuzzer() error {
	_, err = exec.Command("python", "buzz.py").Output()
	if err != nil {
		return err
	}
	return nil
}
