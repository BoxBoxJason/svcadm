package svcadm

import (
	"github.com/boxboxjason/svcadm/internal/config"
	"github.com/boxboxjason/svcadm/pkg/logger"
)

type ServiceAdm interface {
	PreInit() (map[string]string, map[string]string, error)
	PostInit(env_variables map[string]string) error
	CreateUser(user *config.User) error
	CreateAdminUser(user *config.User) error
	Backup(backup_path string) error
	WaitFor() error
	GenerateNginxConf() string
	InitArgs() []string
	GetService() config.Service
	ContainerArgs() []string
	GetServiceName() string
	GetServiceAdmName() string
	Cleanup() ([]string, []string)
}

func CreateUsers(service_adm ServiceAdm, service_adm_slug string) {
	users := config.GetUsers()
	// Create the specified admin users
	for _, user := range users.Admins {
		err := (service_adm).CreateAdminUser(&user)
		if err != nil {
			logger.Error("could not create the admin user " + user.Username)
		} else {
			logger.Debug("created the admin user " + user.Username)
		}
	}
	// Create the specified users
	for _, user := range users.Users {
		err := (service_adm).CreateUser(&user)
		if err != nil {
			logger.Error("could not create the user " + user.Username)
		} else {
			logger.Debug("created the user " + user.Username)
		}
	}
}
