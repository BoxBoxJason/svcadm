# svcadm

svcadm is a command-line utility for managing services on unix and MacOS systems. It is a wrapper based on zsh script that you can use to setup various development environments services. Said services include:
- databases (postgresql)
- code quality tools (sonarqube)
- storage (minio)
- web servers (nginx)
- integrity scan (trivy, clamav)
- and more to come...

## Table of Contents
- [Installation](#installation)
- [Usage](#usage)
  - [psqladm](#psqladm)
  - [nginxadm](#nginxadm)
  - [sonaradm](#sonarqubeadm)
  - [safeguard](#safeguard)
  - [minioadm](#minioadm)
- [Customization](#customization)
- [License](#license)

## Installation

1. Make sure you have `zsh` installed on your system. If not, you can install it using the package manager of your choice.
2. Clone the repository to your local machine `git clone https://github.com/boxboxjason/svcadm.git`
3. Navigate to the `svcadm` directory `cd svcadm`
4. Source all of the scripts in your `.zshrc` file by adding `source /path/to/svcadm/*.zsh` to your `.zshrc` file. If you use `oh-my-zsh`, you can simply move the scripts to the `.oh-ùy-zsh/custom` directory.

## Usage

### psqladm

**psqladm** is a wrapper script designed to manage a PostgreSQL container. It provides various commands to set up, manage, backup, and restore PostgreSQL databases using Docker containers. The script allows you to automate common PostgreSQL operations and manage containers more effectively.

#### Commands
- `setup`: Sets up a PostgreSQL container with persistent storage and starts it. This command initializes the required Docker volumes and network settings.
- `add_database`: Adds a new database to the running PostgreSQL container. This command is useful for initializing new projects or environments.
- `backup`: Backs up a specific database from the PostgreSQL container to a designated backup directory. The backup files are stored in the `$POSTGRESQL/backup` directory by default.
- `stop`: Stops the running PostgreSQL container without removing any data or configuration. This command is useful for temporarily halting operations.
- `resume`: Resumes the PostgreSQL container from a stopped state, allowing you to continue operations without loss of data.
- `cleanup`: Removes the PostgreSQL container and all associated Docker volumes and network settings. This is a destructive action and will delete all stored data.
- `status`: Checks and displays the status of the PostgreSQL container, indicating whether it is running, stopped, or not present.

#### Environment Variables
To use the `psqladm` script, several environment variables need to be configured:
- `POSTGRESQL`: This environment variable should point to the path where PostgreSQL-related data and configurations are stored. This path is used to store backups and credential files.
- `POSTGRESQL_CONTAINER_NAME`: Specifies the name of the PostgreSQL Docker container. This name is used to identify and manage the container.
- `SERVICES_NETWORK`: Defines the Docker network in which the PostgreSQL container will be placed. This network should be preconfigured or set up during the `setup` command.

#### Customization
- **Version**: The default version of the PostgreSQL image used in the container is `postgres:15`. To change the version, modify the line `IMAGE_NAME="postgres:15"` in the script.
- **Backup Directory**: The default backup directory is set to `$POSTGRESQL/backup`. To change this, edit the `BACKUP_DIR="$POSTGRESQL/backup"` line in the script.
- **Credentials**: The default superuser for the PostgreSQL container is `postgres`, and its password is generated randomly during setup. This password is stored in a file within the credentials directory, which is defined by `CREDENTIALS_DIR="$POSTGRESQL/.credentials"`. To change the path where credentials are stored, update this line in the script.

#### Dependencies
- **Docker**: Ensure Docker is installed and running on your machine. The script uses Docker commands to manage the PostgreSQL container and related resources.

#### Example Usage
```bash
# Set up and start the PostgreSQL container
psqladm setup

# Add a new database named 'exampledb' with owner 'myuser' and password (OPTIONAL, will be generated randomly if not specify) 'mypassword'
psqladm add_database exampledb myuser mypassword

# Backup the 'exampledb' database
./postgresql.zsh backup exampledb

# Stop the PostgreSQL container
./postgresql.zsh stop

# Resume the PostgreSQL container
./postgresql.zsh resume

# Check the status of the PostgreSQL container
./postgresql.zsh status

# Clean up the PostgreSQL container and all resources
./postgresql.zsh cleanup
```

### sonaradm

**sonaradm** is a wrapper script designed to manage a SonarQube container, which is used for continuous inspection of code quality and security. The script automates various tasks, such as setting up the SonarQube environment, running code scans, managing backups, and handling container lifecycle operations.

#### Commands
- `setup`: Sets up a SonarQube container along with its required volumes and PostgreSQL database, and starts the container. This command initializes the necessary Docker resources and configures the SonarQube environment.
- `scan`: Runs a SonarQube scan on a specified project. This command uses `curl` to interact with the SonarQube server and execute a scan on the provided project key.
- `backup`: Backs up SonarQube data, including project configurations and analysis data, to a specified backup directory. The backup files are stored in the `$SONARQUBE/backup` directory by default.
- `stop`: Stops the running SonarQube container without removing any data or configurations. This is useful for temporarily stopping SonarQube services.
- `resume`: Resumes the SonarQube container from a stopped state, allowing you to continue from where you left off without data loss.
- `cleanup`: Removes the SonarQube container, associated Docker volumes, and network settings. This is a destructive action that deletes all stored data and configurations.
- `status`: Checks and displays the status of the SonarQube container, indicating whether it is running, stopped, or not present.
- `nginxconf`: Generates an Nginx configuration for reverse proxying SonarQube.

#### Environment Variables

To use the `sonaradm` script, several environment variables must be configured:

- `SONARQUBE`: This environment variable should point to the path where SonarQube-related data and configurations are stored. This path is used for backups and other operations.
- `SONARQUBE_CONTAINER_NAME`: Specifies the name of the SonarQube Docker container. This name is used to identify and manage the container.
- `POSTGRESQL_CONTAINER_NAME`: Specifies the name of the PostgreSQL container used by SonarQube for storing analysis data.
- `SERVICES_NETWORK`: Defines the Docker network in which the SonarQube container will be placed. This network should be preconfigured or set up during the `setup` command.

#### Customization
- **SonarQube Version**: To change the SonarQube version or image, edit the respective `IMAGE_NAME` variable in the script.
- **Backup Directory**: The default backup directory is set to `$SONARQUBE/backup`. To change the backup location, modify the `BACKUP_DIR="$SONARQUBE/backup"` line in the script.

#### Dependencies
- **jq**: A lightweight and flexible command-line JSON processor. Ensure `jq` is installed on your machine.
- **curl**: A command-line tool for transferring data with URLs. Used for interacting with the SonarQube server.
- **Docker**: Ensure Docker is installed and running on your machine. The script uses Docker commands to manage the SonarQube container and related resources.

#### Example Usage
```bash
# Set up and start the SonarQube container
sonardm setup

# Run a SonarQube scan on the project with key 'myproject'
sonardm scan myproject

# Backup SonarQube data
sonardm backup

# Stop the SonarQube container
sonardm stop

# Resume the SonarQube container
sonardm resume

# Check the status of the SonarQube container
sonardm status

# Clean up the SonarQube container and all resources
sonardm cleanup
```

### nginxadm

**nginxadm** is a wrapper script designed to manage an Nginx container, which is commonly used as a web server or reverse proxy. The script provides commands to set up the Nginx environment, manage the container lifecycle, and monitor its status.

#### Commands
- `setup`: Initializes and starts the Nginx container with the specified configurations. This command creates the required Docker volumes, network settings, and initializes the Nginx environment. It sets up the Nginx container to serve web traffic with the specified configurations.
- `stop`: Stops the running Nginx container without removing any data or configurations. This is useful for temporarily stopping the Nginx service.
- `resume`: Resumes the Nginx container from a stopped state, allowing you to continue serving web traffic without data loss.
- `cleanup`: Removes the Nginx container and all associated Docker volumes and network settings. This action is irreversible and will delete all stored configurations.
- `status`: Checks and displays the status of the Nginx container, indicating whether it is running, stopped, or not present.

#### Environment Variables

To use the `nginxadm` script, several environment variables need to be configured:

- `NGINX`: This environment variable should point to the path where Nginx-related data and configurations are stored. This path is used for storing Nginx configuration files and logs.
- `NGINX_CONTAINER_NAME`: Specifies the name of the Nginx Docker container. This name is used to identify and manage the container.
- `SONARQUBE_CONTAINER_NAME`: Specifies the name of the SonarQube container (if applicable for reverse proxy configurations).
- `SERVICES_NETWORK`: Defines the Docker network in which the Nginx container will be placed. This network should be preconfigured or set up during the `setup` command.

#### Customization
- **Nginx Version**: To change the Nginx version or image, edit the respective `IMAGE_NAME` variable in the script.
- **Configuration Directory**: The default configuration and log directory is set to `$NGINX/conf` and `$NGINX/logs`, respectively. To change these, modify the respective lines in the script to point to your preferred directories.

#### Dependencies
- **Docker**: Ensure Docker is installed and running on your machine. The script uses Docker commands to manage the Nginx container and related resources.

#### Example Usage

```bash
# Set up and start the Nginx container
./nginx.zsh setup

# Stop the Nginx container
./nginx.zsh stop

# Resume the Nginx container
./nginx.zsh resume

# Check the status of the Nginx container
./nginx.zsh status

# Clean up the Nginx container and all resources
./nginx.zsh cleanup
```

### safeguard

**safeguard** is a utility script designed to continuously monitor directories for malware and automatically quarantine any suspicious files. It uses the `clamscan` command to perform malware scanning and `inotifywait` to monitor directories for changes. Any new files added to the monitored directory are scanned for malware, and if any threats are detected, the files are moved to a quarantine directory for further analysis.

#### Setup
If you plan on using the `safeguard` script without customizing it, ensure that the following steps are completed:
1. Install the `clamav` package on your system to provide the `clamscan` command.
2. Install the `inotify-tools` package on your system to provide the `inotifywait` command.
3. Set up the logrotation for the safeguard log file to prevent it from growing indefinitely.
  You can create a logrotate configuration file for the safeguard log file by creating a new file in `/etc/logrotate.d/safeguard` with the following content:
```
/var/log/scan/safeguard.log {
    monthly
    missingok
    rotate 12
    compress
    notifempty
    create 640 YOUR_USER YOUR_GROUP
}
```
4. You can add the script to your cron jobs to start the process at boot time, ensuring safeguarding of your directories as long as your machine is on. To do this, add `@reboot /usr/bin/zsh -c "source /home/YOURUSER/.zshrc && safeguard start"` to your crontab
  
#### Commands
- `start`: Starts the safeguarding process. This command initiates monitoring of the specified directory (or directories) for any new files and scans them for malware. If a file is detected as infected, it is moved to a quarantine directory.
- `clear`: Stops the safeguarding process and removes all files created by the safeguard command, including logs and quarantined files. This command is useful for resetting the safeguard environment.
- `status`: Checks and displays the current status of the safeguarding process, indicating whether monitoring is active or stopped.
- `directory <directories>`: Allows the user to specify a custom directory (or multiple directories) to safeguard. Note that the monitoring stops at reboot and you will need to start the safeguarding process again.

#### Customization
- **Quarantine Directory**: To change the default directory where infected files are quarantined, modify the `QUARANTINE_DIR="$HOME/.quarantine"` line in the script.
- **Monitoring Directory**: To change the default directory being monitored, modify the `WATCH_DIRS=("$HOME/Downloads" "$HOME/Documents" "$HOME/Music" "$HOME/Pictures" "$HOME/Videos")"` line.
- **Log Directory**: The default log directory is set to `/var/log/scan`. To change this, edit the `LOG_DIR="/var/log/scan"` line in the script.
- **Log File**: The default log file is set to `safeguard.log`. To change this, modify the `LOG_FILE="safeguard.log"` line in the script.

#### Dependencies
- **clamscan**: A command-line antivirus scanner that is part of the ClamAV suite. Ensure `clamscan` is installed and properly configured on your system.
- **inotifywait**: A command-line utility that watches for changes to files and directories using Linux's inotify interface. Ensure `inotifywait` is installed on your system.

#### Example Usage

```bash
# Start the safeguarding process to monitor the default downloads directory
.safeguard start

# Specify a custom directory to monitor for malware
safeguard directory /path/to/custom/directory

# Check the status of the safeguarding process
safeguard status

# Clear all quarantine files and stop the safeguarding process
safeguard clear
```

### minioadm

**minioadm** is a wrapper script designed to manage a MinIO container, which is an object storage service compatible with Amazon S3. The script automates various tasks such as setting up the MinIO environment, managing storage buckets, performing backups, and handling the container lifecycle.

#### Commands
- **`setup`**: Initializes and starts the MinIO container with persistent storage. It sets up the necessary Docker volumes and network configurations.
- **`resume`**: Resumes a stopped MinIO container to continue operations without data loss.
- **`add_bucket`**: Adds a new bucket to the running MinIO container. Useful for organizing storage for different projects or environments.
- **`remove_bucket`**: Removes an existing bucket from the MinIO container.
- **`backup`**: Backs up a specific bucket from the MinIO container to a designated backup directory.
- **`stop`**: Stops the running MinIO container without removing any data or configuration. This command is useful for temporarily halting operations.
- **`cleanup`**: Removes the MinIO container and all associated Docker volumes and network settings. This action is irreversible and will delete all stored data.
- **`status`**: Checks and displays the status of the MinIO container, indicating whether it is running, stopped, or not present.
- **`nginxconf`**: Generates and prints the Nginx configuration for reverse proxying the MinIO service.

#### Environment Variables

To use the `minioadm` script, several environment variables need to be configured:

- **`MINIO`**: This environment variable should point to the path where MinIO-related data and configurations are stored. This path is used for storing configurations and data.
- **`MINIO_CONTAINER_NAME`**: Specifies the name of the MinIO Docker container. This name is used to identify and manage the container.
- **`SERVICES_NETWORK`**: Defines the Docker network in which the MinIO container will be placed. This network should be preconfigured or set up during the `setup` command.

#### Customization
- **MinIO Version**: To change the MinIO version or image, edit the respective `IMAGE_NAME` variable in the script.
- **Bucket Management**: The default command creates a bucket using the MinIO client. To modify bucket management behavior, adjust the corresponding functions in the script.

#### Dependencies
- **Docker**: Ensure Docker is installed and running on your machine. The script uses Docker commands to manage the MinIO container and related resources.

#### Example Usage
```bash
# Set up and start the MinIO container
minioadm setup

# Add a new bucket named 'mybucket' to the MinIO container
minioadm add_bucket mybucket

# Backup the 'mybucket' bucket
minioadm backup mybucket

# Stop the MinIO container
minioadm stop

# Resume the MinIO container
minioadm resume

# Check the status of the MinIO container
minioadm status

# Remove a bucket named 'mybucket' from the MinIO container
minioadm remove_bucket mybucket

# Clean up the MinIO container and all resources
minioadm cleanup
```

## License
This Software is under the Unlicense License. Meaning you can do whatever you want with it.  
See the [LICENSE](LICENSE) file for details.
