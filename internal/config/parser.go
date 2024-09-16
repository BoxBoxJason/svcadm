package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v2"
)

// ParseConfiguration parses the configuration file at the specified path
func ParseConfiguration(config_path string) (Configuration, Users, error) {
	var config Configuration

	// Check if the file exists at specified path
	config_file, err := os.Open(config_path)
	if err != nil {
		return Configuration{}, Users{}, err
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
		return Configuration{}, Users{}, err
	}
	// Unmarshal the file content into the config map
	err = yaml.Unmarshal(file_content, &config)
	if err != nil {
		return Configuration{}, Users{}, err
	}

	var users Users
	errs := validateAccessContent(&config.General.Access)
	if len(errs) > 0 {
		return Configuration{}, Users{}, fmt.Errorf("errors found in configuration file:\n%s", errs)
	}
	users, err = ParseUsers(config.General.Access.Logins)
	if err != nil {
		return Configuration{}, Users{}, err
	}

	return config, users, nil
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
func GetServiceContainerName(config *Configuration, service_name string) string {
	for _, service := range config.Services {
		if service.Name == service_name {
			return service.Container.Name
		}
	}
	return ""
}

// GetService returns the service object from the configuration
func GetService(config *Configuration, service_name string) Service {
	for _, service := range config.Services {
		if service.Name == service_name {
			return service
		}
	}
	return Service{}
}
