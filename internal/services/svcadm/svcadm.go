package svcadm

import (
	"github.com/boxboxjason/svcadm/internal/config"
	"github.com/boxboxjason/svcadm/pkg/logger"
)

type ServiceAdm interface {
	// PreInit creates the necessary environment variables, volumes, ports, cap_adds and entrypoint for the service
	PreInit() (map[string]string, map[string]string, map[int]int, []string, []string, error)
	// PostInit waits for the service to be up and running, creates the users and other necessary configurations
	PostInit() error
	// CreateUser creates a user in the service
	CreateUser(user *config.User) error
	// CreateAdminUser creates an admin user in the service
	CreateAdminUser(user *config.User) error
	// Backup creates a backup of the service at the specified path
	Backup(backup_path string) error
	// WaitFor waits for the service to be up and running
	WaitFor() error
	// GenerateNginxConf generates the nginx configuration for the service
	GenerateNginxConf() string
	// GetService returns the service configuration used by the service adm
	GetService() config.Service
	// GetServiceName returns the name of the service used by the service adm
	GetServiceName() string
	// GetServiceAdmName returns the name of the service adm (used for logging)
	GetServiceAdmName() string
	// Cleanup removes the service and its volumes
	Cleanup() ([]string, []string)
}

func CreateUsers(service_adm ServiceAdm, service_adm_slug string) {
	users := config.GetUsers()
	// Create the specified admin users
	for _, user := range users.Admins {
		err := (service_adm).CreateAdminUser(&user)
		if err != nil {
			logger.Error(service_adm.GetServiceAdmName()+":", "could not create the admin user "+user.Username)
		} else {
			logger.Debug(service_adm.GetServiceAdmName()+":", "created the admin user "+user.Username)
		}
	}
	// Create the specified users
	for _, user := range users.Users {
		err := (service_adm).CreateUser(&user)
		if err != nil {
			logger.Error(service_adm.GetServiceAdmName()+":", "could not create the user "+user.Username)
		} else {
			logger.Debug(service_adm.GetServiceAdmName()+":", "created the user "+user.Username)
		}
	}
}
