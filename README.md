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
  - [safeguard](#safeguard)
- [Customization](#customization)
- [License](#license)

## Installation

1. Make sure you have `zsh` installed on your system. If not, you can install it using the package manager of your choice.
2. Clone the repository to your local machine `git clone https://github.com/boxboxjason/svcadm.git`
3. Navigate to the `svcadm` directory `cd svcadm`
4. Source all of the scripts in your `.zshrc` file by adding `source /path/to/svcadm/*.zsh` to your `.zshrc` file. If you use `oh-my-zsh`, you can simply move the scripts to the `.oh-my-zsh/custom` directory.

## Usage

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
