package vaultadm

import (
	"errors"
	"fmt"
	"path"

	"github.com/boxboxjason/svcadm/internal/config"
	"github.com/boxboxjason/svcadm/internal/constants"
	"github.com/boxboxjason/svcadm/internal/services/svcadm"
	"github.com/boxboxjason/svcadm/pkg/containerutils"
	"github.com/boxboxjason/svcadm/pkg/logger"
	"github.com/boxboxjason/svcadm/pkg/utils"
)

const (
	VAULTADM            = "vaultadm"
	VAULTADM_LOG_PREFIX = "vaultadm:"
)

var (
	VAULT_PATH = path.Join(constants.SVCADM_HOME, "vaultadm")
)

type VaultAdm struct {
	Service config.Service
}

// CreateUser creates a user in the vault
func (v *VaultAdm) CreateUser(user *config.User) error {
	// admin_token, err := containerutils.GetContainerEnvVariable(v.Service.Container.Name, "ADMIN_TOKEN")
	// if err != nil {
	// 	return err
	// }
	return errors.New("not implemented")
}

// CreateAdminUser creates an admin user in the vault database
func (v *VaultAdm) CreateAdminUser(user *config.User) error {
	return errors.New("not implemented")
}

// PreInit sets up the vault database and environment variables
func (v *VaultAdm) PreInit() (map[string]string, map[string]string, map[int]int, []string, []string, error) {
	admin_token, err := utils.GenerateRandomPassword(128)
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}
	return map[string]string{"DOMAIN": fmt.Sprintf("https://%s/vault/", utils.GetHostname()), "ADMIN_TOKEN": admin_token}, nil, nil, nil, nil, nil
}

// PostInit Waits until the vault service is up and running, inits and unseals the vault, creates the users
func (v *VaultAdm) PostInit() error {
	err := v.WaitFor()
	if err != nil {
		return err
	}
	svcadm.CreateUsers(v, VAULTADM)
	return nil
}

func (v *VaultAdm) Backup(backup_path string) error {
	return nil
}

// WaitFor waits for the vault service to be ready
func (v *VaultAdm) WaitFor() error {
	const retry_interval = 10
	max_retries := 15
	for max_retries > 0 {
		err := containerutils.RunContainerCommand(v.Service.Container.Name, "/healthcheck.sh")
		if err == nil {
			logger.Info(VAULTADM_LOG_PREFIX, "vaultwarden container is ready")
			return nil
		}
		max_retries--
		logger.Debug(VAULTADM_LOG_PREFIX, "vaultwarden container is not ready, retrying in", retry_interval, "seconds")
	}
	return fmt.Errorf("timeout waiting for the vaultwarden server to start")
}

// GenerateNginxConf generates the nginx configuration for the vault service
func (v *VaultAdm) GenerateNginxConf() string {
	return fmt.Sprintf(`# Vault
location /%s/ {
      proxy_http_version 1.1;
      proxy_set_header Upgrade $http_upgrade;
      proxy_set_header Connection $connection_upgrade;

      proxy_set_header Host $host;
      proxy_set_header X-Real-IP $remote_addr;
      proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
      proxy_set_header X-Forwarded-Proto $scheme;

      proxy_pass http://%s:80;
}
`, v.Service.Name, v.Service.Container.Name)
}

// GetService returns the service configuration
func (v *VaultAdm) GetService() config.Service {
	return v.Service
}

func (v *VaultAdm) GetServiceName() string {
	return v.Service.Name
}

func (v *VaultAdm) GetServiceAdmName() string {
	return VAULTADM
}

func (v *VaultAdm) Cleanup() ([]string, []string) {
	return []string{}, []string{VAULT_PATH}
}
