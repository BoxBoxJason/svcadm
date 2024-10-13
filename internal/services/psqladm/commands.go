package psqladm

import (
	"errors"
	"fmt"
	"path"
	"time"

	"github.com/boxboxjason/svcadm/internal/config"
	"github.com/boxboxjason/svcadm/internal/services/svcadm"
	"github.com/boxboxjason/svcadm/pkg/containerutils"
	"github.com/boxboxjason/svcadm/pkg/logger"
	"github.com/boxboxjason/svcadm/pkg/utils"
)

type PsqlAdm struct {
	Service config.Service
}

// CreateUser creates a new user in the postgres cluster
func (p *PsqlAdm) CreateUser(user *config.User) error {
	err := containerutils.RunContainerCommand(p.Service.Container.Name, "psql", "-U", "postgres", "-c", fmt.Sprintf("CREATE USER %s WITH PASSWORD '%s';", user.Username, user.Password))
	if err != nil {
		logger.Error("psqladm: Failed to create the user "+user.Username, err)
	} else {
		logger.Info("psqladm: Successfully created the user " + user.Username)
	}
	return err
}

// CreateAdminUser creates a new superuser in the postgres cluster
func (p *PsqlAdm) CreateAdminUser(user *config.User) error {
	err := containerutils.RunContainerCommand(p.Service.Container.Name, "psql", "-U", "postgres", "-c", fmt.Sprintf("CREATE USER %s WITH PASSWORD '%s' SUPERUSER;", user.Username, user.Password))
	if err != nil {
		logger.Error("psqladm: Failed to create the user "+user.Username, err)
	} else {
		logger.Info("psqladm: Successfully created the user " + user.Username)
	}
	return err
}

// CreateDatabase creates a new database with the specified owner
func (p *PsqlAdm) CreateDatabase(database string, owner string) error {
	err := containerutils.RunContainerCommand(p.Service.Container.Name, "psql", "-U", "postgres", "-c", fmt.Sprintf("CREATE DATABASE %s OWNER %s;", database, owner))
	if err != nil {
		logger.Error("Failed to create the database ", database, err)
		return err
	}
	return p.GrantUserDatabasePrivileges(database, owner)
}

// GrantUserDatabasePrivileges grants all privileges on a database to a user
func (p *PsqlAdm) GrantUserDatabasePrivileges(database string, username string) error {
	return containerutils.RunContainerCommand(p.Service.Container.Name, "psql", "-U", "postgres", "-c", fmt.Sprintf("GRANT ALL PRIVILEGES ON DATABASE %s TO %s;", database, username))
}

// DeleteDatabase deletes a database
func (p *PsqlAdm) DeleteDatabase(database string) error {
	return containerutils.RunContainerCommand(p.Service.Container.Name, "psql", "-U", "postgres", "-c", fmt.Sprintf("DROP DATABASE %s;", database))
}

// DeleteUser deletes a user
func (p *PsqlAdm) DeleteUser(username string) error {
	return containerutils.RunContainerCommand(p.Service.Container.Name, "psql", "-U", "postgres", "-c", fmt.Sprintf("DROP USER %s;", username))
}

// Backup creates a backup of the postgres cluster
func (p *PsqlAdm) Backup(backup_path string) error {
	backup_name := path.Join(backup_path, utils.GenerateDatetimeString()+".sql")
	return containerutils.RunContainerCommand(p.Service.Container.Name, "pg_dumpall", "-c", "-U", "postgres", "> "+backup_name)
}

// BackupDatabase creates a backup of a specific database
func (p *PsqlAdm) BackupDatabase(database string, backup_path string) error {
	backup_name := path.Join(backup_path, database+"_"+utils.GenerateDatetimeString()+".sql")
	return containerutils.RunContainerCommand(p.Service.Container.Name, "pg_dump", "-U", "postgres", database, "> "+backup_name)
}

// PreInit generates a random password for the postgres user
func (p *PsqlAdm) PreInit() (map[string]string, map[string]string, error) {
	password, err := utils.GenerateRandomPassword(32)
	if err != nil {
		logger.Error("Failed to generate a random password", err)
		return nil, nil, err
	}
	extended_env := map[string]string{
		"POSTGRES_PASSWORD": password,
	}
	return extended_env, nil, nil
}

// PostInit creates the superusers and users in the postgres cluster
func (p *PsqlAdm) PostInit(env_variables map[string]string) error {
	err := p.WaitFor()
	if err != nil {
		return err
	}

	svcadm.CreateUsers(p, "psqladm")
	return nil
}

// WaitFor waits for the postgres container to be ready
func (p *PsqlAdm) WaitFor() error {
	max_retries := 30
	const retry_interval = 5
	for max_retries > 0 {
		err := containerutils.RunContainerCommand(p.Service.Container.Name, "pg_isready")
		if err == nil {
			logger.Info("postgres container is ready")
			return nil
		}
		logger.Debug("postgres container is not ready, retrying in", retry_interval, "seconds")
		max_retries--
		time.Sleep(retry_interval * time.Second)
	}
	return errors.New("timeout exceeded, postgres container is not ready")
}

// GenerateNginxConf generates the nginx configuration for the postgres cluster
func (p *PsqlAdm) GenerateNginxConf() string {
	return ""
}

// InitArgs returns the additional arguments / command required to start the postgres container
func (p *PsqlAdm) InitArgs() []string {
	return []string{}
}

// GetService returns the service configuration
func (p *PsqlAdm) GetService() config.Service {
	return p.Service
}

func (p *PsqlAdm) ContainerArgs() []string {
	return []string{}
}
