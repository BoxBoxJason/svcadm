package minioadm

import (
	"fmt"
	"path"

	"github.com/boxboxjason/svcadm/internal/config"
	"github.com/boxboxjason/svcadm/internal/services/svcadm"
	"github.com/boxboxjason/svcadm/pkg/containerutils"
	"github.com/boxboxjason/svcadm/pkg/logger"
	"github.com/boxboxjason/svcadm/pkg/utils"
)

const (
	MINIOADM            = "minioadm"
	MINIOADM_LOG_PREFIX = "minioadm:"
)

type MinioAdm struct {
	Service config.Service
}

// CreateUser creates a user in the minio server
func (m *MinioAdm) CreateUser(user *config.User) error {
	err := containerutils.RunContainerCommand(m.Service.Container.Name, "mc", "admin", "user", "add", m.Service.Container.Name, user.Username, user.Password)
	if err != nil {
		return err
	}
	err = containerutils.RunContainerCommand(m.Service.Container.Name, "mc", "admin", "policy", "attach", m.Service.Container.Name, "readwrite", "--user", user.Username)
	if err != nil {
		logger.Error(MINIOADM_LOG_PREFIX, "failed to attach the readwrite policy to the user "+user.Username)
	}
	return err
}

// CreateAdminUser creates an admin user in the minio server
func (m *MinioAdm) CreateAdminUser(user *config.User) error {
	err := containerutils.RunContainerCommand(m.Service.Container.Name, "mc", "admin", "user", "add", m.Service.Container.Name, user.Username, user.Password)
	if err != nil {
		return err
	}
	err = containerutils.RunContainerCommand(m.Service.Container.Name, "mc", "admin", "policy", "attach", m.Service.Container.Name, "consoleAdmin", "--user", user.Username)
	if err != nil {
		logger.Error(MINIOADM_LOG_PREFIX, "failed to attach the admin policy to the user "+user.Username)
	}
	return err
}

// BackupBucket creates a backup of a minio bucket on the operator's machine
func (m *MinioAdm) BackupBucket(bucket_name string, backup_path string) error {
	backup_name := path.Join(backup_path, bucket_name+".tar.xz")
	err := containerutils.RunContainerCommand(m.Service.Container.Name, "tar", "-cJf", "/tmp/"+bucket_name+".tar.xz", "/data/"+bucket_name)
	if err != nil {
		logger.Error(MINIOADM_LOG_PREFIX, "failed to backup the minio bucket")
		return err
	}
	err = containerutils.CopyContainerResource(m.Service.Container.Name, fmt.Sprintf("/tmp/%s.tar.xz", bucket_name), backup_name)
	if err != nil {
		logger.Error(MINIOADM_LOG_PREFIX, "failed to copy the minio bucket backup on the host machine")
		return err
	}
	return containerutils.RunContainerCommand(m.Service.Container.Name, "rm", "-f", fmt.Sprintf("/tmp/%s.tar.xz", bucket_name))
}

// Backup creates a backup of all minio buckets on the operator's machine
func (m *MinioAdm) Backup(backup_name string) error {
	err := containerutils.RunContainerCommand(m.Service.Container.Name, "tar", "-cJf", "/tmp/all.tar.xz", "/data")
	if err != nil {
		logger.Error(MINIOADM_LOG_PREFIX, "failed to backup the minio data")
		return err
	}
	err = containerutils.CopyContainerResource(m.Service.Container.Name, "/tmp/all.tar.xz", backup_name)
	if err != nil {
		logger.Error(MINIOADM_LOG_PREFIX, "failed to copy the minio data backup")
		return err
	} else {
		logger.Info(MINIOADM_LOG_PREFIX, "successfully backed up the minio data to "+backup_name)
	}
	return containerutils.RunContainerCommand(m.Service.Container.Name, "rm", "-f", "/tmp/all.tar.xz")
}

// CreateBucket creates a new bucket in the minio server
func (m *MinioAdm) CreateBucket(bucket_name string) error {
	return containerutils.RunContainerCommand(m.Service.Container.Name, "mc", "mb", fmt.Sprintf("%s/%s", m.Service.Container.Name, bucket_name))
}

// DeleteBucket deletes a bucket from the minio server
func (m *MinioAdm) DeleteBucket(bucket_name string) error {
	return containerutils.RunContainerCommand(m.Service.Container.Name, "mc", "rb", fmt.Sprintf("%s/%s", m.Service.Container.Name, bucket_name), "--force")
}

// PreInit runs the pre init steps for the minio server
func (m *MinioAdm) PreInit() (map[string]string, map[string]string, map[int]int, []string, []string, error) {
	extended_env := make(map[string]string)
	var err error

	// Set the environment variables for the minio server
	root_password, err := utils.GenerateRandomPassword(32)
	if err != nil {
		logger.Error(MINIOADM_LOG_PREFIX, "failed to generate a random password")
		return nil, nil, nil, nil, nil, err
	}

	extended_env["MINIO_ROOT_USER"] = "minioadmin"
	extended_env["MINIO_ROOT_PASSWORD"] = root_password
	extended_env["MINIO_BROWSER_REDIRECT_URL"] = fmt.Sprintf("https://%s/minio", utils.GetHostname())

	return extended_env, nil, nil, nil, []string{"server", "/data", "--console-address", ":9001"}, nil
}

// PostInit runs the post init steps for the minio server
func (m *MinioAdm) PostInit() error {
	// Wait for the minio server to be ready
	err := m.WaitFor()
	if err != nil {
		logger.Error(err)
		return err
	}

	minio_root_user, err := containerutils.GetContainerEnvVariable(m.Service.Container.Name, "MINIO_ROOT_USER")
	if err != nil {
		return err
	}
	minio_root_password, err := containerutils.GetContainerEnvVariable(m.Service.Container.Name, "MINIO_ROOT_PASSWORD")
	if err != nil {
		return err
	}

	// Fix the default aliases because its not done by default ???????????
	for _, alias := range []string{"local", "gcs", "s3", "play"} {
		_ = containerutils.RunContainerCommand(m.Service.Container.Name, "mc", "alias", "remove", alias)
	}
	err = containerutils.RunContainerCommand(m.Service.Container.Name, "mc", "alias", "set", m.Service.Container.Name, "http://localhost:9000", minio_root_user, minio_root_password)
	if err != nil {
		return err
	}

	svcadm.CreateUsers(m, MINIOADM)

	return nil
}

// WaitFor waits until the minio server is up and running
func (m *MinioAdm) WaitFor() error {
	return containerutils.WaitForContainerReadiness(m.Service.Container.Name, 5, 30)
}

// GenerateNginxConf generates the nginx reverse proxy configuration for the minio server
func (m *MinioAdm) GenerateNginxConf() string {
	return fmt.Sprintf(`# MinIO Web UI
location /%s/ {
#   rewrite ^/%s/(.*) /$1 break;

	proxy_set_header Host $http_host;
	proxy_set_header X-Real-IP $remote_addr;
	proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
	proxy_set_header X-Forwarded-Proto $scheme;
	proxy_set_header X-NginX-Proxy true;

	real_ip_header X-Real-IP;

	proxy_connect_timeout 300;

	proxy_http_version 1.1;
	proxy_set_header Upgrade $http_upgrade;
	proxy_set_header Connection "upgrade";

	chunked_transfer_encoding off;

	proxy_pass http://%s:9001/;
}
# MinIO API
location /%s-api/ {
#   rewrite ^/%s-api/(.*) /$1 break;

	proxy_set_header Host $http_host;
	proxy_set_header X-Real-IP $remote_addr;
	proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
	proxy_set_header X-Forwarded-Proto $scheme;

	proxy_connect_timeout 300;

	proxy_http_version 1.1;
	proxy_set_header Connection "";
	chunked_transfer_encoding off;

	proxy_pass http://%s:9000/;
}`, m.Service.Name, m.Service.Name, m.Service.Container.Name, m.Service.Name, m.Service.Name, m.Service.Container.Name)
}

// GetService returns the service object from the configuration
func (m *MinioAdm) GetService() config.Service {
	return m.Service
}

func (m *MinioAdm) GetServiceName() string {
	return m.Service.Name
}

func (m *MinioAdm) GetServiceAdmName() string {
	return MINIOADM
}

func (m *MinioAdm) Cleanup() ([]string, []string) {
	return []string{}, []string{}
}
