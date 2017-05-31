#!/usr/bin/env python
import RPi.GPIO as GPIO
import time

Buzzer = 11    # pin11

def setup(pin):
	global BuzzerPin
	BuzzerPin = pin
	GPIO.setmode(GPIO.BOARD)       # Numbers GPIOs by physical location
	GPIO.setup(BuzzerPin, GPIO.OUT)
	GPIO.output(BuzzerPin, GPIO.HIGH)

def on():
	GPIO.output(BuzzerPin, GPIO.LOW)

def off():
	GPIO.output(BuzzerPin, GPIO.HIGH)

def beep(x):
	on()
	time.sleep(x)

def loop():
	beep(0.25)

def destroy():
	GPIO.output(BuzzerPin, GPIO.HIGH)
	GPIO.cleanup()                     # Release resource

if __name__ == '__main__':     # Program start from here
	setup(Buzzer)
	loop()
	destroy()

