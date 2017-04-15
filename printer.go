package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
)

// https://learn.adafruit.com/networked-thermal-printer-using-cups-and-raspberry-pi/overview
// hacked together from the above tutorial

func printTheChit(braceletNum int, amount, foodItem string) error {
	file, err := ioutil.TempFile(os.TempDir(), "toPrint")
	defer os.Remove(file.Name())

	number := fmt.Sprintf("#%v", braceletNum)
	item := strings.ToUpper(foodItem)
	amount = fmt.Sprintf("x%s", amount)
	space := "______"

	_, err = file.WriteString(fmt.Sprintf("%s\n%s\n%s\n%s", number, item, amount, space))
	if err != nil {
		return err
	}

	_, err = exec.Command("lpr", "-o cpi=3.5", "-o lpi=2", file.Name()).Output()
	if err != nil {
		return err
	}

	return nil
}
