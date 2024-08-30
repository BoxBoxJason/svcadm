#!/bin/zsh

# PostgreSQL service management script wrapper
# Usage: psqladm {setup|add_database|backup|stop|resume|cleanup|status}
# Arguments:
#   setup: Set up PostgreSQL container & volume, start the container
#   add_database: Add a database to the PostgreSQL container
#   backup: Backup a database from the PostgreSQL container
#   stop: Stop PostgreSQL container
#   resume: Resume PostgreSQL container
#   cleanup: Remove PostgreSQL container & volume
#   status: Check PostgreSQL container status
#
# Environment variables:
#   POSTGRESQL: Path to the PostgreSQL service
#   POSTGRESQL_CONTAINER_NAME: Name of the PostgreSQL container
#   SERVICES_NETWORK: Name of the services network
#
# Dependencies: docker
#
# Author: BoxBoxJason

psqladm() {
    # Check if required environment variables are set
    if [ -z "$POSTGRESQL" ]; then
        echo "POSTGRESQL environment variable not set. Please set it to the path of the PostgreSQL service"
        return 2
    elif [ -z "$POSTGRESQL_CONTAINER_NAME" ]; then
        echo "POSTGRESQL_CONTAINER_NAME environment variable not set. Please set it to the name of the PostgreSQL container"
        return 2
    elif [ -z "$SERVICES_NETWORK" ]; then
        echo "SERVICES_NETWORK environment variable not set. Please set it to the name of the PostgreSQL network"
        return 2
    elif ! docker network ls | grep -q $SERVICES_NETWORK; then
        echo "Services network not found. Please create it using 'docker network create $SERVICES_NETWORK'"
        return 2
    fi

    local cmd=$1
    shift

    local VOLUME_NAME="postgresql_data"
    local IMAGE_NAME="postgres:15"
    local BACKUP_DIR="$POSTGRESQL/backup"
    local CREDENTIALS_DIR="$POSTGRESQL/.credentials"

    # Set up PostgreSQL container, volume and start the container
    setup() {
        if docker ps --filter "name=$POSTGRESQL_CONTAINER_NAME" --filter "status=running" | grep -q $POSTGRESQL_CONTAINER_NAME; then
            echo "PostgreSQL container is already running"
            return 1
        fi

        mkdir -p "$CREDENTIALS_DIR"
        echo "Creating Docker network and volume for PostgreSQL"
        docker volume create $VOLUME_NAME
        if [ ! -f "$CREDENTIALS_DIR/.pgpass" ]; then
            echo "Generating random password for postgres user"
            generate_password > "$CREDENTIALS_DIR/.pgpass"
        fi

        docker run -d --name $POSTGRESQL_CONTAINER_NAME \
        --network $SERVICES_NETWORK \
        -v "$VOLUME_NAME:/var/lib/postgresql/data" \
        -e "POSTGRES_PASSWORD=$(cat "$CREDENTIALS_DIR/.pgpass")" \
        $IMAGE_NAME
    }

    # Add a database to the PostgreSQL container
    add_database() {
        # Check if container is running
        if docker ps --filter "name=$POSTGRESQL_CONTAINER_NAME" --filter "status=running" | grep -q $POSTGRESQL_CONTAINER_NAME; then
            echo "Adding database $1 with owner $2 to PostgreSQL container"
            local PASSWORD="$(generate_password)"
            # Check if database and owner are provided
            if [ -z "$1" ] || [ -z "$2" ]; then
                echo "Usage: postgresql add_database <database> <owner> <password (optional)>"
                return 2
            # Check if password is provided
            elif [ -n "$3" ]; then
                PASSWORD="$3"
            fi

            # Check if credentials directory exists
            mkdir -p "$CREDENTIALS_DIR"
            # Check if database already exists
            if [[ -f "$CREDENTIALS_DIR/$1.user" ]] || [[ -f "$CREDENTIALS_DIR/$1.pass" ]]; then
                echo "Database $1 already exists"
                return 1
            fi
            # Save credentials to files
            echo "$2" > "$CREDENTIALS_DIR/$1.user"
            echo "$PASSWORD" > "$CREDENTIALS_DIR/$1.pass"

            # Create the database and set the owner
            docker exec $POSTGRESQL_CONTAINER_NAME psql -U postgres -c "CREATE DATABASE $1;"
            docker exec $POSTGRESQL_CONTAINER_NAME psql -U postgres -c "CREATE USER $2 WITH PASSWORD '$PASSWORD';"
            docker exec $POSTGRESQL_CONTAINER_NAME psql -U postgres -c "ALTER DATABASE $1 OWNER TO $2;"
            docker exec $POSTGRESQL_CONTAINER_NAME psql -U postgres -c "GRANT ALL PRIVILEGES ON DATABASE $1 TO $2;"
        else
            echo "PostgreSQL container is not running, please use 'postgresql start' to start the container"
            return 1
        fi
    }

    # Stop PostgreSQL container
    stop() {
        docker stop $POSTGRESQL_CONTAINER_NAME && \
        echo "PostgreSQL container stopped."
    }

    # Resume PostgreSQL container
    resume() {
        echo "Resuming PostgreSQL container"
        docker start $POSTGRESQL_CONTAINER_NAME
    }

    # Remove PostgreSQL container and volume
    cleanup() {
        docker rm -f $POSTGRESQL_CONTAINER_NAME
        docker volume rm $VOLUME_NAME
        rm -rf $CREDENTIALS_DIR
        if [ -d "$BACKUP_DIR" ]; then
            echo "Keeping backup directory $BACKUP_DIR, you may want to remove it manually"
        fi
    }

    # Check PostgreSQL container status
    status() {
        if docker ps --filter "name=$POSTGRESQL_CONTAINER_NAME" --filter "status=running" | grep -q $POSTGRESQL_CONTAINER_NAME; then
            # Check PostgreSQL status inside the container
            if docker exec $POSTGRESQL_CONTAINER_NAME pg_isready -U postgres; then
                echo "Healthy"
            else
                echo "Not healthy"
            fi
        else
            echo "Stopped"
        fi
    }

    # Backup a database from the PostgreSQL container
    backup() {
        if [ -z "$1" ]; then
            echo "Usage: postgresql backup <database>"
            return 2
        fi
        mkdir -p "$BACKUP_DIR"
        if [ "$1" = "*" ]; then
            echo "Backing up all databases from PostgreSQL container"
            local DATABASES=$(docker exec $POSTGRESQL_CONTAINER_NAME psql -U postgres -t -c "SELECT datname FROM pg_database WHERE datistemplate = false;")
            for db in $DATABASES; do
                psqladm backup $db
            done
        else
            echo "Backing up PostgreSQL database $1 to $BACKUP_DIR/$1.sql"
            docker exec $POSTGRESQL_CONTAINER_NAME pg_dump -U postgres $1 > "$BACKUP_DIR/$1.sql"
        fi
    }

    case "$cmd" in
        setup)
            setup
            ;;
	    add_database)
	        add_database "$@"
	        ;;
        backup)
            backup "$@"
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
            echo "Usage: psqladm {setup|add_database|backup|stop|resume|cleanup|status}"
            return 1
            ;;
    esac
}

# Set up autocompletion for the psqladm command
_psqladm_autocomplete() {
    compadd setup add_database backup stop resume cleanup status
}

compdef _psqladm_autocomplete psqladm
