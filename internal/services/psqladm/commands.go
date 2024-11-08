package psqladm

import (
	"database/sql"
	"errors"
	"fmt"
	"path"
	"time"

	_ "github.com/lib/pq"

	"github.com/boxboxjason/svcadm/internal/config"
	"github.com/boxboxjason/svcadm/internal/services/svcadm"
	"github.com/boxboxjason/svcadm/pkg/containerutils"
	"github.com/boxboxjason/svcadm/pkg/logger"
	"github.com/boxboxjason/svcadm/pkg/utils"
)

const (
	PSQLADM            = "psqladm"
	PSQLADM_LOG_PREFIX = "psqladm:"
	DB_CONNECTION_INFO = "host=127.0.0.1 port=5432 user=postgres password=%s dbname=postgres sslmode=disable"
)

type PsqlAdm struct {
	Service config.Service
}

// CreateUser creates a new user in the postgres cluster
func (p *PsqlAdm) CreateUser(user *config.User) error {
	db, err := p.openDB()
	if err != nil {
		return err
	}
	defer db.Close()
	query := fmt.Sprintf("CREATE USER %s WITH PASSWORD '%s';", user.Username, user.Password)
	_, err = db.Exec(query)
	if err != nil {
		logger.Error(PSQLADM_LOG_PREFIX, "failed to create the user", user.Username, err)
	} else {
		logger.Info(PSQLADM_LOG_PREFIX, "successfully created the user", user.Username)
	}
	return err
}

// CreateAdminUser creates a new superuser in the postgres cluster
func (p *PsqlAdm) CreateAdminUser(user *config.User) error {
	db, err := p.openDB()
	if err != nil {
		return err
	}
	defer db.Close()

	query := fmt.Sprintf("CREATE USER %s WITH PASSWORD '%s' SUPERUSER;", user.Username, user.Password)
	_, err = db.Exec(query)
	if err != nil {
		logger.Error(PSQLADM_LOG_PREFIX, "failed to create the superuser", user.Username, err)
	} else {
		logger.Info(PSQLADM_LOG_PREFIX, "successfully created the superuser", user.Username)
	}
	return err
}

// CreateDatabase creates a new database with the specified owner
func (p *PsqlAdm) CreateDatabase(database string, owner string) error {
	db, err := p.openDB()
	if err != nil {
		return err
	}
	defer db.Close()

	query := fmt.Sprintf("CREATE DATABASE %s OWNER %s;", database, owner)
	_, err = db.Exec(query)
	if err != nil {
		logger.Error(PSQLADM_LOG_PREFIX, "failed to create the database", database, err)
	} else {
		logger.Info(PSQLADM_LOG_PREFIX, "successfully created the database", database)
	}
	return err
}

// GrantUserDatabasePrivileges grants all privileges on a database to a user
func (p *PsqlAdm) GrantUserDatabasePrivileges(database string, username string) error {
	db, err := p.openDB()
	if err != nil {
		return err
	}
	defer db.Close()

	query := fmt.Sprintf("GRANT ALL PRIVILEGES ON DATABASE %s TO %s;", database, username)
	_, err = db.Exec(query)
	if err != nil {
		logger.Error(PSQLADM_LOG_PREFIX, "failed to grant all privileges on database", database, "to user ", username, err)
	} else {
		logger.Info(PSQLADM_LOG_PREFIX, "successfully granted all privileges on database", database, "to user ", username)
	}
	return err
}

// DeleteDatabase deletes a database
func (p *PsqlAdm) DeleteDatabase(database string) error {
	db, err := p.openDB()
	if err != nil {
		return err
	}
	defer db.Close()

	query := fmt.Sprintf("DROP DATABASE %s;", database)
	_, err = db.Exec(query)
	if err != nil {
		logger.Error(PSQLADM_LOG_PREFIX, "failed to delete the database", database, err)
	} else {
		logger.Info(PSQLADM_LOG_PREFIX, "successfully deleted the database", database)
	}
	return err
}

// DeleteUser deletes a user
func (p *PsqlAdm) DeleteUser(username string) error {
	db, err := p.openDB()
	if err != nil {
		return err
	}
	defer db.Close()

	query := fmt.Sprintf("DROP USER %s;", username)
	_, err = db.Exec(query)
	if err != nil {
		logger.Error(PSQLADM_LOG_PREFIX, "failed to delete the user", username, err)
	} else {
		logger.Info(PSQLADM_LOG_PREFIX, "successfully deleted the user", username)
	}
	return err
}

// Backup creates a backup of the postgres cluster
func (p *PsqlAdm) Backup(backup_path string) error {
	backup_name := path.Join(backup_path, utils.GenerateDatetimeString()+".sql")
	db, err := p.openDB()
	if err != nil {
		query := fmt.Sprintf("\\q | pg_dumpall -U postgres > %s", backup_name)
		_, err = db.Exec(query)
		if err != nil {
			logger.Error(PSQLADM_LOG_PREFIX, "failed to create a backup", err)
		} else {
			logger.Info(PSQLADM_LOG_PREFIX, "successfully created a backup")
		}
		return err
	}
	defer db.Close()

	query := fmt.Sprintf("pg_dumpall > %s;", backup_path)
	_, err = db.Exec(query)
	if err != nil {
		logger.Error(PSQLADM_LOG_PREFIX, "failed to backup the postgres cluster", err)
	} else {
		logger.Info(PSQLADM_LOG_PREFIX, "successfully backed up the postgres cluster")
	}
	return err
}

// BackupDatabase creates a backup of a specific database
func (p *PsqlAdm) BackupDatabase(database string, backup_path string) error {
	backup_name := path.Join(backup_path, database+"_"+utils.GenerateDatetimeString()+".sql")
	db, err := p.openDB()
	if err != nil {
		return err
	}
	defer db.Close()

	query := fmt.Sprintf("pg_dump %s > %s;", database, backup_name)
	_, err = db.Exec(query)
	if err != nil {
		logger.Error(PSQLADM_LOG_PREFIX, "failed to backup the database", database, err)
	} else {
		logger.Info(PSQLADM_LOG_PREFIX, "successfully backed up the database", database)
	}
	return err
}

// PreInit generates a random password for the postgres user
func (p *PsqlAdm) PreInit() (map[string]string, map[string]string, map[int]int, []string, []string, error) {
	password, err := utils.GenerateRandomPassword(32)
	if err != nil {
		logger.Error(PSQLADM_LOG_PREFIX, "failed to generate a random password")
		return nil, nil, nil, nil, nil, err
	}
	extended_env := map[string]string{
		"POSTGRES_PASSWORD": password,
	}
	return extended_env, nil, map[int]int{5432: 5432}, nil, nil, nil
}

// PostInit creates the superusers and users in the postgres cluster
func (p *PsqlAdm) PostInit() error {
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
		db, err := p.openDB()
		if err == nil && db.Ping() == nil {
			logger.Info(PSQLADM_LOG_PREFIX, "postgres container is ready")
			return nil
		}
		time.Sleep(retry_interval * time.Second)
		max_retries--
		logger.Debug(PSQLADM_LOG_PREFIX, "postgres container is not ready, retrying in", retry_interval, "seconds")
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

func (p *PsqlAdm) openDB() (*sql.DB, error) {
	postgres_password, err := containerutils.GetContainerEnvVariable(p.Service.Container.Name, "POSTGRES_PASSWORD")
	if err != nil {
		logger.Error(PSQLADM_LOG_PREFIX, "failed to get the postgres password", err)
		return nil, err
	}
	db, err := sql.Open("postgres", fmt.Sprintf(DB_CONNECTION_INFO, postgres_password))
	if err != nil {
		logger.Error(PSQLADM_LOG_PREFIX, "failed to open the database connection", err)
		return nil, err
	}
	return db, err
}
