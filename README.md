# Point of Sale

## Introduction
This application is two things. First, it is a point of sale system designed for small businesses that have an entry fee and also sell various items. Second, it is an introductory/example/demo programming application that bridges the gap between a `hello world` or `TODO` list example and advanced tutorials.

Additionally, it combines a printer for food orders and a food notification buzzer when coupled with a Raspberry Pi.

## Dependencies
### Languages
- `go`
- `mysql`
- `python` (optional - for the buzzer)

### Hardware (optional)
- a Raspberry Pi 2 (other versions require a different printer)
- a [Mini Thermal Receipt Printer](https://www.adafruit.com/product/600)
- the buzzer from a [Sunfounder Sensor Kit](https://www.sunfounder.com/starterkit/arduino/sensor-kit-v2-0.html)

Note: the hardware will need to be connected together and configured. See the [complete setup details](INSTALL.md) for more information.

## Install
You'll need glide to install the dependencies:
- `go get github.com/Masterminds/glide`
#### Get the code
- `go get github.com/zramsay/point-of-sale`
#### Install the dependencies
- `glide install`
#### Create the database
- `mysql -u root -p -e "CREATE DATABASE myBusiness"`
#### Load it
- `mysql -u root -p myBusiness < sample.sql`
#### Compile and run the app
- `go build && ./point-of-sale -password yourMySQLpasswordHere`
#### View it
- visit http://localhost:8081 in your browser 

See [here](INSTALL.md) for detailed installation and setup instructions.

## Motivation
This app was built to address the needs of a small family business for which existing point of sale systems did not offer a solution. It doubles as an educational tool for budding developers by covering several concepts in as few lines of code as possible.
