#!/bin/zsh

# Services management script wrapper
# Usage: svcadm {setup|stop|resume|cleanup|status}
# Arguments:
#   setup: Set up services containers, volumes and network, start the containers
#   stop: Stop services containers
#   resume: Resume services containers
#   cleanup: Remove services containers, volumes and network
#   status: Check services containers status
#
# Environment variables:
#   SERVICES: Path to the services directory
#   POSTGRESQL: Path to the PostgreSQL service
#   POSTGRESQL_CONTAINER_NAME: Name of the PostgreSQL container
#   SONARQUBE: Path to the SonarQube service
#   SONARQUBE_CONTAINER_NAME: Name of the SonarQube container
#   NGINX: Path to the nginx service
#   NGINX_CONTAINER_NAME: Name of the nginx container
#   MINIO: Path to the MinIO service
#   MINIO_CONTAINER_NAME: Name of the MinIO container
#   SERVICES_NETWORK: Name of the services network
#
# Dependencies: docker, jq, curl
#
# Author: BoxBoxJason

# Services environment variables
export SERVICES=~/services
export SERVICES_NETWORK="services-network"

# PostgreSQL environment variables
export POSTGRESQL=$SERVICES/postgresql
export POSTGRESQL_CONTAINER_NAME="postgresql"

# SonarQube environment variables
export SONARQUBE=$SERVICES/sonarqube
export SONARQUBE_CONTAINER_NAME="sonarqube"

# Nginx environment variables
export NGINX=$SERVICES/nginx
export NGINX_CONTAINER_NAME="nginx"

# MinIO environment variables
export MINIO=$SERVICES/minio
export MINIO_CONTAINER_NAME="minio"

# ClamAV environment variables
export CLAMAV=$SERVICES/clamav
export CLAMAV_CONTAINER_NAME="clamav"

generate_password() {
    cat /dev/urandom | tr -dc 'a-zA-Z0-9' | fold -w 32 | head -n 1
}

svcadm() {
    # Check if required environment variables are set
    if [ -z "$SERVICES" ]; then
        echo "SERVICES environment variable not set. Please set it to the path of the services directory"
        return 2
    elif [ -z "$POSTGRESQL" ]; then
        echo "POSTGRESQL environment variable not set. Please set it to the path of the PostgreSQL service"
        return 2
    elif [ -z "$POSTGRESQL_CONTAINER_NAME" ]; then
        echo "POSTGRESQL_CONTAINER_NAME environment variable not set. Please set it to the name of the PostgreSQL container"
        return 2
    elif [ -z "$SONARQUBE" ]; then
        echo "SONARQUBE environment variable not set. Please set it to the path of the SonarQube service"
        return 2
    elif [ -z "$SONARQUBE_CONTAINER_NAME" ]; then
        echo "SONARQUBE_CONTAINER_NAME environment variable not set. Please set it to the name of the SonarQube container"
        return 2
    elif [ -z "$NGINX" ]; then
        echo "NGINX environment variable not set. Please set it to the path of the nginx service"
        return 2
    elif [ -z "$NGINX_CONTAINER_NAME" ]; then
        echo "NGINX_CONTAINER_NAME environment variable not set. Please set it to the name of the nginx container"
        return 2
    elif [ -z "$MINIO" ]; then
        echo "MINIO environment variable not set. Please set it to the path of the MinIO service"
        return 2
    elif [ -z "$MINIO_CONTAINER_NAME" ]; then
        echo "MINIO_CONTAINER_NAME environment variable not set. Please set it to the name of the MinIO container"
        return 2
    elif [ -z "$SERVICES_NETWORK" ]; then
        echo "SERVICES_NETWORK environment variable not set. Please set it to the name of the services network"
        return 2
    fi

    local cmd=$1
    shift

    mkdir -p $SERVICES

    # Services network
    if ! docker network ls | grep -q $SERVICES_NETWORK; then
        docker network create $SERVICES_NETWORK
    fi

    # Set up services containers, volumes and network, start the containers
    setup() {
        if docker ps --filter "name=$POSTGRESQL_CONTAINER_NAME" --filter "status=running" | grep -q $POSTGRESQL_CONTAINER_NAME; then
            echo "PostgreSQL container is already running"
            return 1
        elif docker ps --filter "name=$SONARQUBE_CONTAINER_NAME" --filter "status=running" | grep -q $SONARQUBE_CONTAINER_NAME; then
            echo "SonarQube container is already running"
            return 1
        elif docker ps --filter "name=$NGINX_CONTAINER_NAME" --filter "status=running" | grep -q $NGINX_CONTAINER_NAME; then
            echo "nginx container is already running"
            return 1
        elif docker ps --filter "name=$MINIO_CONTAINER_NAME" --filter "status=running" | grep -q $MINIO_CONTAINER_NAME; then
            echo "MinIO container is already running"
            return 1
        fi
        psqladm setup
        sonaradm setup
        minioadm setup
        nginxadm setup
    }

    backup() {
        psqladm backup *
        sonaradm backup
        minioadm backup *
    }

    # Stop services containers
    stop() {
        nginxadm stop
        sonaradm stop
        minioadm stop
        psqladm stop
    }

    # Resume services containers
    resume() {
        psqladm resume
        sonaradm resume
        minioadm resume
        nginxadm resume
    }

    # Remove services containers, volumes and network
    cleanup() {
        nginxadm cleanup
        sonaradm cleanup
        psqladm cleanup
        minioadm cleanup
        docker network rm -f $SERVICES_NETWORK
    }

    # Check services containers status
    status() {
        echo "PostgreSQL container status:"
        psqladm status
        echo "SonarQube container status:"
        sonaradm status
        echo "nginx container status:"
        nginxadm status
        echo "MinIO container status:"
        minioadm status
    }

    case $cmd in
        setup)
            setup
            ;;
        backup)
            backup
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
        *)
            echo "Usage: svcadm {setup|backup|stop|resume|cleanup|status}"
            return 2
            ;;
    esac
}

# Set up autocompletion for the svcadm command
_svcadm_autocomplete() {
    compadd setup backup stop resume cleanup status
}

compdef _svcadm_autocomplete svcadm
