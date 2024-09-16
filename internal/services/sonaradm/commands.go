package sonaradm

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"time"

	"github.com/boxboxjason/svcadm/internal/config"
	"github.com/boxboxjason/svcadm/internal/services/psqladm"
	"github.com/boxboxjason/svcadm/pkg/logger"
	"github.com/boxboxjason/svcadm/pkg/utils"
)

// CreateSonarQubeDatabase creates the sonarqube database and user with a generated password
func createSonarQubeDatabase(container_operator string, postgresql_container_name string) (string, error) {
	const sonarqube_user = "sonarqube"
	const sonarqube_database = "sonarqube"

	password, err := utils.GenerateRandomPassword(32)
	if err != nil {
		return "", err
	}
	err = psqladm.CreateUser(container_operator, postgresql_container_name, sonarqube_user, password)
	if err != nil {
		return "", err
	}
	return password, psqladm.CreateDatabase(container_operator, postgresql_container_name, sonarqube_database, sonarqube_user)
}

// GenerateSonarQubeNginxConf generates the nginx configuration for the sonarqube service
func GenerateSonarQubeNginxConf(sonarqube_service *config.Service) string {
	return fmt.Sprintf(`# SonarQube
location /sonarqube {
    proxy_pass http://%s:9000;
    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    proxy_set_header X-Forwarded-Proto $scheme;
}`, sonarqube_service.Container.Name)
}

// PreInitSonarQube sets up the sonarqube database and environment variables
func PreInitSonarQube(container_operator string, sonarqube_service *config.Service, postgresql_service *config.Service) (map[string]string, map[string]string, error) {
	additional_env := make(map[string]string)
	additional_volumes := make(map[string]string)

	// Create the sonarqube database
	db_password, err := createSonarQubeDatabase(container_operator, postgresql_service.Container.Name)
	if err != nil {
		return nil, nil, err
	}

	additional_env["SONARQUBE_JDBC_URL"] = fmt.Sprintf("jdbc:postgresql://%s:5432/sonarqube", postgresql_service.Container.Name)
	additional_env["SONARQUBE_JDBC_USERNAME"] = "sonarqube"
	additional_env["SONARQUBE_JDBC_PASSWORD"] = db_password
	additional_env["SONARQUBE_WEB_CONTEXT"] = "/sonarqube"
	additional_env["SONAR_ES_CONNECTION_TIMEOUT"] = "1000"

	return additional_env, additional_volumes, nil
}

// PostInitSonarQube Waits until the sonarqube service is up and running, then deletes the default admin user and creates the specified ones
func PostInitSonarQube(container_operator string, sonarqube_service *config.Service, users *config.Users) error {
	err := waitForSonarQube(container_operator, sonarqube_service)
	if err != nil {
		return err
	}

	// Create the specified users
	for _, user := range users.Users {
		err := createUser(container_operator, sonarqube_service, user.Username, user.Password)
		if err != nil {
			return err
		}
	}
	// Create the specified admin users
	for _, user := range users.Admins {
		err := createAdminUser(container_operator, sonarqube_service, user.Username, user.Password)
		if err != nil {
			return err
		}
	}
	// Delete the default admin user
	cmd := exec.Command(container_operator, "exec", sonarqube_service.Container.Name, "curl -kfsL  POST -u admin:admin http://localhost:9000/sonarqube/api/users/deactivate?login=admin")
	return cmd.Run()
}

// waitForSonarQube waits until the sonarqube server is up and running
func waitForSonarQube(container_operator string, sonarqube_service *config.Service) error {
	const curl_command = "curl -kfsL http://localhost:9000/sonarqube/api/system/status"
	max_retry := 60
	const retry_interval = 5
	var cmd = exec.Command(container_operator, "exec", sonarqube_service.Container.Name, curl_command)
	status := "unknown"
	for status != "UP" && max_retry > 0 {
		response, err := cmd.Output()
		if err == nil {
			var result map[string]string
			err = json.Unmarshal(response, &result)
			if err == nil {
				status = result["status"]
			}
		}
		max_retry--
		time.Sleep(retry_interval * time.Second)
	}
	return nil
}

// createUser creates a new user in the sonarqube server
func createUser(container_operator string, sonarqube_service *config.Service, username string, password string) error {
	curl_command := "curl -kfsL -X POST -u admin:admin -d \"login=" + username + "&name=" + username + "&password=" + password + "\" http://localhost:9000/sonarqube/api/users/create"
	cmd := exec.Command(container_operator, "exec", sonarqube_service.Container.Name, curl_command)
	_, err := cmd.Output()
	return err
}

// createAdminUser creates a new admin user in the sonarqube server
func createAdminUser(container_operator string, sonarqube_service *config.Service, username string, password string) error {
	curl_command := "curl -kfsL -X POST -u admin:admin -d \"login=" + username + "&name=" + username + "&password=" + password + "&scmAccounts=" + username + "&groups=sonar-administrators\" http://localhost:9000/sonarqube/api/users/create"
	cmd := exec.Command(container_operator, "exec", sonarqube_service.Container.Name, curl_command)
	_, err := cmd.Output()
	return err
}

// BackupSonarqube creates a backup of the sonarqube database and data
func BackupSonarqube(container_operator string, sonarqube_service *config.Service, backup_name string, postgresql_container string) error {
	err := psqladm.BackupDatabase(container_operator, postgresql_container, "sonarqube", backup_name)
	if err != nil {
		logger.Error("Failed to backup the sonarqube PostgreSQL database", err)
	}

	err = utils.RunContainerCommand(container_operator, sonarqube_service.Container.Name, "tar -czf /tmp/sonarqube-backup.tar.gz /opt/sonarqube/data")
	if err != nil {
		logger.Error("Failed to backup the sonarqube data", err)
		return err
	}
	cmd := exec.Command(container_operator, "cp", sonarqube_service.Container.Name+":/tmp/sonarqube-backup.tar.gz", backup_name)
	err = cmd.Run()
	if err != nil {
		logger.Error("Failed to copy the sonarqube data backup", err)
		return err
	}

	return utils.RunContainerCommand(container_operator, sonarqube_service.Container.Name, "rm -f /tmp/sonarqube-backup.tar.gz")
}
