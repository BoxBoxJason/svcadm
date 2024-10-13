package sonaradm

import (
	"encoding/json"
	"errors"
	"fmt"
	"path"
	"time"

	"github.com/boxboxjason/svcadm/internal/config"
	"github.com/boxboxjason/svcadm/internal/services/psqladm"
	"github.com/boxboxjason/svcadm/internal/services/svcadm"
	"github.com/boxboxjason/svcadm/pkg/containerutils"
	"github.com/boxboxjason/svcadm/pkg/formatutils"
	"github.com/boxboxjason/svcadm/pkg/logger"
	"github.com/boxboxjason/svcadm/pkg/utils"
)

const (
	SONAR_DB_USER = "sonarqube"
	SONAR_DB_NAME = "sonarqube"
)

type SonarAdm struct {
	Service config.Service
}

// PreInit sets up the sonarqube database and environment variables
func (s *SonarAdm) PreInit() (map[string]string, map[string]string, error) {
	additional_env := make(map[string]string)
	additional_volumes := make(map[string]string)

	// Create the sonarqube database
	db_password, err := utils.GenerateRandomPassword(32)
	if err != nil {
		return nil, nil, err
	}
	postgres_service := config.GetService("postgresql")
	p := psqladm.PsqlAdm{Service: postgres_service}

	err = p.CreateUser(&config.User{Username: SONAR_DB_USER, Password: db_password})
	if err != nil {
		logger.Error("Failed to create the sonarqube PostgreSQL user", err)
		return nil, nil, err
	} else {
		logger.Info("Successfully created the sonarqube PostgreSQL user")
	}
	err = p.CreateDatabase(SONAR_DB_NAME, SONAR_DB_NAME)
	if err != nil {
		logger.Error("Failed to create the sonarqube PostgreSQL database", err)
		return nil, nil, err
	} else {
		logger.Info("Successfully created the sonarqube PostgreSQL database")
	}

	// Generate a random password for the admin user
	admin_password, err := utils.GenerateRandomPassword(32)
	if err != nil {
		return nil, nil, err
	}
	additional_env["ADMIN_PASSWORD"] = admin_password

	additional_env["SONAR_JDBC_URL"] = fmt.Sprintf("jdbc:postgresql://%s:5432/%s", p.Service.Container.Name, SONAR_DB_NAME)
	additional_env["SONAR_JDBC_USERNAME"] = SONAR_DB_USER
	additional_env["SONAR_JDBC_PASSWORD"] = db_password

	if s.Service.Nginx {
		additional_env["SONAR_WEB_CONTEXT"] = "/sonarqube"
	}
	additional_env["SONAR_ES_CONNECTION_TIMEOUT"] = "1000"

	return additional_env, additional_volumes, nil
}

// PostInit Waits until the sonarqube service is up and running, then deletes the default admin user and creates the specified ones
func (s *SonarAdm) PostInit(env_variables map[string]string) error {
	err := s.WaitFor()
	if err != nil {
		return err
	}

	// Change the password of the default admin user
	err = containerutils.RunContainerCommand(s.Service.Container.Name, "curl", "-kfL", "-X", "POST", "http://localhost:9000/sonarqube/api/users/change_password", "-u", "admin:admin", "-d", "login=admin", "-d", "password="+env_variables["ADMIN_PASSWORD"], "-d", "previousPassword=admin")
	if err != nil {
		logger.Error("sonaradm: Failed to change the password of the default admin user", err)
		return err
	}
	svcadm.CreateUsers(s, "sonaradm")
	return nil
}

// WaitFor waits until the sonarqube server is up and running
func (s *SonarAdm) WaitFor() error {
	curl_command := []string{"curl", "-kfsL", "http://localhost:9000/sonarqube/api/system/status"}
	max_retry := 30
	const retry_interval = 10
	for max_retry > 0 {
		response, err := containerutils.RunContainerCommandWithOutput(s.Service.Container.Name, curl_command...)
		if err == nil {
			var result map[string]string
			err = json.Unmarshal(response, &result)
			if err == nil {
				if result["status"] == "UP" {
					logger.Info("sonarqube container is ready")
					return nil
				}
			}
		}
		logger.Debug("sonarqube container is not ready, retrying in", retry_interval, "seconds")
		max_retry--
		time.Sleep(retry_interval * time.Second)
	}
	return errors.New("timeout waiting for the sonarqube server to start")
}

// CreateUser creates a new user in the sonarqube server
func (s *SonarAdm) CreateUser(user *config.User) error {
	admin_password, err := s.retrieveAdminPassword()
	if err != nil {
		return err
	}
	return containerutils.RunContainerCommand(s.Service.Container.Name, "curl", "-kfL", "-X", "POST", "-u", "admin:"+admin_password,
		"-H", "Content-Type: application/json",
		"-d", fmt.Sprintf(`{"login":"%s","name":"%s","password":"%s"}`, user.Username, user.Username, user.Password),
		"http://localhost:9000/sonarqube/api/v2/users-management/users")
}

// CreateAdminUser creates a new admin user in the sonarqube server
func (s *SonarAdm) CreateAdminUser(user *config.User) error {
	admin_password, err := s.retrieveAdminPassword()
	if err != nil {
		return err
	}

	// Create the user and retrieve its ID
	err = s.CreateUser(user)
	if err != nil {
		logger.Error("sonaradm: failed to create the user", err)
		return err
	}

	// Curl the users to get the user ID
	users_output, err := containerutils.RunContainerCommandWithOutput(s.Service.Container.Name, "curl", "-kfs", "-u", "admin:"+admin_password,
		"http://localhost:9000/sonarqube/api/v2/users-management/users?q="+user.Username)
	if err != nil {
		return err
	}

	user_id, err := formatutils.RetrieveNestedId(users_output, "users", "login", user.Username, "id")
	if err != nil || user_id == "" {
		logger.Error("sonaradm: could not find the user ID for", user.Username, "the user was created but not added to the sonar-administrators group")
		return errors.New("could not find the user ID for " + user.Username)
	}

	// Curl the groups to get the group ID
	groups_response, err := containerutils.RunContainerCommandWithOutput(s.Service.Container.Name, "curl", "-kfs", "-u", "admin:"+admin_password,
		"http://localhost:9000/sonarqube/api/v2/authorizations/groups?q=sonar-administrators")
	if err != nil {
		logger.Error("sonaradm: could not retrieve the groups, the user was created but not added to the sonar-administrators group")
		return err
	}

	group_id, err := formatutils.RetrieveNestedId(groups_response, "groups", "name", "sonar-administrators", "id")
	if err != nil || group_id == "" {
		logger.Error("sonaradm: could not find the sonar-administrators group, the user was created but not added to the sonar-administrators group")
		return errors.New("sonaradm: could not find the sonar-administrators group")
	}

	return containerutils.RunContainerCommand(s.Service.Container.Name, "curl", "-kf", "-X", "POST",
		"-u", "admin:"+admin_password, "http://localhost:9000/sonarqube/api/v2/authorizations/group-memberships",
		"-H", "Content-Type: application/json",
		"-d", fmt.Sprintf(`{"userId":"%s","groupId":"%s"}`, user_id, group_id))
}

// Backup creates a backup of the sonarqube database and data
func (s *SonarAdm) Backup(backup_path string) error {
	postgres_service := config.GetService("postgresql")
	p := psqladm.PsqlAdm{Service: postgres_service}

	backup_name := utils.GenerateDatetimeString()
	err := p.BackupDatabase(SONAR_DB_NAME, path.Join(s.Service.Backup.Location, backup_name+".sql"))
	if err != nil {
		logger.Error("Failed to backup the sonarqube PostgreSQL database", err)
	} else {
		logger.Info("Successfully backed up the sonarqube PostgreSQL database to " + path.Join(s.Service.Backup.Location, backup_name+".sql"))
	}

	err = containerutils.RunContainerCommand(s.Service.Container.Name, "tar", "-cJf", "/tmp/sonarqube-backup.tar.xz", "$SONARQUBE_HOME/conf/", "$SONARQUBE_HOME/extensions/", "$SONARQUBE_HOME/data/")
	if err != nil {
		logger.Error("Failed to backup the sonarqube data", err)
		return err
	}
	err = containerutils.CopyContainerFile(s.Service.Container.Name, "/tmp/sonarqube-backup.tar.xz", backup_path)
	if err != nil {
		logger.Error("Failed to copy the sonarqube data backup", err)
		return err
	}
	logger.Info("Successfully backed up the sonarqube data to " + backup_path)

	return containerutils.RunContainerCommand(s.Service.Container.Name, "rm", "-f", "/tmp/sonarqube-backup.tar.xz")
}

// GenerateNginxConf generates the nginx configuration for the sonarqube service
func (s *SonarAdm) GenerateNginxConf() string {
	return fmt.Sprintf(`# SonarQube
location /%s/ {
    proxy_pass http://%s:9000;
    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    proxy_set_header X-Forwarded-Proto $scheme;
}`, s.Service.Name, s.Service.Container.Name)
}

// InitArgs returns the additional arguments / command required to start the sonarqube container
func (s *SonarAdm) InitArgs() []string {
	return []string{}
}

// GetService returns the service configuration
func (s *SonarAdm) GetService() config.Service {
	return s.Service
}

func (s *SonarAdm) ContainerArgs() []string {
	return []string{}
}

func (s *SonarAdm) retrieveAdminPassword() (string, error) {
	return containerutils.GetContainerEnvVariable(s.Service.Container.Name, "ADMIN_PASSWORD")
}
