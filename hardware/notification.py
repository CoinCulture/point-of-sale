#!/usr/bin/env python
import RPi.GPIO as GPIO
import time

''' commented out buzzer; implemented light below
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
'''

colors = [0xFF00, 0x00FF, 0x0FF0, 0xF00F]
pins = (13, 15)  # pins is a dict

GPIO.setmode(GPIO.BOARD)       # Numbers GPIOs by physical location
GPIO.setup(pins, GPIO.OUT)   # Set pins' mode is output
GPIO.output(pins, GPIO.LOW)  # Set pins to LOW(0V) to off led

p_R = GPIO.PWM(pins[0], 2000)  # set Frequece to 2KHz
p_G = GPIO.PWM(pins[1], 2000)

p_R.start(0)      # Initial duty Cycle = 0(leds off)
p_G.start(0)

def map(x, in_min, in_max, out_min, out_max):
	return (x - in_min) * (out_max - out_min) / (in_max - in_min) + out_min

def setColor(col):   # For example : col = 0x1122
	R_val = col  >> 8
	G_val = col & 0x00FF

	R_val = map(R_val, 0, 255, 0, 100)
	G_val = map(G_val, 0, 255, 0, 100)

	p_R.ChangeDutyCycle(R_val)     # Change duty cycle
	p_G.ChangeDutyCycle(G_val)

def loop():
	setColor(colors[0])
	time.sleep(0.25)

	setColor(colors[1])
	time.sleep(0.25)

	setColor(colors[2])
	time.sleep(0.25)

	setColor(colors[3])
	time.sleep(0.25)

def destroy():
	p_R.stop()
	p_G.stop()
	GPIO.output(pins, GPIO.LOW)    # Turn off all leds
	GPIO.cleanup()

if __name__ == "__main__":
	loop()
	destroy()
