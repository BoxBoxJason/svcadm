# svcadm

svcadm is a command-line utility for deploying and managing development team services on a local machine. It is designed to simplify the process of setting up and managing development environments by providing a simple and intuitive interface for starting, stopping, and monitoring services.
- Source code hosting & CI/CD (GitLab)
- Secret management (Hashicorp Vault)
- Team communication (Mattermost)
- Databases (PostgreSQL)
- Code quality tools (SonarQube)
- Artifact repository (Minio)
- Integrity scans & malware protection (ClamAV, Trivy)
- Reverse proxy (Nginx)
- and more to come...

## Table of Contents
- [Usage](#usage)
  - [svcadm](#svcadm)
  - [safeguard](#safeguard)
- [Customization](#customization)
- [License](#license)

## Usage

### svcadm

**svcadm** is a command-line utility for deploying and managing development team services on a local machine. It provides a simple and intuitive interface for starting, stopping, and monitoring services. The `svcadm` utility is designed to simplify the process of setting up and managing development environments by providing a unified interface for managing various services.

The configuration of svcadm is done through a configuration file (which by default is) located at `~/.svcadm/svcadm.yaml`. This file contains the configuration for the services that can be managed by `svcadm`. The configuration file is in YAML format and contains all the necessary information to manage the services automatically.

#### Setup
1. Download the `svcadm` binary corresponding to your operating system from the [releases page](https://github.com/BoxBoxJason/svcadm/releases)
2. Ensure that the `svcadm` binary is executable by running `chmod +x svcadm`
3. Move the `svcadm` binary to a directory in your `PATH` to make it accessible from anywhere on your system.
4. Check the available commands by running `svcadm --help`

#### Getting Started
1. Generate the default configuration file by running `svcadm config default`, this will download the default configuration file from the internet and save it to `~/.svcadm/svcadm.yaml`
2. Edit the configuration file to match your environment. You can enable or disable services as needed.
3. Start the services by running `svcadm setup`
4. Check the status of the services by running `svcadm status`
5. Stop the services by running `svcadm stop`

#### Commands
- `config <command>`: Manage the configuration file
  - `default`: Download the default configuration file
  - `edit`: Open the configuration file in the default editor
  - `generate`: Generates a new configuration file by prompting the user for the necessary information
  - `show`: Display the contents of the configuration file
  - `validate`: Validate the configuration file for syntax errors and invalid / improperly configured services
- `setup`: Start the services defined in the configuration file
- `status <services>`: Check the status of the services
- `stop <services>`: Stop the services specified in the command
- `logs <services>`: Display the logs for the specified services
- `restart <services>`: Restart the specified services
- `backup <services>`: Backup the data for the specified services
- `cleanup <services>`: Cleanup the data for the specified services

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

## License
This Software is under the Unlicense License. Meaning you can do whatever you want with it.
See the [LICENSE](LICENSE) file for details.
