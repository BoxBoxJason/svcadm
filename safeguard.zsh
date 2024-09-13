#!/bin/zsh

# Safeguard script to monitor and quarantine files for malware
#
# This script monitors the downloads directory and moves any new files to a quarantine directory if they are infected with malware.
# It uses the clamscan command to scan the files and the inotifywait command to monitor the downloads directory for new files.
# If a new file is detected, it is scanned with clamscan, and if it is found to be infected, it is moved to the quarantine directory.
# A log file is maintained to keep track of the scanning and quarantine activities, and a notification is sent to the user if a file is moved to quarantine.
#
# Usage: safeguard {start|clear|status|directory <directories>}
# Arguments:
#   start: Start the safeguarding process
#   clear: Stop the safeguarding process and remove all files created by the safeguard command
#   status: Check the status of the safeguarding process
#   directory: Safeguard a specific directory
#
# Dependencies: clamscan, inotifywait
#
# Author: BoxBoxJason

safeguard() {
    # Check that libnotify is installed
    command -v notify-send >/dev/null 2>&1 || { echo >&2 "notify-send not found, please install it."; exit 1; }
    # Check that clamscan is installed
    command -v clamscan >/dev/null 2>&1 || { echo >&2 "clamscan not found, please install it."; exit 1; }

    local cmd=$1
    shift

    local WATCH_DIRS=("$HOME/Downloads" "$HOME/Documents" "$HOME/Music" "$HOME/Pictures" "$HOME/Videos")
    local LOG_DIR="/var/log/scan"
    local LOG_FILE="$LOG_DIR/safeguard.log"
    local QUARANTINE_DIR="$HOME/.quarantine"

    # Start the monitoring and securing all of the directories
    start() {
        # Update the clamav database
        freshclam
        if [ $? -ne 0 ]; then
            echo "Could not run freshclam, please try using 'sudo freshclam'"
        fi
        # Create the quarantine directory if it does not exist
        mkdir -p "$QUARANTINE_DIR"
        # Create the log directory if it does not exist and check write permissions
        if [ ! -d "$LOG_DIR" ]; then
            mkdir -p "$LOG_DIR"
            if [ "$?" -ne 0 ]; then
                echo "Error creating $LOG_DIR directory, create it manually and grant $(whoami) write permissions"
                return 13
            elif [ ! -w "$LOG_FILE" ]; then
                echo "Error: $(whoami) does not have write permissions for $LOG_FILE"
                return 13
            fi
        fi
        # Check if safeguard is already running (check if inotifywait is monitoring any of the directories)
        if pgrep -x "inotifywait" >/dev/null; then
            INOTIFY_PROCESSES=$(ps -eo pid,cmd | grep '[i]notifywait')
            for WATCH_DIR in "${WATCH_DIRS[@]}"; do
                if echo "$INOTIFY_PROCESSES" | grep -q "$WATCH_DIR"; then
                    echo "Safeguard is already running for $WATCH_DIR"
                    return 1
                fi
            done
        fi
        # Start safeguarding
        echo "$(date +"%Y-%m-%d %H:%M:%S"): Safeguarding starting" | tee -a "$LOG_FILE"
        safeguard directory "${WATCH_DIRS[@]}"
    }

    # Stop the safeguarding & Remove all files created by the safeguard command
    clear() {
        if pgrep -x "inotifywait" >/dev/null; then
            INOTIFY_PROCESSES=$(ps -eo pid,cmd | grep '[i]notifywait')
            for WATCH_DIR in "${WATCH_DIRS[@]}"; do
                if echo "$INOTIFY_PROCESSES" | grep -q "$WATCH_DIR"; then
                    INOTIFY_PID=$(echo "$INOTIFY_PROCESSES" | grep "$WATCH_DIR" | awk '{print $1}')
                    kill "$INOTIFY_PID"
                    echo "$(date +"%Y-%m-%d %H:%M:%S"): Safeguarding stopped for $WATCH_DIR" | tee -a "$LOG_FILE"
                fi
            done
            echo "$(date +"%Y-%m-%d %H:%M:%S"): Safeguarding stopped" | tee -a "$LOG_FILE"
            rm -rf "$QUARANTINE_DIR"
            echo "$(date +"%Y-%m-%d %H:%M:%S"): Quarantine directory removed" | tee -a "$LOG_FILE"
        else
            echo "Safeguard is not running"
            return 3
        fi
    }

    # Check the status of the safeguard command
    status() {
        if pgrep -x "inotifywait" >/dev/null; then
            echo "Safeguard is running and monitoring the following directories:"
            ps -eo pid,cmd | grep '[i]notifywait' | awk '{print $8}'
            return 0
        else
            echo "Safeguard is not running"
            return 3
        fi
    }

    # Safeguard a specific directory
    directory() {
        if [ -z "$1" ]; then
            echo "Usage: safeguard directory <directories (space separated)>"
            return 2
        else
            if [ ! -d "$1" ]; then
                echo "Directory $1 does not exist"
                return 2
            fi
            inotifywait -m -e create --format '%w%f' "$1" | while read NEWFILE
            do
                NOTICE_TIME=$(date +"%Y-%m-%d %H:%M:%S")
                clamscan --move="$QUARANTINE_DIR" "$NEWFILE"
                if [ $? -eq 1 ]; then
                    MESSAGE="Warning: $NEWFILE might be infected and has been moved to $QUARANTINE_DIR !"
                    echo "$NOTICE_TIME: $MESSAGE" | tee -a "$LOG_FILE"
                    notify-send "MALAWARE SUSPICION ALERT" "$MESSAGE" -u critical -i dialog-warning -t 30000 -a "ClamAV" -c "important"
                else
                    echo "$NOTICE_TIME: $NEWFILE scan OK" | tee -a "$LOG_FILE"
                fi
            done &
            if [ "$?" -eq 0 ]; then
                echo "$(date +"%Y-%m-%d %H:%M:%S"): Safeguarding $1" | tee -a "$LOG_FILE"
            fi
            if [ -n "$2" ]; then
                shift
                safeguard directory "$@"
            fi
        fi
    }

    case $cmd in
        start)
            start
            ;;
        clear)
            clear
            ;;
        status)
            status
            ;;
        directory)
            directory "$@"
            ;;
        *)
            echo "Usage: safeguard {start|clear|status|directory <directories>}"
            return 2
            ;;
    esac
}

# Set up autocompletion for the safeguard command
_safeguard_autocomplete() {
    compadd start clear status directory
}

compdef _safeguard_autocomplete safeguard
