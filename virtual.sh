#!/bin/bash

# Run virtual cluster for testing.
# Make six regions, one mds and nine osds in each.

set -e

NIL=./nil
DIR=virt
PID=$DIR/pid

# Region names follow ISO-3166-1
REGIONS=("KR" "US" "HK" "SG" "JP" "DE")
GWBASEPORT=50000
MDSBASEPORT=51000
DSBASEPORT=52000

usage() {
    echo
    echo "Usage: $0 [-p] [-s] [-h]"
    echo "Options:"
    echo "  -p Purge virtual cluster"
    echo "  -s Create mysql user and schema"
    echo "  -h Show this screen"
    echo
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
        pids=$(cat $DIR/pid)
        for pid in $pids; do
            kill -9 $pid &
        done
    fi

    # Remove virtual cluster directory.
    rm -rf $DIR

    for region in ${REGIONS[@]}; do
	mysql -utestNil -pnil nil${region} -e "DROP TABLE IF EXISTS bucket;"
	mysql -utestNil -pnil nil${region} -e "DROP TABLE IF EXISTS region;"
	mysql -utestNil -pnil nil${region} -e "DROP TABLE IF EXISTS user;"
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
      --raft-local-cluster-addr localhost:$((GWBASEPORT - 1)) \
      --raft-local-cluster-region $region \
      --raft-dir $workdir/raftdir \
      -l $workdir/log &
    echo $! >> $PID
}

function runds() {
    local region="$1"
    local numds="$2"
    local port="$3"
    local workdir=$DIR/$region/ds$numds

    mkdir -p $workdir
    # Run ds.
    $NIL osd \
      -p $port \
      -l $workdir/log &
    echo $! >> $PID
}

function main() {
    purge

    for region in ${REGIONS[@]}; do
        echo "set region $region ..."
        runregion "$region" 1 1 9
        sleep 3
    done
}

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
