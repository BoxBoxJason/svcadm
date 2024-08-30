#!/bin/zsh

# MinIO service management script wrapper
# Usage: minioadm {setup|resume|add_bucket|remove_bucket|backup|stop|cleanup|status}
# Arguments:
#   setup: Set up MinIO container, start the container
#   resume: Resume MinIO container
#   add_bucket: Add a bucket to the MinIO container
#   remove_bucket: Remove a bucket from the MinIO container
#   backup: Backup a bucket from the MinIO container
#   stop: Stop MinIO container
#   cleanup: Remove MinIO container
#   status: Check MinIO container status
#   nginxconf: Print the nginx configuration for MinIO
#
# Environment variables:
#   MINIO: Path to the MinIO service
#   MINIO_CONTAINER_NAME: Name of the MinIO container
#   SERVICES_NETWORK: Name of the services network
#
# Dependencies: docker
#
# Author: BoxBoxJason

minioadm() {
    # Check if required environment variables are set
    if [ -z "$MINIO" ]; then
        echo "MINIO environment variable not set. Please set it to the path of the MinIO service"
        return 2
    elif [ -z "$MINIO_CONTAINER_NAME" ]; then
        echo "MINIO_CONTAINER_NAME environment variable not set. Please set it to the name of the MinIO container"
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

    local IMAGE_NAME="minio/minio:latest"
    local VOLUME_NAME="minio_data"
    local CREDENTIALS_DIR="$MINIO/.credentials"
    local USERNAME_FILE="$CREDENTIALS_DIR/.miniouser"
    local PASSWORD_FILE="$CREDENTIALS_DIR/.miniopass"
    local BACKUP_DIR="$MINIO/backup"

    # Set up MinIO container and start the container
    setup() {
        if docker ps --filter "name=$MINIO_CONTAINER_NAME" --filter "status=running" | grep -q $MINIO_CONTAINER_NAME; then
            echo "MinIO container is already running"
            return 1
        fi
        mkdir -p "$CREDENTIALS_DIR"
        if [ -f "$USERNAME_FILE" ] && [ -f "$PASSWORD_FILE" ]; then
            echo "Using existing MinIO credentials"
        else
            echo "Generating random credentials for MinIO"
            echo "$(whoami)" > "$USERNAME_FILE"
            generate_password > "$PASSWORD_FILE"
        fi
        docker volume create $VOLUME_NAME
        docker run -d --name $MINIO_CONTAINER_NAME \
            --network $SERVICES_NETWORK \
            -v $VOLUME_NAME:/data \
            -e "MINIO_ROOT_USER=$(cat $USERNAME_FILE)" \
            -e "MINIO_ROOT_PASSWORD=$(cat $PASSWORD_FILE)" \
            $IMAGE_NAME server /data --console-address ":9001"
        if [ $? -ne 0 ]; then
            echo "Error starting MinIO container"
            return 1
        fi
        echo "MinIO container started successfully"
    }

    # Resume MinIO container
    resume() {
        if docker ps --filter "name=$MINIO_CONTAINER_NAME" --filter "status=running" | grep -q $MINIO_CONTAINER_NAME; then
            echo "MinIO container is already running"
            return 1
        fi
        echo "Resuming MinIO container"
        docker start $MINIO_CONTAINER_NAME
    }

    # Stop MinIO container
    stop() {
        docker stop $MINIO_CONTAINER_NAME && \
        echo "MinIO container stopped."
    }

    # Remove MinIO container
    cleanup() {
        docker rm -f $MINIO_CONTAINER_NAME
        docker volume rm $VOLUME_NAME
        rm -rf $CREDENTIALS_DIR
        if [ -d "$BACKUP_DIR" ]; then
            echo "Keeping backup directory $BACKUP_DIR, you may want to remove it manually"
        fi
    }

    # Check MinIO container status
    status() {
        if docker ps --filter "name=$MINIO_CONTAINER_NAME" --filter "status=running" | grep -q $MINIO_CONTAINER_NAME; then
            echo "Healthy"
            return 0
        else
            echo "Stopped"
            return 1
        fi
    }

    # Add a bucket to the MinIO container
    add_bucket() {
        if [ -z "$1" ]; then
            echo "Usage: minioadm add_bucket <bucket>"
            return 2
        elif docker ps --filter "name=$MINIO_CONTAINER_NAME" --filter "status=running" | grep -q $MINIO_CONTAINER_NAME; then
            echo "Adding bucket $1 to MinIO container"
            docker exec $MINIO_CONTAINER_NAME mc mb "minio/$1"
        else
            echo "MinIO container is not running, please use 'minioadm resume' to start the container"
            return 1
        fi
    }

    # Remove a bucket from the MinIO container
    remove_bucket() {
        if [ -z "$1" ]; then
            echo "Usage: minioadm remove_bucket <bucket>"
            return 2
        elif docker ps --filter "name=$MINIO_CONTAINER_NAME" --filter "status=running" | grep -q $MINIO_CONTAINER_NAME; then
            echo "Removing bucket $1 from MinIO container"
            docker exec $MINIO_CONTAINER_NAME mc rb --force "minio/$1"
        else
            echo "MinIO container is not running, please use 'minioadm resume' to start the container"
            return 1
        fi
    }

    # Backup a bucket from the MinIO container
    backup() {
        if [ -z "$1" ]; then
            echo "Usage: minioadm backup <buckets (space separated)>"
            return 2
        elif docker ps --filter "name=$MINIO_CONTAINER_NAME" --filter "status=running" | grep -q $MINIO_CONTAINER_NAME; then
            mkdir -p "$BACKUP_DIR"
            if [ "$1" = "*" ]; then
                echo "Backing up all buckets from MinIO container"
                local BUCKETS=$(docker exec $MINIO_CONTAINER_NAME mc ls "minio" | awk '{print $2}')
                for bucket in $BUCKETS; do
                    backup $bucket
                done
            else
                echo "Backing up bucket $1 from MinIO container"
                docker exec $MINIO_CONTAINER_NAME mc mirror "minio/$1" "$BACKUP_DIR/$1" && \
                docker exec $MINIO_CONTAINER_NAME tar -czf "/data/$1.tar.gz" -C "/data" "$1" && \
                docker cp "$MINIO_CONTAINER_NAME:/data/$1.tar.gz" "$BACKUP_DIR"
                docker exec $MINIO_CONTAINER_NAME rm "/data/$1.tar.gz"
                if [ -n "$2" ]; then
                    shift
                    backup "$@"
                fi
            fi
        else
            echo "MinIO container is not running, please use 'minioadm resume' to start the container"
            return 1
        fi
    }

    # Print the nginx configuration for MinIO
    nginxconf() {
        cat <<EOF
    location /minio {
        proxy_pass http://$MINIO_CONTAINER_NAME:9001;
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;
        rewrite ^/minio/(.*)$ /$1 break;
    }

    location /minio-api {
        proxy_pass http://$MINIO_CONTAINER_NAME:9000;
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;
        rewrite ^/minio-api/(.*)$ /$1 break;
    }
EOF
    }


    case "$cmd" in
        setup)
            setup
            ;;
        resume)
            resume
            ;;
        stop)
            stop
            ;;
        cleanup)
            cleanup
            ;;
        status)
            status
            ;;
        add_bucket)
            add_bucket "$@"
            ;;
        remove_bucket)
            remove_bucket "$@"
            ;;
        backup)
            backup "$@"
            ;;
        nginxconf)
            nginxconf
            ;;
        *)
            echo "Usage: minioadm {setup|resume|add_bucket|remove_bucket|backup|stop|cleanup|status|nginxconf}"
            return 2
            ;;
    esac
}

# Set up autocompletion for the minioadm command
_minioadm_autocomplete() {
    compadd setup resume add_bucket remove_bucket backup stop cleanup status
}

compdef _minioadm_autocomplete minioadm
