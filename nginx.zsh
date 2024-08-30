#!/bin/zsh

# Nginx service management script wrapper
# Usage: nginxadm {setup|stop|resume|cleanup|status}
# Arguments:
#   setup: Set up nginx container, start the container
#   stop: Stop nginx container
#   resume: Resume nginx container
#   cleanup: Remove nginx container
#   status: Check nginx container status
#
# Environment variables:
#   NGINX: Path to the nginx service
#   NGINX_CONTAINER_NAME: Name of the nginx container
#   SONARQUBE_CONTAINER_NAME: Name of the SonarQube container
#   SERVICES_NETWORK: Name of the services network
#
# Dependencies: docker
#
# Author: BoxBoxJason

nginxadm() {
    # Check if required environment variables are set
    if [ -z "$NGINX" ]; then
        echo "NGINX environment variable not set. Please set it to the path of the nginx service"
        return 2
    elif [ -z "$NGINX_CONTAINER_NAME" ]; then
        echo "NGINX_CONTAINER_NAME environment variable not set. Please set it to the name of the nginx container"
        return 2
    elif [ -z "$SONARQUBE_CONTAINER_NAME" ]; then
        echo "SONARQUBE_CONTAINER_NAME environment variable not set. Please set it to the name of the SonarQube container"
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

    local IMAGE_NAME="nginx:stable-alpine"
    local HTTP_PORT="80"
    local HTTPS_PORT="443"
    local SERVICES=("sonaradm" "minioadm")
    local NGINX_CONF="$NGINX/conf"
    local KEY_PATH="/etc/ssl/private/$(hostname).key"
    local CERT_PATH="/etc/ssl/certs/$(hostname).crt"

    # Set up nginx config & start the container
    setup() {
        mkdir -p "$NGINX_CONF"
        if [[ -z $KEY_PATH || -z $CERT_PATH ]]; then
            echo "Certificate OR Key do not exist, please provide them at $KEY_PATH and $CERT_PATH for TLS setup"
            return 1
        elif docker ps --filter "name=$NGINX_CONTAINER_NAME" --filter "status=running" | grep -q $NGINX_CONTAINER_NAME; then
            echo "nginx container is already running"
            return 1
        fi
        echo "server {
    listen 80;
    server_name $(hostname);
    return 301 https://\$host\$request_uri;
}

server {
    listen 443 ssl;
    server_name $(hostname);
    ssl_certificate $CERT_PATH;
    ssl_certificate_key $KEY_PATH;
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers HIGH:!aNULL:!MD5;
" > $NGINX_CONF/nginx.conf
        for service in "${SERVICES[@]}"; do
            $service nginxconf >> $NGINX_CONF/nginx.conf
        done

        echo "}" >> $NGINX_CONF/nginx.conf
        # Start the nginx container
        docker run -d --name $NGINX_CONTAINER_NAME \
            --network $SERVICES_NETWORK \
            -v $NGINX_CONF/nginx.conf:/etc/nginx/conf.d/default.conf \
            -v $CERT_PATH:$CERT_PATH \
            -v $KEY_PATH:$KEY_PATH \
            -p $HTTP_PORT:80 \
            -p $HTTPS_PORT:443 \
            $IMAGE_NAME
    }

    # Stop nginx container
    stop() {
        docker stop $NGINX_CONTAINER_NAME && \
        echo "nginx container stopped."
    }

    # Resume nginx container
    resume() {
        docker start $NGINX_CONTAINER_NAME && \
        echo "nginx container resumed."
    }

    # Remove nginx container
    cleanup() {
        rm -rf $NGINX
        docker rm -f $NGINX_CONTAINER_NAME && \
        echo "nginx container removed."
    }

    # Get the status of the nginx container
    status() {
        if docker ps --filter "name=$NGINX_CONTAINER_NAME" | grep -q $NGINX_CONTAINER_NAME; then
            echo "Healthy"
        else
            echo "Stopped"
        fi
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
        *)
            echo "Usage: nginxadm {setup|stop|resume|cleanup|status}"
            return 1
            ;;
    esac
}

# Set up autocompletion for the nginxadm command
_nginxadm_autocomplete() {
    compadd setup stop resume cleanup status
}

compdef _nginxadm_autocomplete nginxadm
