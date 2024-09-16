package psqladm

import (
	"errors"
	"time"

	"github.com/boxboxjason/svcadm/internal/config"
	"github.com/boxboxjason/svcadm/pkg/utils"
)

// CreateUser creates a new user in the postgres cluster
func CreateUser(container_operator string, container_name string, username string, password string) error {
	return utils.RunContainerCommand(container_operator, container_name, "psql -U postgres -c \"CREATE USER "+username+" WITH PASSWORD '"+password+"';\"")
}

// CreateSuperUser creates a new superuser in the postgres cluster
func CreateSuperUser(container_operator string, container_name string, username string, password string) error {
	return utils.RunContainerCommand(container_operator, container_name, "psql -U postgres -c \"CREATE USER "+username+" WITH PASSWORD '"+password+"' SUPERUSER;\"")
}

// CreateDatabase creates a new database with the specified owner
func CreateDatabase(container_operator string, container_name string, database string, owner string) error {
	err := utils.RunContainerCommand(container_operator, container_name, "psql -U postgres -c \"CREATE DATABASE "+database+" OWNER "+owner+";\"")
	if err != nil {
		return err
	}
	return GrantUserDatabasePrivileges(container_operator, container_name, database, owner)
}

// GrantUserDatabasePrivileges grants all privileges on a database to a user
func GrantUserDatabasePrivileges(container_operator string, container_name string, database string, username string) error {
	return utils.RunContainerCommand(container_operator, container_name, "psql -U postgres -c \"GRANT ALL PRIVILEGES ON DATABASE "+database+" TO "+username+";\"")
}

// DeleteDatabase deletes a database
func DeleteDatabase(container_operator string, container_name string, database string) error {
	return utils.RunContainerCommand(container_operator, container_name, "psql -U postgres -c \"DROP DATABASE "+database+";\"")
}

// DeleteUser deletes a user
func DeleteUser(container_operator string, container_name string, username string) error {
	return utils.RunContainerCommand(container_operator, container_name, "psql -U postgres -c \"DROP USER "+username+";\"")
}

// BackupDatabase creates a backup of a database
func BackupDatabase(container_operator string, container_name string, database string, backup_name string) error {
	if database == "*" {
		return utils.RunContainerCommand(container_operator, container_name, "pg_dumpall -U postgres > "+backup_name)
	}
	return utils.RunContainerCommand(container_operator, container_name, "pg_dump -U postgres "+database+" > "+backup_name)
}

// PreInitPostgreSQL generates a random password for the postgres user
func PreInitPostgreSQL() (map[string]string, map[string]string, error) {
	password, err := utils.GenerateRandomPassword(32)
	if err != nil {
		return nil, nil, err
	}
	extended_env := map[string]string{
		"POSTGRES_PASSWORD": password,
	}
	return extended_env, nil, nil
}

// PostInitPostgreSQL creates the superusers and users in the postgres cluster
func PostInitPostgreSQL(container_operator string, container_name string, users *config.Users) error {
	err := waitForPostgres(container_operator, container_name)
	if err != nil {
		return err
	}

	for _, user := range users.Admins {
		err := CreateSuperUser(container_operator, container_name, user.Username, user.Password)
		if err != nil {
			return err
		}
	}
	for _, user := range users.Users {
		err := CreateUser(container_operator, container_name, user.Username, user.Password)
		if err != nil {
			return err
		}
	}
	return nil
}

// waitForPostgres waits for the postgres container to be ready
func waitForPostgres(container_operator string, container_name string) error {
	max_retries := 30
	const retry_interval = 5
	for max_retries > 0 {
		err := utils.RunContainerCommand(container_operator, container_name, "pg_isready")
		if err == nil {
			return nil
		}
		max_retries--
		time.Sleep(retry_interval * time.Second)
	}
	return errors.New("postgres container is not ready")
}
