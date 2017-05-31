#!/bin/bash

# the "$1" is the filename being passed in as an argument to the bash file
lpr -o cpi=3.5 -o lpi=2 $1
