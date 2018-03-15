#!/bin/bash

# Run virtual cluster for testing.
# Make six regions, one mds and three ds in each.

set -e

NIL=./nil
DIR=virt
CERTSDIR="$(readlink -f .certs)"
PID=$DIR/pid
MNT=$DIR/mnt
PENDINGCMD=$DIR/pending

# Region names follow ISO-3166-1
REGIONS=("KR" "US" "HK" "SG" "JP" "DE")
GWBASEPORT=50000
MDSBASEPORT=51000
DSBASEPORT=52000

# Disk configuration.
DISKSIZE=100 # megabytes
DISKNUM=3    # per ds

# Save settings.
AUTOMOUNT_OPEN=""

usage() {
    echo
    echo "Usage: $0 [-p] [-s] [-h]"
    echo "Options:"
    echo "  -p Purge virtual cluster"
    echo "  -s Create mysql user and schema"
    echo "  -h Show this screen"
    echo
}

function changeset() {
    AUTOMOUNT_OPEN="$(gsettings get org.gnome.desktop.media-handling automount-open)"
    gsettings set org.gnome.desktop.media-handling automount-open false
}

function restore() {
    gsettings set org.gnome.desktop.media-handling automount-open $AUTOMOUNT_OPEN
}

function createsdisks() {
    local numdisks="$1"
    local size="$2"
    local workdir="$3"
    local dsport="$4"

    for i in $(eval echo "{1..$numdisks}"); do
        local dev=$workdir/dev$i

        # Creates a disk image.
        dd bs=1M count=$size if=/dev/zero of=$dev
        mkfs.xfs $dev

        # Add to ds.
        echo $dev >> $MNT
        echo $NIL ds volume add dev$i -p $dsport >> $PENDINGCMD
    done
}

function createschema() {
    # Read root password of mysql.
    read -sp "MySQL root password: " rootpasswd
    echo ""

    mysql -uroot -p${rootpasswd} -e "CREATE USER IF NOT EXISTS testNil@localhost IDENTIFIED BY 'nil';"

    for region in ${REGIONS[@]}; do
        mysql -uroot -p${rootpasswd} -e "CREATE DATABASE IF NOT EXISTS nil${region};"
        mysql -uroot -p${rootpasswd} -e "GRANT ALL PRIVILEGES ON nil${region}.* TO 'testNil'@'localhost';"
        mysql -uroot -p${rootpasswd} -e "FLUSH PRIVILEGES;"
    done
}

function purge() {
    # Kill all running processes of virtual cluster.
    if [ -e $PID ]; then
        pids=$(cat $PID)
        
        # Sends a stop signal to running processes.
        for pid in $pids; do
            kill $pid || true
        done

        sleep 1

        # Sends a kill signal to running processes.
        for pid in $pids; do
            kill -9 $pid || true
        done
    fi

    # Unmount all disks in the virtual cluster.
    if [ -e $MNT ]; then
        devs=$(cat $MNT)
        for dev in $devs; do
            umount $dev || true
        done
    fi

    sleep 1

    # Remove virtual cluster directory.
    rm -rf $DIR

    for region in ${REGIONS[@]}; do
	mysql -utestNil -pnil nil${region} -e "DROP TABLE IF EXISTS bucket;"
	mysql -utestNil -pnil nil${region} -e "DROP TABLE IF EXISTS region;"
	mysql -utestNil -pnil nil${region} -e "DROP TABLE IF EXISTS user;"
	mysql -utestNil -pnil nil${region} -e "DROP TABLE IF EXISTS volume;"
	mysql -utestNil -pnil nil${region} -e "DROP TABLE IF EXISTS node;"
	mysql -utestNil -pnil nil${region} -e "DROP TABLE IF EXISTS cmap;"
    done
}

function runregion() {
    local region="$1"
    local numgw="$2"
    local nummds="$3"
    local numds="$4"

    mkdir -p $DIR/$region

    # Create gw.
    for i in $(eval echo "{1..$numgw}"); do
        local port=$GWBASEPORT
        rungw "$region" "$i" "$port"
        GWBASEPORT=$((GWBASEPORT + 1))
    done

    # Create mds.
    for i in $(eval echo "{1..$nummds}"); do
        local port=$MDSBASEPORT
        runmds "$region" "$i" "$port"
        MDSBASEPORT=$((MDSBASEPORT + 1))
    done

    # Create ds.
    for i in $(eval echo "{1..$numds}"); do
        local port=$DSBASEPORT
        runds "$region" "$i" "$port"
        DSBASEPORT=$((DSBASEPORT + 1))
    done
}

function rungw() {
    local region="$1"
    local numgw="$2"
    local port="$3"
    local workdir=$DIR/$region/gw$numgw

    mkdir -p $workdir

    # Run gw.
    $NIL gw \
      -p $port \
      --first-mds localhost:$MDSBASEPORT \
      --secure-certs-dir $CERTSDIR \
      -l $workdir/log &
    echo $! >> $PID
}

function runmds() {
    local region="$1"
    local nummds="$2"
    local port="$3"
    local workdir=$DIR/$region/mds$nummds

    mkdir -p $workdir

    # Run mds.
    $NIL mds \
      -p $port \
      --mysql-user testNil \
      --mysql-database nil$region \
      --work-dir $workdir \
      --raft-local-cluster-addr localhost:$((GWBASEPORT - 1)) \
      --raft-local-cluster-region $region \
      --swim-coordinator-addr localhost:$MDSBASEPORT \
      --secure-certs-dir $CERTSDIR \
      -l log &
    echo $! >> $PID
}

function runds() {
    local region="$1"
    local numds="$2"
    local port="$3"
    local workdir=$DIR/$region/ds$numds

    mkdir -p $workdir

    # Run ds.
    $NIL ds \
      -p $port \
      --swim-coordinator-addr localhost:$((MDSBASEPORT - 1)) \
      --work-dir $workdir \
      --secure-certs-dir $CERTSDIR \
      -l log &
    echo $! >> $PID

    createsdisks "$DISKNUM" "$DISKSIZE" "$workdir" "$port"
}

function main() {
    purge

    for region in ${REGIONS[@]}; do
        echo "set region $region ..."
        runregion "$region" 1 1 3
        sleep 3
    done

    # Execute pending command.
    if [ -e $PENDINGCMD ]; then
        # Give some time to each cluster member can join the membership.
        sleep 60

        # Read line by line ...
        while read cmd; do
            $($cmd)
        done < $PENDINGCMD
    fi
}

# Run as root.
if [ $UID -ne 0 ]; then
    exec sudo -- "$0" "$@"
fi

# Change settings and save it.
# Restore it when the program is finished.
changeset

trap restore SIGINT
trap restore EXIT

while getopts psh o; do
    case $o in
    p)
        purge
        exit 0
        ;;
    s)
        createschema
        exit 0
        ;;
    h)
        usage
        exit 0
        ;;
    ?)
        usage
        exit 1
        ;;
    esac
done

main

echo "Making virtual cluster done; exit status is $?"
