package vaultadm

import (
	"errors"
	"fmt"
	"path"
	"strings"
	"time"

	"github.com/boxboxjason/svcadm/internal/config"
	"github.com/boxboxjason/svcadm/internal/constants"
	"github.com/boxboxjason/svcadm/internal/services/svcadm"
	"github.com/boxboxjason/svcadm/pkg/containerutils"
	"github.com/boxboxjason/svcadm/pkg/fileutils"
	"github.com/boxboxjason/svcadm/pkg/logger"
)

const (
	VAULTADM            = "vaultadm"
	VAULTADM_LOG_PREFIX = "vaultadm:"
)

var (
	VAULT_PATH      = path.Join(constants.SVCADM_HOME, "vaultadm")
	ROOT_TOKEN_PATH = path.Join(VAULT_PATH, ".root_token")
	SEAL_KEY_PATH   = path.Join(VAULT_PATH, ".seal_%d")
)

type VaultAdm struct {
	Service config.Service
}

// CreateUser creates a user in the vault
func (v *VaultAdm) CreateUser(user *config.User) error {
	return containerutils.RunContainerCommand(v.Service.Container.Name, "vault", "write", "-address=http://localhost:8200", fmt.Sprintf("auth/userpass/users/%s", user.Username), fmt.Sprintf("password=%s", user.Password))
}

// CreateAdminUser creates an admin user in the vault database
func (v *VaultAdm) CreateAdminUser(user *config.User) error {
	return containerutils.RunContainerCommand(v.Service.Container.Name, "vault", "write", "-address=http://localhost:8200", fmt.Sprintf("auth/userpass/users/%s", user.Username), fmt.Sprintf("password=%s", user.Password), "policies=admin")
}

// PreInit sets up the vault database and environment variables
func (v *VaultAdm) PreInit() (map[string]string, map[string]string, map[int]int, []string, []string, error) {
	return nil, nil, nil, []string{"IPC_LOCK"}, []string{"server"}, nil
}

// PostInit Waits until the vault service is up and running, inits and unseals the vault, creates the users
func (v *VaultAdm) PostInit() error {
	err := v.WaitFor()
	if err != nil {
		return err
	}
	// Initialize the vault
	const key_threshold = 3
	init_command := []string{"vault", "operator", "init", "-address=http://localhost:8200", "-key-shares=5", fmt.Sprintf("-key-threshold=%d", key_threshold)}
	output, err := containerutils.RunContainerCommandWithOutput(v.Service.Container.Name, init_command...)
	if err != nil {
		logger.Error(VAULTADM_LOG_PREFIX, "could not initialize the vault")
		return err
	}
	fmt.Println(string(output))

	// Collect unseal keys & root token
	unseal_keys := []string{}
	var root_token string
	for _, line := range strings.Split(string(output), "\n") {
		if strings.Contains(line, "Unseal Key") {
			unseal_keys = append(unseal_keys, strings.Split(line, ": ")[1])
		}
		if strings.Contains(line, "Initial Root Token") {
			root_token = strings.Split(line, ": ")[1]
		}
	}

	// Save the root token
	saveAdminToken(root_token)

	// Save the unseal keys
	for i, key := range unseal_keys {
		saveSealKey(key, i+1)
	}

	// Unseal the vault
	unseal_command := []string{"vault", "operator", "unseal", "-address=http://localhost:8200"}
	for _, key := range unseal_keys[:key_threshold] {
		err = containerutils.RunContainerCommand(v.Service.Container.Name, append(unseal_command, key)...)
		if err != nil {
			logger.Error(VAULTADM_LOG_PREFIX, "could not unseal the vault")
			return err
		}
	}
	logger.Info(VAULTADM_LOG_PREFIX, "vault unsealed")

	// Login with the root token
	err = containerutils.RunContainerCommand(v.Service.Container.Name, "vault", "login", "-address=http://localhost:8200", root_token)
	if err != nil {
		logger.Error(VAULTADM_LOG_PREFIX, "could not login with the root token")
		return err
	}

	// Enable the userpass auth method
	err = containerutils.RunContainerCommand(v.Service.Container.Name, "vault", "auth", "enable", "-address=http://localhost:8200", "userpass")
	if err != nil {
		logger.Error(VAULTADM_LOG_PREFIX, "could not enable the userpass auth method")
		return err
	}

	// Create the admin policy
	err = containerutils.RunContainerCommand(v.Service.Container.Name, "sh", "-c", `echo 'path "/sys/*" {
	capabilities = ["create", "read", "update", "delete", "list", "sudo"]
}
path "/secret/*" {
	capabilities = ["create", "read", "update", "delete", "list"]
}' > admin.hcl`)
	if err != nil {
		logger.Error(VAULTADM_LOG_PREFIX, "could not create the admin policy file")
		return err
	}
	err = containerutils.RunContainerCommand(v.Service.Container.Name, "vault", "policy", "write", "-address=http://localhost:8200", "admin", "admin.hcl")
	if err != nil {
		logger.Error(VAULTADM_LOG_PREFIX, "could not create the admin policy")
		return err
	} else {
		logger.Debug(VAULTADM_LOG_PREFIX, "admin policy created")
	}

	containerutils.RunContainerCommand(v.Service.Container.Name, "rm", "-f", "admin.hcl")

	svcadm.CreateUsers(v, VAULTADM)
	return nil
}

// WaitFor waits for the vault service to be ready
func (v *VaultAdm) WaitFor() error {
	max_retries := 30
	const retry_interval = 5
	healthcheck_command := []string{"vault", "status", "-address=http://localhost:8200"}
	for max_retries > 0 {
		output, err := containerutils.RunContainerCommandWithOutput(v.Service.Container.Name, healthcheck_command...)
		if err == nil || strings.Contains(string(output), "Build Date") {
			logger.Info(VAULTADM_LOG_PREFIX, "vault container is ready")
			return nil
		}
		logger.Debug(VAULTADM_LOG_PREFIX, "vault container not ready yet, retrying in", retry_interval, "seconds")
		time.Sleep(retry_interval * time.Second)
		max_retries--
	}
	return errors.New("timeout waiting for vault to be ready")
}

// Backup creates a backup of the vault
func (v *VaultAdm) Backup(backup_path string) error { // TODO
	return nil
}

// GenerateNginxConf generates the nginx configuration for the vault service
func (v *VaultAdm) GenerateNginxConf() string {
	return fmt.Sprintf(`# Vault
location /%s/ {
	proxy_http_version 1.1;
	proxy_set_header Upgrade $http_upgrade;

	proxy_set_header Host $host;
	proxy_set_header X-Real-IP $remote_addr;
	proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
	proxy_set_header X-Forwarded-Proto $scheme;

	proxy_pass http://%s:8200/;
}`, v.Service.Name, v.Service.Container.Name)
}

// GetService returns the service configuration
func (v *VaultAdm) GetService() config.Service {
	return v.Service
}

// saveAdminToken saves the admin token to a file
func saveAdminToken(admin_token string) {
	err := fileutils.WriteToFile(ROOT_TOKEN_PATH, admin_token)
	if err != nil {
		logger.Error(VAULTADM_LOG_PREFIX, "could not save the root token, consider saving it manually at", ROOT_TOKEN_PATH)
	} else {
		logger.Info(VAULTADM_LOG_PREFIX, "root token saved at", ROOT_TOKEN_PATH)
	}
}

// saveSealKey saves the seal key to a file
func saveSealKey(seal_key string, index int) {
	err := fileutils.WriteToFile(fmt.Sprintf(SEAL_KEY_PATH, index), seal_key)
	if err != nil {
		logger.Error(VAULTADM_LOG_PREFIX, "could not save the seal key, consider saving it manually at", fmt.Sprintf(SEAL_KEY_PATH, index))
	} else {
		logger.Info(VAULTADM_LOG_PREFIX, "seal key saved at", fmt.Sprintf(SEAL_KEY_PATH, index))
	}
}

// getAdminToken returns the admin token from a file
func getAdminToken() (string, error) {
	return fileutils.GetFileContent(ROOT_TOKEN_PATH)
}

// getSealKey returns the seal key from a file
func getSealKey(index int) (string, error) {
	return fileutils.GetFileContent(fmt.Sprintf(SEAL_KEY_PATH, index))
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
