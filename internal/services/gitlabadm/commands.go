package gitlabadm

import (
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
	GITLAB_DB_USER       = "gitlab"
	GITLAB_DB_NAME       = "gitlab"
	GITLABADM            = "gitlabadm"
	GITLABADM_LOG_PREFIX = "gitlabadm:"
)

type GitLabAdm struct {
	Service config.Service
}

// CreateAdminUser creates a new admin user in the gitlab instance
func (g *GitLabAdm) CreateAdminUser(user *config.User) error {
	return containerutils.RunContainerCommand(g.Service.Container.Name, "gitlab-rails", "runner", "-e", "production", fmt.Sprintf("u = User.new(name: '%s', username: '%s', email: '%s', password: '%s', password_confirmation: '%s', admin: true); u.assign_personal_namespace(Organizations::Organization.default_organization); u.skip_confirmation!; u.save!", user.Username, user.Username, user.Email, user.Password, user.Password))
}

// CreateUser creates a new user in the gitlab instance
func (g *GitLabAdm) CreateUser(user *config.User) error {
	return containerutils.RunContainerCommand(g.Service.Container.Name, "gitlab-rails", "runner", "-e", "production", fmt.Sprintf("u = User.new(name: '%s', username: '%s', email: '%s', password: '%s', password_confirmation: '%s'); u.assign_personal_namespace(Organizations::Organization.default_organization); u.skip_confirmation!; u.save!", user.Username, user.Username, user.Email, user.Password, user.Password))
}

// Backup creates a backup of the gitlab instance and saves it to the specified path
func (g *GitLabAdm) Backup(backup_path string) error {
	postgres_service := config.GetService("postgresql")
	p := psqladm.PsqlAdm{Service: postgres_service}

	backup_name := utils.GenerateDatetimeString()
	err := p.BackupDatabase(GITLAB_DB_NAME, path.Join(g.Service.Backup.Location, backup_name+".sql"))
	if err != nil {
		logger.Error(GITLABADM_LOG_PREFIX, "Failed to backup the GitLab PostgreSQL database")
	} else {
		logger.Info(GITLABADM_LOG_PREFIX, "Successfully backed up the GitLab PostgreSQL database to "+path.Join(g.Service.Backup.Location, backup_name+".sql"))
	}

	err = containerutils.RunContainerCommand(g.Service.Container.Name, "gitlab-backup", "create", "BACKUP="+backup_name)
	if err != nil {
		logger.Error(GITLABADM_LOG_PREFIX, "Failed to create the GitLab backup")
		return err
	}
	container_backup_path := fmt.Sprintf("/var/opt/gitlab/backups/%s_gitlab_backup.tar", backup_name)
	err = containerutils.CopyContainerResource(g.Service.Container.Name, container_backup_path, backup_path)
	if err != nil {
		logger.Error(GITLABADM_LOG_PREFIX, "Failed to copy the GitLab backup onto the host machine")
		return err
	}

	logger.Info(GITLABADM_LOG_PREFIX, "Successfully backed up the GitLab data to "+backup_path)

	return containerutils.RunContainerCommand(g.Service.Container.Name, "rm", "-f", container_backup_path)
}

// GenerateNginxConf generates the nginx configuration for the gitlab instance
func (g *GitLabAdm) GenerateNginxConf() string {
	return fmt.Sprintf(`# GitLab
location /%s/ {
	proxy_pass https://%s:443;
	proxy_set_header Host $host;
	proxy_set_header X-Real-IP $remote_addr;
	proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
	proxy_set_header X-Forwarded-Proto $scheme;
}`, g.Service.Name, g.Service.Container.Name)
}

// PostInit creates the users in the gitlab instance
func (g *GitLabAdm) PostInit(env_variables map[string]string) error {
	err := g.WaitFor()
	if err != nil {
		logger.Error(err)
		return err
	}

	svcadm.CreateUsers(g, GITLABADM)

	return nil
}

// PreInit generates a random password for the root user and sets up the gitlab database
func (g *GitLabAdm) PreInit() (map[string]string, map[string]string, []string, []string, error) {
	root_password, err := utils.GenerateRandomPassword(32)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	postgres_service := config.GetService("postgresql")
	p := psqladm.PsqlAdm{Service: postgres_service}

	// Set up the postgres user
	postgres_password, err := utils.GenerateRandomPassword(32)
	if err != nil {
		logger.Error(GITLABADM_LOG_PREFIX, "failed to generate a random password for the PostgreSQL user")
		return nil, nil, nil, nil, err
	}
	err = p.CreateUser(&config.User{Username: GITLAB_DB_USER, Password: postgres_password})
	if err != nil {
		logger.Error(GITLABADM_LOG_PREFIX, "failed to create the GitLab PostgreSQL user")
		return nil, nil, nil, nil, err
	} else {
		logger.Info(GITLABADM_LOG_PREFIX, "successfully created the GitLab PostgreSQL user")
	}
	// Set up the gitlab database
	err = p.CreateDatabase(GITLAB_DB_NAME, GITLAB_DB_USER)
	if err != nil {
		logger.Error(GITLABADM_LOG_PREFIX, "failed to create the GitLab PostgreSQL database")
		return nil, nil, nil, nil, err
	} else {
		logger.Info(GITLABADM_LOG_PREFIX, "successfully created the GitLab PostgreSQL database")
	}

	external_url := ""
	// Check if the gitlab service will be proxified by nginx
	if g.Service.Nginx {
		external_url = fmt.Sprintf("external_url 'https://%s/gitlab';", utils.GetHostname())
	}

	extended_env := map[string]string{
		"GITLAB_ROOT_PASSWORD":  root_password,
		"GITLAB_OMNIBUS_CONFIG": g.Service.Container.Env["GITLAB_OMNIBUS_CONFIG"] + fmt.Sprintf(" %s gitlab_rails['db_adapter'] = 'postgresql'; gitlab_rails['db_encoding'] = 'unicode'; gitlab_rails['db_database'] = '%s'; gitlab_rails['db_username'] = '%s'; gitlab_rails['db_password'] = '%s'; gitlab_rails['db_host'] = '%s'; gitlab_rails['db_port'] = '5432'; gitlab_rails['db_pool'] = 10", external_url, GITLAB_DB_NAME, GITLAB_DB_USER, postgres_password, postgres_service.Container.Name),
	}
	return extended_env, nil, nil, nil, nil
}

// WaitFor waits for the gitlab instance to be ready, using the curl readiness check
func (g *GitLabAdm) WaitFor() error {
	const retry_interval = 20
	max_retries := 15
	for max_retries > 0 {
		err := containerutils.RunContainerCommand(g.Service.Container.Name, "gitlab-healthcheck")
		if err == nil {
			logger.Info(GITLABADM_LOG_PREFIX, "gitlab container is ready")
			return nil
		}
		logger.Info(GITLABADM_LOG_PREFIX, "gitlab container is not ready yet, retrying in", retry_interval, " seconds")
		max_retries--
		time.Sleep(retry_interval * time.Second)
	}
	return fmt.Errorf("timed out waiting for the gitlab container to be ready")
}

// GetService returns the service object from the configuration
func (g *GitLabAdm) GetService() config.Service {
	return g.Service
}

func (g *GitLabAdm) GetServiceName() string {
	return g.Service.Name
}

func (g *GitLabAdm) GetServiceAdmName() string {
	return GITLABADM
}

func (g *GitLabAdm) Cleanup() ([]string, []string) {
	return []string{}, []string{}
}
