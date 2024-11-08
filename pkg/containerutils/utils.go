package containerutils

import (
	"context"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"

	"github.com/boxboxjason/svcadm/pkg/logger"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/go-connections/nat"
)

const (
	ALPHANUMERICS_REGEX = `^[a-zA-Z0-9_-]+$`
)

var (
	VALID_RESTART_POLICY = regexp.MustCompile(`^(always|no|unless-stopped|on-failure:\d+)$`)
	VALID_VOLUME_NAME    = VALID_CONTAINER_NAME
	VALID_CONTAINER_NAME = regexp.MustCompile(ALPHANUMERICS_REGEX)
	VALID_NETWORK_NAME   = VALID_CONTAINER_NAME
)

func formatEnvVariables(env map[string]string) []string {
	var env_variables []string
	for variable, value := range env {
		env_variables = append(env_variables, variable+"="+value)
	}
	return env_variables
}

func formatVolumes(volumes map[string]string) []string {
	var volume_bindings []string
	for host_path, container_path := range volumes {
		volume_bindings = append(volume_bindings, host_path+":"+container_path)
	}
	return volume_bindings
}

func formatPortBindings(ports map[int]int) nat.PortMap {
	port_bindings := nat.PortMap{}

	for hostPort, containerPort := range ports {
		container_port_key := nat.Port(fmt.Sprintf("%d/tcp", containerPort)) // Assuming TCP protocol, might need to be more flexible
		host_port_binding := nat.PortBinding{HostPort: fmt.Sprintf("%d", hostPort)}

		port_bindings[container_port_key] = append(port_bindings[container_port_key], host_port_binding)
	}

	return port_bindings
}

// formatRestartPolicy converts a string to a container.RestartPolicy
// Supposes that the input string is a valid restart policy
func formatRestartPolicy(policy string) container.RestartPolicy {
	policy_split := strings.Split(policy, ":")
	var restart_policy container.RestartPolicy
	if len(policy_split) == 1 {
		restart_policy.Name = container.RestartPolicyMode(policy_split[0])
	} else if len(policy_split) == 2 {
		max_retries := policy_split[1]
		max_retries_int, err := strconv.Atoi(max_retries)
		if err != nil {
			panic(err)
		}
		restart_policy.Name = container.RestartPolicyMode(policy_split[0])
		restart_policy.MaximumRetryCount = max_retries_int
	} else {
		panic("Invalid restart policy")
	}
	return restart_policy
}

func PullImage(image_name string) error {
	ctx := context.Background()

	// Check if the image already exists
	exists, err := ImageAlreadyExists(image_name)
	if err != nil {
		return err
	} else if exists {
		logger.Debug("Image", image_name, "is already present")
		return nil
	}

	logger.Info("Pulling image", image_name)
	out, err := cli.ImagePull(ctx, image_name, image.PullOptions{})
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err := io.Copy(io.Discard, out); err != nil {
		return err
	}

	logger.Info("Image", image_name, "pulled successfully")
	return nil
}

func ImageAlreadyExists(image_name string) (bool, error) {
	ctx := context.Background()
	filter := filters.NewArgs()
	filter.Add("reference", image_name)

	images, err := cli.ImageList(ctx, image.ListOptions{Filters: filter})
	if err != nil {
		return false, err
	}
	for _, image := range images {
		for _, tag := range image.RepoTags {
			if tag == image_name {
				return true, nil
			}
		}
	}
	return false, nil
}
