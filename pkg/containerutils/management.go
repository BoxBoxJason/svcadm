package containerutils

import (
	"context"
	"fmt"
	"time"

	"github.com/boxboxjason/svcadm/pkg/logger"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/volume"
	"github.com/docker/docker/client"
)

// ResumeContainer resumes a container by its name
func ResumeContainer(container_name string) error {
	ctx := context.Background()
	return cli.ContainerStart(ctx, container_name, container.StartOptions{})
}

// RemoveContainer deletes a container by its name, and optionally deletes its volumes
func RemoveContainer(container_name string, delete_volumes bool) error {
	ctx := context.Background()
	return cli.ContainerRemove(ctx, container_name, container.RemoveOptions{Force: true, RemoveVolumes: delete_volumes})
}

// FetchContainerStatus returns the status of a container by its name
func FetchContainerStatus(container_name string) (string, error) {
	ctx := context.Background()
	containerJSON, err := cli.ContainerInspect(ctx, container_name)
	if err != nil {
		return "", err
	}
	return containerJSON.State.Status, nil
}

// WaitForContainerReadiness waits for a container to be in a running state,
// with a configurable retry interval and maximum number of retries
func WaitForContainerReadiness(container_name string, retry_interval int, max_retries int) error {
	time.Sleep(time.Duration(retry_interval) * time.Second)
	for max_retries > 0 {
		status, err := FetchContainerStatus(container_name)
		if err != nil {
			return err
		}
		if status == "running" {
			logger.Info(container_name, "container is ready")
			return nil
		}
		time.Sleep(time.Duration(retry_interval) * time.Second)
		max_retries--
	}
	return fmt.Errorf("timed out waiting for container %s to be ready", container_name)
}

// CreateVolume creates a volume with a given name, driver, and labels
func CreateVolume(volume_name string, driver string, labels map[string]string) error {
	ctx := context.Background()
	_, err := cli.VolumeCreate(ctx, volume.CreateOptions{Name: volume_name, Driver: driver, Labels: labels})
	return err
}

// RemoveVolume deletes a volume by its name
// WARNING: force is set to true to remove the volume even if it is in use
func RemoveVolume(volume_name string, force bool) error {
	ctx := context.Background()
	return cli.VolumeRemove(ctx, volume_name, force)
}

// CheckContainerExists checks if a container exists by its name
func CheckContainerExists(container_name string) (bool, error) {
	ctx := context.Background()
	_, err := cli.ContainerInspect(ctx, container_name)
	if err != nil {
		if client.IsErrNotFound(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// CheckContainerRunsWithImage checks if a container is running with a specific image
func CheckContainerRunsWithImage(container_name, image string) (bool, error) {
	ctx := context.Background()
	containerJSON, err := cli.ContainerInspect(ctx, container_name)
	if err != nil {
		return false, err
	}
	return containerJSON.Config.Image == image, nil
}

// CheckVolumeExists checks if a volume exists by its name
func CheckVolumeExists(volume_name string) bool {
	ctx := context.Background()
	filters := filters.NewArgs()
	filters.Add("name", volume_name)

	volumes, err := cli.VolumeList(ctx, volume.ListOptions{Filters: filters})
	if err != nil {
		return false
	}
	for _, volume := range volumes.Volumes {
		if volume.Name == volume_name {
			return true
		}
	}
	return false
}
