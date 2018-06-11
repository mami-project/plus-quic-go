#!/bin/sh

CERTS="./rootCACert.pem"
COLLECT="16"
IFACE="eno1"
LOGFILE=""
ODIR="./tmp/"
URLS="https://localhost:6121/data/256;https://localhost:6121/data/4KiB"
RUNS="4"
RESULTDIR="/monroe/results"
WAITTO=100
WAITFROM=10

while true
do
	echo "Invoking puic-poll..."
	./puic-poll -certs=$CERTS -wait-to=$WAITTO -wait-from=$WAITFROM -collect=$COLLECT -iface=$IFACE -logfile=$LOGFILE -odir=$ODIR -urls=$URLS -runs=$RUNS
	echo "Moving results..."
	mv $ODIR/puic-poll* $RESULTDIR
done
