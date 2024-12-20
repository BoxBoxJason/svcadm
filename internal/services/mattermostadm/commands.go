package mattermostadm

import (
	"encoding/json"
	"fmt"
	"path"
	"time"

	"github.com/boxboxjason/svcadm/internal/config"
	"github.com/boxboxjason/svcadm/internal/services/psqladm"
	"github.com/boxboxjason/svcadm/internal/services/svcadm"
	"github.com/boxboxjason/svcadm/pkg/containerutils"
	"github.com/boxboxjason/svcadm/pkg/logger"
	"github.com/boxboxjason/svcadm/pkg/utils"
)

const (
	MATTERMOST_DB_USER       = "mattermost"
	MATTERMOST_DB_NAME       = "mattermost"
	MATTERMOSTADM            = "mattermostadm"
	MATTERMOSTADM_LOG_PREFIX = "mattermostadm:"
)

type MattermostAdm struct {
	Service config.Service
}

// CreateUser creates a user in the mattermost server
func (m *MattermostAdm) CreateUser(user *config.User) error {
	return containerutils.RunContainerCommand(m.Service.Container.Name, "mmctl", "user", "create", "--email", user.Email, "--username", user.Username, "--password", user.Password, "--local")
}

// CreateAdminUser creates an admin user in the mattermost server
func (m *MattermostAdm) CreateAdminUser(user *config.User) error {
	return containerutils.RunContainerCommand(m.Service.Container.Name, "mmctl", "user", "create", "--email", user.Email, "--username", user.Username, "--password", user.Password, "--system-admin", "--local")
}

// PreInit sets up the mattermost database and environment variables, err
func (m *MattermostAdm) PreInit() (map[string]string, map[string]string, map[int]int, []string, []string, error) {
	// Create the mattermost database
	db_password, err := utils.GenerateRandomPassword(32)
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}
	postgres_service := config.GetService("postgresql")
	p := psqladm.PsqlAdm{Service: postgres_service}

	err = p.CreateUser(&config.User{Username: MATTERMOST_DB_USER, Password: db_password})
	if err != nil {
		logger.Error(MATTERMOSTADM_LOG_PREFIX, "failed to create the mattermost PostgreSQL user")
		return nil, nil, nil, nil, nil, err
	} else {
		logger.Info(MATTERMOSTADM_LOG_PREFIX, "successfully created the mattermost PostgreSQL user")
	}
	err = p.CreateDatabase(MATTERMOST_DB_NAME, MATTERMOST_DB_USER)
	if err != nil {
		logger.Error(MATTERMOSTADM_LOG_PREFIX, "failed to create the mattermost PostgreSQL database")
		return nil, nil, nil, nil, nil, err
	} else {
		logger.Info(MATTERMOSTADM_LOG_PREFIX, "successfully created the mattermost PostgreSQL database")
	}

	extended_env := map[string]string{
		"MM_SQLSETTINGS_DATASOURCE":          fmt.Sprintf("postgres://%s:%s@%s:5432/%s?binary_parameters=yes&sslmode=disable&connect_timeout=10", MATTERMOST_DB_USER, db_password, p.Service.Container.Name, MATTERMOST_DB_NAME),
		"MM_SERVICESETTINGS_ENABLELOCALMODE": "true",
	}

	if m.Service.Nginx {
		extended_env["MM_SERVICESETTINGS_SITEURL"] = fmt.Sprintf("https://%s/mattermost", utils.GetHostname())
	}

	return extended_env, nil, nil, nil, nil, nil
}

// PostInit creates the users and admins in the mattermost service
func (m *MattermostAdm) PostInit() error {
	err := m.WaitFor()
	if err != nil {
		logger.Error(err)
		return err
	}
	svcadm.CreateUsers(m, MATTERMOSTADM)

	return nil
}

// Backup creates a backup of the mattermost service
func (m *MattermostAdm) Backup(backup_path string) error {
	postgres_service := config.GetService("postgresql")
	p := psqladm.PsqlAdm{Service: postgres_service}

	backup_name := utils.GenerateDatetimeString()
	err := p.BackupDatabase(MATTERMOST_DB_NAME, path.Join(m.Service.Backup.Location, backup_name+".sql"))
	if err != nil {
		logger.Error(MATTERMOSTADM_LOG_PREFIX, "failed to backup the mattermost PostgreSQL database")
		return err
	} else {
		logger.Info(MATTERMOSTADM_LOG_PREFIX, "successfully backed up the mattermost PostgreSQL database to "+path.Join(m.Service.Backup.Location, backup_name+".sql"))
	}

	err = containerutils.RunContainerCommand(m.Service.Container.Name, "mattermost", "export", "bulk", "--all", "--destination", path.Join("tmp", backup_name+".zip"))
	if err != nil {
		logger.Error(MATTERMOSTADM_LOG_PREFIX, "failed to export the mattermost data")
		return err
	}
	err = containerutils.CopyContainerResource(m.Service.Container.Name, path.Join("tmp", backup_name+".zip"), backup_path)
	if err != nil {
		logger.Error(MATTERMOSTADM_LOG_PREFIX, "failed to copy the mattermost backup onto the host machine")
		return err
	}
	return containerutils.RunContainerCommand(m.Service.Container.Name, "rm", "-f", path.Join("tmp", backup_name+".zip"))
}

// GenerateNginxConf generates the nginx configuration for the mattermost service
func (m *MattermostAdm) GenerateNginxConf() string {
	return fmt.Sprintf(`location /%s {
	proxy_pass http://%s:8065;
	proxy_set_header Host $host;
	proxy_set_header X-Real-IP $remote_addr;
	proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
	proxy_set_header X-Forwarded-Proto $scheme;
	proxy_set_header Upgrade $http_upgrade;
	proxy_redirect off;
}`, m.Service.Name, m.Service.Container.Name)
}

// WaitFor waits until the mattermost server is up and running
func (m *MattermostAdm) WaitFor() error {
	curl_command := []string{"curl", "-kfsL", "http://localhost:8065/mattermost/api/v4/system/ping"}
	const retry_interval = 20
	max_retries := 15
	for max_retries > 0 {
		response, err := containerutils.RunContainerCommandWithOutput(m.Service.Container.Name, curl_command...)
		if err == nil {
			var result map[string]string
			err = json.Unmarshal(response, &result)
			if err == nil && result["status"] == "OK" {
				logger.Info(MATTERMOSTADM_LOG_PREFIX, "mattermost container is ready")
				return nil
			}
		}
		logger.Debug(MATTERMOSTADM_LOG_PREFIX, "mattermost container is not ready, retrying in", retry_interval, "seconds")
		max_retries--
		time.Sleep(retry_interval * time.Second)
	}
	return fmt.Errorf("timeout waiting for mattermost to be ready")
}

// GetService returns the service object from the configuration
func (m *MattermostAdm) GetService() config.Service {
	return m.Service
}

func (m *MattermostAdm) GetServiceName() string {
	return m.Service.Name
}

func (m *MattermostAdm) GetServiceAdmName() string {
	return MATTERMOSTADM
}

func (m *MattermostAdm) Cleanup() ([]string, []string) {
	return []string{}, []string{}
}
