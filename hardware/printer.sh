#!/bin/bash

# do the time
dateTime=$(date +%T)
echo $dateTime > tempTime.txt
lpr tempTime.txt
rm tempTime.txt

# the "$1" is the filename being passed in as an argument to the bash file
lpr -o cpi=3.5 -o lpi=2 $1
