package config

import (
	"fmt"
	"os"

	"github.com/boxboxjason/svcadm/pkg/containerutils"
	"gopkg.in/yaml.v2"
)

// ParseConfiguration parses the configuration file at the specified path
func ParseConfiguration(config_path string) (Configuration, string, error) {
	var config Configuration

	// Check if the file exists at specified path
	config_file, err := os.Open(config_path)
	if err != nil {
		return Configuration{}, "", err
	}
	defer func(config_file *os.File) {
		err := config_file.Close()
		if err != nil {
			panic(err)
		}
	}(config_file)
	// Read the file content
	file_content, err := os.ReadFile(config_path)
	if err != nil {
		return Configuration{}, "", err
	}
	// Unmarshal the file content into the config map
	err = yaml.Unmarshal(file_content, &config)
	if err != nil {
		return Configuration{}, "", err
	}

	errs := validateAccessContent(&config.General.Access)
	if len(errs) > 0 {
		return Configuration{}, "", fmt.Errorf("errors found in configuration file:\n%s", errs)
	}

	return config, config.General.Access.Logins, nil
}

// ParseUsers parses the users file at the specified path
func ParseUsers(users_path string) (Users, error) {
	var users Users

	// Check if the file exists at specified path
	users_file, err := os.Open(users_path)
	if err != nil {
		return Users{}, err
	}
	defer func(users_file *os.File) {
		err := users_file.Close()
		if err != nil {
			panic(err)
		}
	}(users_file)
	// Read the file content
	file_content, err := os.ReadFile(users_path)
	if err != nil {
		return Users{}, err
	}
	// Unmarshal the file content into the users map
	err = yaml.Unmarshal(file_content, &users)
	if err != nil {
		return Users{}, err
	}

	return users, nil
}

// GetServiceContainerName returns the container name of a service
func GetServiceContainerName(service_name string) string {
	config := GetConfiguration()
	for _, service := range config.Services {
		if service.Name == service_name {
			return service.Container.Name
		}
	}
	return ""
}

// GetService returns the service object from the configuration
func GetService(service_name string) Service {
	config := GetConfiguration()
	for _, service := range config.Services {
		if service.Name == service_name {
			return service
		}
	}
	return Service{}
}

func GetServiceVolumes(service *Service) map[string]string {
	volumes := make(map[string]string)
	for volume, path := range service.Persistence.Volumes {
		if containerutils.VALID_VOLUME_NAME.MatchString(volume) {
			volumes[volume] = path
		}
	}
	return volumes
}
