package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
)

// notification can be a buzzer or a light, or something else ...
func activateNotification() error {
	_, err = exec.Command("python", "hardware/notification.py").Output()
	if err != nil {
		return err
	}
	return nil
}

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

	_, err = exec.Command("./hardware/printer.sh", file.Name()).Output()
	if err != nil {
		fmt.Printf("printer error:\n%v", err)
		return err
	}

	return nil
}
