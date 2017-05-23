# Point of Sale

Here we cover the installation and configuration of software and hardware from start to finish. We'll assume all the hardware listed in the README has been acquired. The end result is a stand-alone Raspberry Pi (RPi) running the point of sale application as a web server. The RPi is connected to your home (or business) router; this allows any other device to access the application

## Raspberry Pi (RPi)
- these instructions apply to the RPi 2 with Raspbian installed. See the [official documentation](https://www.raspberrypi.org/documentation/) for setup information. Your RPi should automatically have access to the internet once connected to a router via a Ethernet cable.
- without proper configuration, a RPi will be at risk of being hacked. We've found [this tutorial](https://mattwilcox.net/web-development/setting-up-a-secure-home-web-server-with-raspberry-pi) particularly useful for a basic security setup (stop reading at "Install web server software") although a google search for "raspberry pi secure configuration" will go a long way.
- if skipping a secure configuration, you **MUST:** 1) flash your memory card, 2) re-install Raspbian, and 3) start over by securing your RPi **before** running an application that you plan on exposing to the world.

### Install Go
- this application is written using the [Go programming language](https://golang.org/). To make edits to the code and to install the program (compile the code into a binary), we need to install Go. Later in the tutorial, you'll see that we're using the Python programming language: it comes pre-installed on the Raspbian operating system.
- run `sudo apt-get install golang` and you should be good to "go".

### Install MySQL
- while the "logic" (how it operates) of the application is written using Go, we're also going to need a database to store & retrieve information as required by the application. We're using MySQL is the world's most popular open source database. The SQL stands for "Structured Query Language". Within the Go code we connect and "talk to" the MySQL server. It is also possible to talk to the MySQL server via the MySQL command line tool.
- run `sudo apt-get install mysql-server`; you'll be prompted for a root password. Choose a strong one and remember it/write it down
- to ensure a secure installation, run `mysql_secure_installation` and enter `Y` all the way down.
- now we need to tell MySQL to create a new database; run `mysql -u root -p -e "CREATE DATABASE myBusiness` 
- with the database in existence, we can then load the sample.sql file: `mysql -u root -p myBusiness < sample.sql`. This file contains MySQL tables (visits, transactions, items) used by the Go application to manage each category. Once the database is loaded, you can edit the rows manually from the `mysql` tool. For your custom application, you'll definitely want to edit the `items` tables to contain the food/drink/miscellaneous items that you plan on selling.

### Get the source code
- you'll need the code from this repository to install the application
- if `git` is not already installed, run `sudo apt-get install git`
- then run `git clone https://github.com/CoinCulture/point-of-sale.git`, which will "clone" the repository and all of its code to your RPi
- next, enter the directory in which the repo was cloned: `cd point-of-sale/`
- now, we need some dependencies: `glide install`
- finally, build and run the application: `go build && ./point-of-sale -password yourMYSQLrootPasswordHere`
- note that the previous command does two things. First, `go build` *compiles the application into a binary*. This has to be done after every edit to the go code. This binary is placed in the current working directory (the folder you are in, try running `ls` to see its files). The `&&` seperates the two commands and runs them in sequence. You could just as easily run `go build`, wait, then run `./point-of-sale -password yourMYSQLrootPasswordHere`. The latter command runs the binary with the `-password` *flag*, which takes the actual password as an *argument*.
- the app should now be running. To access it locally (i.e., from the RPi directly) visit `localhost:8080` in the browser. To access it from another device connected to the same router, first [find the IP of your RPi](https://www.raspberrypi.org/documentation/remote-access/ip-address.md) then visit `IP:8080` in the browser. Using `nmap` (per the linked tutorial) is the usual way of getting IP addresses of connected devices.

This is all fun and well, however, we have not integrated our hardware. Most of the app will work, but the printer and buzzer features need to be configured.

## Printer
- the printer in question is the [Mini Thermal Receipt Printer](https://www.adafruit.com/product/600) from Adafruit. It comes in less-expensive kits that are more bare-bones, but we definitely recommend this package for ease of setup.
- there are several good printer tutorials on the Adafruit website (and they should all be read if you plan on experimenting). We found the most useful/relevant to be the [Networked Thermal Printer using RPi and CUPS](https://learn.adafruit.com/networked-thermal-printer-using-cups-and-raspberry-pi/overview), despite the fact that we are not (necesarilly) interested in networking the printer and only care about the `lpr` part of CUPS. The tutorial shows basic cable connections that need to be made and the commands required to run.
- it turned out to be easier to bypass/omit (some) of the steps involving zj-58 (but go through them all at least once). For example, notice that we call this command: `lpr -o cpi=3.5 -o lpi=2 fileName`, after having written a temporary file with the relevant information for the printer. The `cpi` and `lpi` arguments to the `-o` flag determines the size of the text printed. Although these features are available to be set in the zj-58 printer control panel, it will be important to have several sizes of text, for example, when implementing timestamps. Some knowledge of `lpr` (see [here, as the tutorial recommends](https://www.cups.org/doc/options.html)) will be important to customize your application. 
- that's basically it. A one line command with a couple flag options and a file of text. Issues #3, #6, and #17 all deal with improvements involving the printer.

## Buzzer
- an awesome sensor kit for the Raspberry Pi 2 is the [Sunfounder Sensor Kit](https://www.sunfounder.com/starterkit/arduino/sensor-kit-v2-0.html). At $100 on Amazon, it's a great value for the amount of sensors you get. The manual is clear and consice, and the code is [available on GitHub](https://github.com/sunfounder/SunFounder_SensorKit_for_RPi2). The "Active Buzzer" is the simplest and it is lesson 10 on page 55. We're using a modification of [the original code](https://github.com/sunfounder/SunFounder_SensorKit_for_RPi2/blob/master/Python/10_active_buzzer.py) that's essentially slimmed down for simplification.
- once you've connected the buzzer to the RPi and tested (with the original python script, for example) it, you're ready to go.
- you might be asking yourself, what is this buzzer for? Well, it turns out that the cook sometimes likes to snooze in the back room. Since he's not in the kitchen where the food orders are printing, an easy way to notify him was necessary.

That's basically it, re-run the app now that your hardware is hooked up and start placing orders!
