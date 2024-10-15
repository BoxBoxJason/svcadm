package config

import (
	"fmt"
	"os"
	"regexp"

	"github.com/boxboxjason/svcadm/pkg/logger"
)

const (
	ALPHANUMERICS_REGEX = `^[a-zA-Z0-9_-]+$`
)

var (
	VALID_VOLUME_NAME    = regexp.MustCompile(ALPHANUMERICS_REGEX)
	VALID_CONTAINER_NAME = regexp.MustCompile(ALPHANUMERICS_REGEX)
	VALID_NETWORK_NAME   = regexp.MustCompile(ALPHANUMERICS_REGEX)
	VALID_USERNAME       = regexp.MustCompile(`^[a-zA-Z0-9_-]{3,20}$`)
	VALID_RESTART_POLICY = regexp.MustCompile(`^(always|no|unless-stopped|on-failure:\d+)$`)
)

// ValidateConfiguration ensures that the services in the configuration file are valid
func ValidateConfiguration() {
	config := GetConfiguration()
	valid_services := retrieveValidServices()

	all_errors := make([]string, 0)

	// Check if the container operator content is valid
	all_errors = append(all_errors, validateContainerOperatorContent(&config.General.ContainerOperator)...)

	// Check if the services are valid
	for _, service := range config.Services {
		logger.Debug(fmt.Sprintf("validating service: %s", service.Name))
		if _, ok := valid_services[service.Name]; !ok {
			all_errors = append(all_errors, fmt.Sprintf("invalid service found: %s", service.Name))
		} else {
			all_errors = append(all_errors, validateContainerContent(&service.Container)...)
			all_errors = append(all_errors, validatePersistenceContent(&service.Persistence)...)
			all_errors = append(all_errors, validateBackupContent(&service.Backup)...)
		}
	}

	if len(all_errors) > 0 {
		logger.Fatal(fmt.Sprintf("errors found in configuration file:\n%s", all_errors))
	}
	logger.Debug("configuration is valid")
}

// validateContainerContent ensures that the container content in the configuration file is valid
func validateContainerContent(container *Container) []string {
	errors := make([]string, 0)

	if container == nil {
		return []string{"container content is empty"}
	}

	// Check if the restart policy is valid
	if !VALID_RESTART_POLICY.MatchString(container.Restart) {
		errors = append(errors, fmt.Sprintf("invalid restart policy for container %s: %s", container.Name, container.Restart))
	}

	// Check if the container name is valid
	if !VALID_CONTAINER_NAME.MatchString(container.Name) {
		errors = append(errors, fmt.Sprintf("invalid container name: %s", container.Name))
	}

	// Check if the ports are valid (if specified), are positive integers, lower than 65536
	if len(container.Ports) > 0 {
		for _, port := range container.Ports {
			if port < 0 || port > 65535 {
				errors = append(errors, fmt.Sprintf("invalid port for container %s: %d", container.Name, port))
			}
			if container.Ports[port] < 0 || container.Ports[port] > 65535 {
				errors = append(errors, fmt.Sprintf("invalid port mapping for container %s: %d", container.Name, container.Ports[port]))
			}
		}

	}

	return errors
}

// validatePersistenceContent ensures that the persistence content in the configuration file is valid (if enabled)
func validatePersistenceContent(persistence *Persistence) []string {
	errors := make([]string, 0)

	if persistence == nil || !persistence.Enabled {
		return errors
	}

	// Check if the volumes are valid
	for volume := range persistence.Volumes {
		if !VALID_VOLUME_NAME.MatchString(volume) {
			if _, err := os.Stat(volume); err != nil {
				errors = append(errors, fmt.Sprintf("invalid volume name: %s", volume))
			}
		}
	}

	return errors
}

// validateBackupContent ensures that the backup content in the configuration file is valid (if enabled)
func validateBackupContent(backup *Backup) []string {
	errors := make([]string, 0)

	if !backup.Enabled {
		return errors
	}

	// Check if the backup schedule is valid
	if backup.Frequency != "daily" && backup.Frequency != "weekly" && backup.Frequency != "monthly" {
		errors = append(errors, fmt.Sprintf("invalid backup frequency: %s", backup.Frequency))
	}

	// Check if the backup retention is valid
	if backup.Retention < 0 {
		errors = append(errors, fmt.Sprintf("invalid backup retention: %d", backup.Retention))
	}

	return errors
}

func validateContainerOperatorContent(container_operator *ContainerOperator) []string {
	if container_operator == nil {
		return []string{"container operator content is empty"}
	}

	errors := make([]string, 0)

	// Check if the container operator name is valid
	if container_operator.Name != "docker" && container_operator.Name != "podman" {
		errors = append(errors, fmt.Sprintf("invalid container operator name: %s", container_operator.Name))
	}

	// Check if the services network is valid
	if !VALID_NETWORK_NAME.MatchString(container_operator.Network.Name) {
		errors = append(errors, fmt.Sprintf("invalid network name: %s", container_operator.Network.Name))
	}

	// Check if the services network driver is valid
	if container_operator.Network.Driver != "bridge" && container_operator.Network.Driver != "host" && container_operator.Network.Driver != "none" {
		errors = append(errors, fmt.Sprintf("invalid network driver: %s", container_operator.Network.Driver))
	}

	return errors

}

func validateAccessContent(access *Access) []string {
	if access == nil {
		return []string{"access content is empty"}
	}

	errors := make([]string, 0)

	// Check if the logins file path is valid
	if access.Logins == "" {
		return []string{"logins file path is empty"}
	} else if _, err := os.Stat(access.Logins); err != nil {
		return []string{fmt.Sprintf("error checking logins file path: %s", err)}
	}

	if access.Encryption.Enabled {
		// Check if the encryption key is valid
		if access.Encryption.Key == "" {
			errors = append(errors, "encryption key is empty")
		}
		if access.Encryption.Salt == "" {
			errors = append(errors, "encryption salt is empty")
		}
	}

	return errors
}

// ValidateUsers ensures that the users in the users file are valid
func ValidateUsers() {
	users := GetUsers()

	valid_count := 0

	// Check if the users are valid
	for _, user := range users.Users {
		if !validateUsername(user.Username) {
			fmt.Println("invalid username: " + user.Username)
		}
		if !validatePassword(user.Password) {
			fmt.Println("invalid password for user " + user.Username)
		}
		valid_count++
	}

	// Check if the admin users are valid
	for _, user := range users.Admins {
		if !validateUsername(user.Username) {
			fmt.Println("invalid username: " + user.Username)
		}
		if !validatePassword(user.Password) {
			fmt.Println("invalid password for user " + user.Username)
		}
		valid_count++
	}

	if valid_count == 0 {
		logger.Fatal("no (valid) users found in users file")
	}
	logger.Debug("users are valid")
}

// validateUsername ensures that the username uses alphanumeric characters, underscores and hyphens, the username must be between 3 and 20 characters
func validateUsername(username string) bool {
	return VALID_USERNAME.MatchString(username)
}

// validatePassword ensures that the password is between 6 and 32 characters long, with no spaces at the beginning or end
func validatePassword(password string) bool {
	if len(password) < 6 || len(password) > 32 {
		return false
	}
	return password[0] != ' ' && password[len(password)-1] != ' '
}
