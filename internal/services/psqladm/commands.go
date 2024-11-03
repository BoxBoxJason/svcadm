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

const (
	PSQLADM            = "psqladm"
	PSQLADM_LOG_PREFIX = "psqladm:"
)

type PsqlAdm struct {
	Service config.Service
}

// CreateUser creates a new user in the postgres cluster
func (p *PsqlAdm) CreateUser(user *config.User) error {
	err := containerutils.RunContainerCommand(p.Service.Container.Name, "psql", "-U", "postgres", "-c", fmt.Sprintf("CREATE USER %s WITH PASSWORD '%s';", user.Username, user.Password))
	if err != nil {
		logger.Error(PSQLADM_LOG_PREFIX, "failed to create the user "+user.Username)
	} else {
		logger.Info(PSQLADM_LOG_PREFIX, "successfully created the user "+user.Username)
	}
	return err
}

// CreateAdminUser creates a new superuser in the postgres cluster
func (p *PsqlAdm) CreateAdminUser(user *config.User) error {
	err := containerutils.RunContainerCommand(p.Service.Container.Name, "psql", "-U", "postgres", "-c", fmt.Sprintf("CREATE USER %s WITH PASSWORD '%s' SUPERUSER;", user.Username, user.Password))
	if err != nil {
		logger.Error(PSQLADM_LOG_PREFIX, "failed to create the user "+user.Username)
	} else {
		logger.Info(PSQLADM_LOG_PREFIX, "successfully created the user "+user.Username)
	}
	return err
}

// CreateDatabase creates a new database with the specified owner
func (p *PsqlAdm) CreateDatabase(database string, owner string) error {
	err := containerutils.RunContainerCommand(p.Service.Container.Name, "psql", "-U", "postgres", "-c", fmt.Sprintf("CREATE DATABASE %s OWNER %s;", database, owner))
	if err != nil {
		logger.Error(PSQLADM_LOG_PREFIX, "failed to create the database ", database)
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
func (p *PsqlAdm) PreInit() (map[string]string, map[string]string, []string, []string, error) {
	password, err := utils.GenerateRandomPassword(32)
	if err != nil {
		logger.Error(PSQLADM_LOG_PREFIX, "failed to generate a random password")
		return nil, nil, nil, nil, err
	}
	extended_env := map[string]string{
		"POSTGRES_PASSWORD": password,
	}
	return extended_env, nil, nil, nil, nil
}

// PostInit creates the superusers and users in the postgres cluster
func (p *PsqlAdm) PostInit(env_variables map[string]string) error {
	err := p.WaitFor()
	if err != nil {
		return err
	}

	svcadm.CreateUsers(p, PSQLADM)
	return nil
}

// WaitFor waits for the postgres container to be ready
func (p *PsqlAdm) WaitFor() error {
	max_retries := 15
	const retry_interval = 5
	for max_retries > 0 {
		err := containerutils.RunContainerCommand(p.Service.Container.Name, "pg_isready")
		if err == nil {
			logger.Info(PSQLADM_LOG_PREFIX, "postgres container is ready")
			return nil
		}
		logger.Debug(PSQLADM_LOG_PREFIX, "postgres container is not ready, retrying in", retry_interval, "seconds")
		max_retries--
		time.Sleep(retry_interval * time.Second)
	}
	return errors.New("timeout exceeded, postgres container is not ready")
}

// GenerateNginxConf generates the nginx configuration for the postgres cluster
func (p *PsqlAdm) GenerateNginxConf() string {
	return ""
}

// GetService returns the service configuration
func (p *PsqlAdm) GetService() config.Service {
	return p.Service
}

func (p *PsqlAdm) GetServiceName() string {
	return p.Service.Name
}

func (p *PsqlAdm) GetServiceAdmName() string {
	return PSQLADM
}

func (p *PsqlAdm) Cleanup() ([]string, []string) {
	return []string{}, []string{}
}
