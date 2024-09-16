package config

import (
	"fmt"
	"os"
	"regexp"
)

// CheckConfiguration checks the configuration file at the specified path, adds default values to empty fields
func CheckConfiguration(config_path string) (Configuration, Users, error) {
	config, users, err := ParseConfiguration(config_path)
	if err != nil {
		return Configuration{}, Users{}, err
	}

	// Check if the services in the configuration file are valid
	err = validateServicesNames(&config)
	if err != nil {
		return Configuration{}, Users{}, err
	}

	return config, users, nil
}

// validateServicesNames ensures that the services in the configuration file are valid
func validateServicesNames(config *Configuration) error {
	valid_services := retrieveValidServices()

	all_errors := make([]string, 0)

	// Check if the container operator content is valid
	all_errors = append(all_errors, validateContainerOperatorContent(&config.General.ContainerOperator)...)

	// Check if the services are valid
	for _, service := range config.Services {
		if _, ok := valid_services[service.Name]; !ok {
			all_errors = append(all_errors, fmt.Sprintf("invalid service found: %s", service.Name))
			all_errors = append(all_errors, validateContainerContent(&service.Container)...)
			all_errors = append(all_errors, validatePersistenceContent(&service.Persistence)...)
			all_errors = append(all_errors, validateBackupContent(&service.Backup)...)
		}
	}

	if len(all_errors) > 0 {
		return fmt.Errorf("errors found in configuration file:\n%s", all_errors)
	}
	return nil
}

// validateContainerContent ensures that the container content in the configuration file is valid
func validateContainerContent(container *Container) []string {
	errors := make([]string, 0)

	if container == nil {
		return []string{"container content is empty"}
	}

	// Check if the restart policy is valid
	onfailure_pattern := `^on-failure:\d+$`
	re, err := regexp.Compile(onfailure_pattern)
	if err != nil {
		return []string{fmt.Sprintf("error compiling regex: %s", err)}
	}
	if container.Restart != "always" && container.Restart != "no" && container.Restart != "unless-stopped" && !re.MatchString(container.Restart) {
		errors = append(errors, fmt.Sprintf("invalid restart policy for container %s: %s", container.Name, container.Restart))
	}

	// Check if the container name is valid
	re, err = regexp.Compile(`^[a-zA-Z0-9_-]+$`)
	if err != nil {
		return []string{fmt.Sprintf("error compiling regex: %s", err)}
	}
	if !re.MatchString(container.Name) {
		errors = append(errors, fmt.Sprintf("invalid container name: %s", container.Name))
	}

	// Check if the ports are valid (if specified), must be either in format "port:port" OR "nginx"
	if len(container.Ports) > 0 {
		for _, port := range container.Ports {
			if port != "nginx" {
				re, err = regexp.Compile(`^\d+:\d+$`)
				if err != nil {
					errors = append(errors, fmt.Sprintf("error compiling regex: %s", err))
				}
				if !re.MatchString(port) {
					errors = append(errors, fmt.Sprintf("invalid port format for container %s: %s", container.Name, port))
				}
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
	re, err := regexp.Compile(`^[a-zA-Z0-9_-]+$`)
	if err != nil {
		return []string{fmt.Sprintf("error compiling regex: %s", err)}
	}
	for volume := range persistence.Volumes {
		if !re.MatchString(volume) {
			errors = append(errors, fmt.Sprintf("invalid volume name: %s", volume))
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
	re, err := regexp.Compile(`^[a-zA-Z0-9_-]+$`)
	if err != nil {
		return []string{fmt.Sprintf("error compiling regex: %s", err)}
	}
	if !re.MatchString(container_operator.Network.Name) {
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
