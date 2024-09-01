#!/bin/zsh

# ClamAV service management script wrapper
# Usage: clamavadm {setup|stop|resume|cleanup|status}
# Arguments:
#   setup: Set up ClamAV container, start the container
#   stop: Stop ClamAV container
#   resume: Resume ClamAV container
#   cleanup: Remove ClamAV container
#   status: Check ClamAV container status
#   nginxconf: Print the nginx configuration for ClamAV
#
# Environment variables:
#   CLAMAV: Path to the ClamAV service
#   CLAMAV_CONTAINER_NAME: Name of the ClamAV container
#   SERVICES_NETWORK: Name of the services network
#
# Dependencies: docker
#
# Author: BoxBoxJason

clamadm() {
    # Check if required environment variables are set
    if [ -z "$CLAMAV" ]; then
        echo "CLAMAV environment variable not set. Please set it to the path of the ClamAV service"
        return 2
    elif [ -z "$CLAMAV_CONTAINER_NAME" ]; then
        echo "CLAMAV_CONTAINER_NAME environment variable not set. Please set it to the name of the ClamAV container"
        return 2
    elif [ -z "$SERVICES_NETWORK" ]; then
        echo "SERVICES_NETWORK environment variable not set. Please set it to the name of the services network"
        return 2
    elif ! docker network ls | grep -q $SERVICES_NETWORK; then
        echo "Services network not found. Please create it using 'docker network create $SERVICES_NETWORK'"
        return 2
    fi

    local cmd=$1
    shift

    local IMAGE_NAME="clamav/clamav:latest"
    local CLAMAV_CONF="$CLAMAV/scan.conf"

    # Set up ClamAV container and start the container
    setup() {
        if docker ps --filter "name=$CLAMAV_CONTAINER_NAME" --filter "status=running" | grep -q $CLAMAV_CONTAINER_NAME; then
            echo "ClamAV container is already running"
            return 1
        fi
        mkdir -p "$CLAMAV"
        cat <<EOF > $CLAMAV_CONF
LocalSocket /run/clamd.scan/clamd.sock
LocalSocketMode 660
TCPSocket 3310
TCPAddr 0.0.0.0
ConcurrentDatabaseReload no
User clamav
Foreground yes
DisableCache yes
HeuristicScanPrecedence yes
AlertBrokenExecutables yes
AlertBrokenMedia yes
EOF
        docker run -d --name $CLAMAV_CONTAINER_NAME --network $SERVICES_NETWORK -v $CLAMAV_CONF:/etc/clamav/scan.conf $IMAGE_NAME
    }

    # Stop ClamAV container
    stop() {
        if ! docker ps --filter "name=$CLAMAV_CONTAINER_NAME" --filter "status=running" | grep -q $CLAMAV_CONTAINER_NAME; then
            echo "ClamAV container is not running"
            return 1
        fi
        docker stop $CLAMAV_CONTAINER_NAME
    }

    # Resume ClamAV container
    resume() {
        if docker ps --filter "name=$CLAMAV_CONTAINER_NAME" --filter "status=running" | grep -q $CLAMAV_CONTAINER_NAME; then
            echo "ClamAV container is already running"
            return 1
        fi
        docker start $CLAMAV_CONTAINER_NAME
    }

    # Remove ClamAV container
    cleanup() {
        docker rm -f $CLAMAV_CONTAINER_NAME
        rm -rf $CLAMAV
    }

    # Get the status of the ClamAV container
    status() {
        if docker ps --filter "name=$CLAMAV_CONTAINER_NAME" | grep -q $CLAMAV_CONTAINER_NAME; then
            echo "Healthy"
        else
            echo "Stopped"
        fi
    }

    # Scan a file existing on the host filesystem
    scan() { // TODO
        echo "Not implemented"
        return 2
    }

    # Print the nginx configuration for ClamAV
    nginxconf() {
        cat <<EOF
    # ClamAV
    upstream clamav {
        server $CLAMAV_CONTAINER_NAME:3310;
    }
    server {
        listen 3310;
        server_name $(hostname);
        proxy_pass clamav;
    }
EOF
    }

    case "$cmd" in
        setup)
            setup
            ;;
        stop)
            stop
            ;;
        resume)
            resume
            ;;
        cleanup)
            cleanup
            ;;
        status)
            status
            ;;
        nginxconf)
            nginxconf
            ;;
        scan)
            scan $@
            ;;
        *)
            echo "Usage clamadm {setup|stop|resume|cleanup|status|nginxconf|scan}"
            return 2
            ;;
    esac
}

# Set up autocompletion for the clamadm command
_clamadm_autocomplete() {
    compadd setup stop resume cleanup status nginxconf scan
}

compdef _clamadm_autocomplete clamadm
