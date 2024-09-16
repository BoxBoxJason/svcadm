package minioadm

import (
	"fmt"
	"os/exec"
	"time"

	"github.com/boxboxjason/svcadm/internal/config"
	"github.com/boxboxjason/svcadm/pkg/utils"
)

// CreateUser creates a user in the minio server
func CreateUser(operator string, minio_container_name string, user string, password string) error {
	err := utils.RunContainerCommand(operator, minio_container_name, fmt.Sprintf("mc admin user add %s %s %s", minio_container_name, user, password))
	if err != nil {
		return err
	}
	return utils.RunContainerCommand(operator, minio_container_name, fmt.Sprintf("mc admin policy attach %s readwrite user=%s", minio_container_name, user))
}

// CreateAdminUser creates an admin user in the minio server
func CreateAdminUser(operator string, minio_container_name string, user string, password string) error {
	err := utils.RunContainerCommand(operator, minio_container_name, fmt.Sprintf("mc admin user add %s %s %s --admin", minio_container_name, user, password))
	if err != nil {
		return err
	}
	return utils.RunContainerCommand(operator, minio_container_name, fmt.Sprintf("mc admin policy attach %s consoleAdmin user=%s", minio_container_name, user))
}

// BackupBucket creates a backup of a minio bucket on the operator's machine
func BackupBucket(operator string, minio_container_name string, bucket_name string, backup_name string) error {
	err := utils.RunContainerCommand(operator, minio_container_name, fmt.Sprintf("tar -cJf /tmp/%s.tar.xz /data/%s", bucket_name, bucket_name))
	if err != nil {
		return err
	}
	cmd := exec.Command("docker", "cp", fmt.Sprintf("%s:/tmp/%s.tar.xz %s", minio_container_name, bucket_name, backup_name))
	err = cmd.Run()
	if err != nil {
		return err
	}
	return utils.RunContainerCommand(operator, minio_container_name, fmt.Sprintf("rm /tmp/%s.tar.xz", bucket_name))
}

// BackupAllBuckets creates a backup of all minio buckets on the operator's machine
func BackupAllBuckets(operator string, minio_container_name string, backup_name string) error {
	err := utils.RunContainerCommand(operator, minio_container_name, "tar -cJf /tmp/all.tar.xz /data")
	if err != nil {
		return err
	}
	cmd := exec.Command("docker", "cp", fmt.Sprintf("%s:/tmp/all.tar.xz %s", minio_container_name, backup_name))
	err = cmd.Run()
	if err != nil {
		return err
	}
	return utils.RunContainerCommand(operator, minio_container_name, "rm /tmp/all.tar.xz")
}

// CreateBucket creates a new bucket in the minio server
func CreateBucket(operator string, minio_container_name string, bucket_name string) error {
	return utils.RunContainerCommand(operator, minio_container_name, fmt.Sprintf("mc mb %s/%s", minio_container_name, bucket_name))
}

// DeleteBucket deletes a bucket from the minio server
func DeleteBucket(operator string, minio_container_name string, bucket_name string) error {
	return utils.RunContainerCommand(operator, minio_container_name, fmt.Sprintf("mc rb %s/%s --force", minio_container_name, bucket_name))
}

func PreInitMinio(operator string, minio_service *config.Service) (map[string]string, map[string]string, error) {
	extended_env := make(map[string]string)
	extended_volumes := make(map[string]string)
	var err error

	// Set the environment variables for the minio server
	root_user := minio_service.Container.Env["MINIO_ROOT_USER"]
	if root_user == "" {
		root_user = "svcadm"
	}
	root_password := minio_service.Container.Env["MINIO_ROOT_PASSWORD"]
	if root_password == "" {
		root_password, err = utils.GenerateRandomPassword(32)
		if err != nil {
			return nil, nil, err
		}
	}
	extended_env["MINIO_ROOT_USER"] = root_user
	extended_env["MINIO_ROOT_PASSWORD"] = root_password

	// Add the volumes for the minio server
	extended_volumes["minio-data"] = "/data"

	return extended_env, extended_volumes, nil
}

// PostInitMinio runs the post init steps for the minio server
func PostInitMinio(operator string, minio_container_name string, minio_root_user string, minio_root_password string, users *config.Users) error {
	// Wait for the minio server to be ready
	err := waitForMinio(operator, minio_container_name)
	if err != nil {
		return err
	}

	// Fix the default aliases because its not done by default ???????????
	for _, alias := range []string{"local", "gcs", "s3", "play"} {
		_ = utils.RunContainerCommand(operator, minio_container_name, fmt.Sprintf("mc alias remove %s", alias))
	}
	err = utils.RunContainerCommand(operator, minio_container_name, fmt.Sprintf("mc alias set %s http://localhost:9000 %s %s", minio_container_name, minio_root_user, minio_root_password))
	if err != nil {
		return err
	}

	// Create the users
	for _, user := range users.Users {
		err = CreateUser(operator, minio_container_name, user.Username, user.Password)
		if err != nil {
			return err
		}
	}
	// Create the admin users
	for _, user := range users.Admins {
		err = CreateAdminUser(operator, minio_container_name, user.Username, user.Password)
		if err != nil {
			return err
		}
	}

	return nil
}

// waitForMinio waits until the minio server is up and running
func waitForMinio(container_operator string, minio_container_name string) error {
	ready := false
	readiness_command := "mc ready " + minio_container_name
	max_retries := 30
	const retry_interval = 5
	for !ready && max_retries > 0 {
		err := utils.RunContainerCommand(container_operator, minio_container_name, readiness_command)
		if err == nil {
			return nil
		}
		max_retries--
		time.Sleep(retry_interval * time.Second)
	}
	return fmt.Errorf("minio server is not ready after %d retries", max_retries)
}

// GenerateMinIONginxConf generates the nginx reverse proxy configuration for the minio server
func GenerateMinIONginxConf(minio_service *config.Service) string {
	return fmt.Sprintf(`# MinIO Web UI
location /minio/ {
	rewrite ^/minio/(.*) /$1 break;

	proxy_set_header Host $http_host;
	proxy_set_header X-Real-IP $remote_addr;
	proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
	proxy_set_header X-Forwarded-Proto $scheme;
	proxy_set_header X-NginX-Proxy true;

	proxy_set_header Accept-Encoding "";
	proxy_http_version 1.1;
	proxy_set_header Connection "";

	proxy_buffering off;

	proxy_pass http://%s:9001/;
}
# MinIO API
location /minio-api/ {
	proxy_pass http://%s:9000;
	proxy_set_header Host $http_host;
	proxy_set_header X-Real-IP $remote_addr;
	proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
	proxy_set_header X-Forwarded-Proto $scheme;

	proxy_http_version 1.1;
	proxy_set_header Connection "";

	proxy_buffering off;
}`, minio_service.Container.Name, minio_service.Container.Name)
}
