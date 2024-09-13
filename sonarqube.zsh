#!/bin/zsh

# SonarQube service management script wrapper
# Usage: sonaradm {setup|scan|backup|stop|resume|cleanup|status}
# Arguments:
#   setup: Set up SonarQube container, volumes and database start the container
#   scan: Run a SonarQube scan on a project
#   backup: Backup SonarQube data
#   stop: Stop SonarQube container
#   resume: Resume SonarQube container
#   cleanup: Remove SonarQube container, volumes and network
#   status: Check SonarQube container status
#   nginxconf: Print the nginx configuration for SonarQube
#
# Environment variables:
#   SONARQUBE: Path to the SonarQube service
#   SONARQUBE_CONTAINER_NAME: Name of the SonarQube container
#   POSTGRESQL_CONTAINER_NAME: Name of the PostgreSQL container
#   SERVICES_NETWORK: Name of the services network
#
# Dependencies: jq, curl, docker
#
# Author: BoxBoxJason

sonaradm() {
    # Check if required environment variables are set
    if [ -z "$SONARQUBE" ]; then
        echo "SONARQUBE environment variable not set. Please set it to the path of the SonarQube service"
        return 2
    elif [ -z "$SONARQUBE_CONTAINER_NAME" ]; then
        echo "SONARQUBE_CONTAINER_NAME environment variable not set. Please set it to the name of the SonarQube container"
        return 2
    elif [ -z "$POSTGRESQL_CONTAINER_NAME" ]; then
        echo "POSTGRESQL_CONTAINER_NAME environment variable not set. Please set it to the name of the PostgreSQL container"
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

    local IMAGE_NAME="sonarqube:10.6-community"
    local BACKUP_DIR="$SONARQUBE/backup"
    local SONAR_DB_NAME="sonarqube"
    local SONAR_DB_USER="sonar"

    # Set up SonarQube volumes, create the database and start the container
    setup() {
        if docker ps --filter "name=$SONARQUBE_CONTAINER_NAME" --filter "status=running" | grep -q $SONARQUBE_CONTAINER_NAME; then
            echo "SonarQube container is already running"
            return 1
        fi
        echo "Creating Docker volumes for SonarQube..."
        docker volume create sonarqube_data
        docker volume create sonarqube_extensions
        docker volume create sonarqube_logs
        echo "Starting SonarQube container"
        # Create the SonarQube database
        if docker ps --filter "name=$POSTGRESQL_CONTAINER_NAME" --filter "status=running" | grep -q $POSTGRESQL_CONTAINER_NAME; then
            psqladm add_database $SONAR_DB_NAME $SONAR_DB_USER
            docker exec $POSTGRESQL_CONTAINER_NAME psql -U postgres -c "GRANT ALL PRIVILEGES ON SCHEMA public TO $SONAR_DB_USER;"
        else
            echo "PostgreSQL container not running. Please start it using 'psqladm setup'"
            return 1
        fi

        # Start the SonarQube container
        docker run -d --name $SONARQUBE_CONTAINER_NAME \
            --network $SERVICES_NETWORK \
            -v sonarqube_data:/opt/sonarqube/data \
            -v sonarqube_extensions:/opt/sonarqube/extensions \
            -v sonarqube_logs:/opt/sonarqube/logs \
            -e SONAR_JDBC_URL=jdbc:postgresql://$POSTGRESQL_CONTAINER_NAME/sonarqube \
            -e SONAR_JDBC_USERNAME=$(cat $POSTGRESQL/.credentials/sonarqube.user) \
            -e SONAR_JDBC_PASSWORD=$(cat $POSTGRESQL/.credentials/sonarqube.pass) \
            -e SONAR_WEB_CONTEXT=/sonarqube \
            -e SONAR_ES_CONNECTION_TIMEOUT=1000 \
            $IMAGE_NAME
    }

    # Stop SonarQube container
    stop() {
        docker stop $SONARQUBE_CONTAINER_NAME && \
        echo "SonarQube container stopped."
    }

    # Resume SonarQube container
    resume() {
        docker start $SONARQUBE_CONTAINER_NAME && \
        echo "SonarQube container resumed."
    }

    # Run a SonarQube scan on a project
    scan() {
        sudo docker run --rm -v "$1:/usr/src" -v /etc/pki/ca-trust/source/anchors:/tmp/cacerts -v /etc/hosts:/etc/hosts sonarsource/sonar-scanner-cli -X
     }

    # Backup SonarQube data
    backup() {
        if [ ! -d "$BACKUP_DIR" ]; then
            mkdir -p "$BACKUP_DIR"
        fi
        docker exec $SONARQUBE_CONTAINER_NAME tar -czf /opt/sonarqube/backup/sonarqube_backup.tar.gz /opt/sonarqube/data
        docker cp $SONARQUBE_CONTAINER_NAME:/opt/sonarqube/backup/sonarqube_backup.tar.gz $BACKUP_DIR
        psqladm backup
        echo "SonarQube data backed up to $BACKUP_DIR/sonarqube_backup.tar.gz"
    }

    # Remove SonarQube container, volume and network
    cleanup() {
        docker rm -f $SONARQUBE_CONTAINER_NAME
        docker volume rm sonarqube_data sonarqube_extensions sonarqube_logs
        echo "SonarQube container & volumes removed."
        if [ -d "$BACKUP_DIR" ]; then
            echo "Keeping backup data in $BACKUP_DIR"
        fi
    }

    # Check SonarQube container status
    status() {
        if docker ps --filter "name=$SONARQUBE_CONTAINER_NAME" --filter "status=running" | grep -q $SONARQUBE_CONTAINER_NAME; then
            echo "Healthy"
            return 0
        else
            echo "Stopped"
            return 1
        fi
    }

    # Print the nginx configuration for SonarQube
    nginxconf() {
        cat <<EOF
        # SonarQube
        location /sonarqube {
            proxy_pass http://$SONARQUBE_CONTAINER_NAME:9000;
            proxy_set_header Host \$host;
            proxy_set_header X-Real-IP \$remote_addr;
            proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto \$scheme;
        }
EOF
    }

    case "$cmd" in
        setup)
            setup
            ;;
	    scan)
	        scan "$@"
	    ;;
        stop)
            stop
            ;;
        resume)
            resume
            ;;
        backup)
            backup
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
        *)
            echo "Usage: sonaradm {setup|scan|backup|stop|resume|cleanup|status|nginxconf}"
            return 1
            ;;
    esac
}

# Set up autocompletion for the sonaradm command
_sonaradm_autocomplete() {
    compadd setup scan backup stop resume cleanup status
}

compdef _sonaradm_autocomplete sonaradm
