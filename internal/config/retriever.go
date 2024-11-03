package config

import (
	"github.com/boxboxjason/svcadm/pkg/utils"
	"gopkg.in/yaml.v2"
)

// Retrieve the default configuration from the github repository
func DefaultConfiguration() (Configuration, error) {
	github_url := "https://raw.githubusercontent.com/BoxBoxJason/svcadm/templates/svcadm.yaml"
	var config Configuration

	// Get the file content from the GitHub repository
	file_content, err := utils.GetFileContent(github_url)
	if err != nil {
		return Configuration{}, err
	}
	// Unmarshal the file content into the config map
	err = yaml.Unmarshal(file_content, &config)
	if err != nil {
		return Configuration{}, err
	}

	return config, nil
}

// Retrieve the list of valid services from the github repository
func retrieveValidServices() map[string]bool {
	return map[string]bool{
		"nginx":      true,
		"gitlab":     true,
		"mattermost": true,
		"sonarqube":  true,
		"postgresql": true,
		"clamav":     true,
		"trivy":      true,
		"minio":      true,
		"vault":      true,
	}
}

func serviceIsEnabled(service string) bool {
	config := GetConfiguration()
	for _, s := range config.Services {
		if s.Name == service && s.Enabled {
			return true
		}
	}
	return false
}
