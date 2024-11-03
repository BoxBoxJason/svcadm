package containerutils

import (
	"context"
	"fmt"
	"io"

	"github.com/boxboxjason/svcadm/pkg/fileutils"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
)

func CreateContainer(container_name string, image string, labels map[string]string, volumes map[string]string, ports map[int]int, env map[string]string, restart_policy string, cap_add []string, cmd []string) error {
	ctx := context.Background()

	// Container configuration
	container_config := &container.Config{
		Image:  image,
		Env:    formatEnvVariables(env),
		Cmd:    cmd,
		Labels: labels,
	}

	// Host configuration
	host_config := &container.HostConfig{
		RestartPolicy: formatRestartPolicy(restart_policy),
		Binds:         formatVolumes(volumes),
		PortBindings:  formatPortBindings(ports),
		CapAdd:        cap_add,
	}

	// Network configuration
	networkConfig := &network.NetworkingConfig{
		EndpointsConfig: map[string]*network.EndpointSettings{
			GetContainersNetwork(): {},
		},
	}

	// Pull image
	err := PullImage(image)
	if err != nil {
		return err
	}

	// Create and start container
	resp, err := cli.ContainerCreate(ctx, container_config, host_config, networkConfig, nil, container_name)
	if err != nil {
		return err
	}
	return cli.ContainerStart(ctx, resp.ID, container.StartOptions{})
}

func StopContainer(container_name string) error {
	ctx := context.Background()
	return cli.ContainerStop(ctx, container_name, container.StopOptions{})
}

func RunContainerCommand(container_name string, command ...string) error {
	ctx := context.Background()

	resp, err := cli.ContainerExecCreate(ctx, container_name, container.ExecOptions{Cmd: command, AttachStderr: true})
	if err != nil {
		return err
	}
	err = cli.ContainerExecStart(ctx, resp.ID, container.ExecStartOptions{})
	// TODO: remove debug statement
	if err != nil {
		fmt.Println(err)
	}
	return err
}

func RunContainerCommandWithOutput(container_name string, command ...string) ([]byte, error) {
	ctx := context.Background()

	resp, err := cli.ContainerExecCreate(ctx, container_name, container.ExecOptions{Cmd: command})
	if err != nil {
		return []byte{}, err
	}
	out, err := cli.ContainerExecAttach(ctx, resp.ID, container.ExecStartOptions{})
	if err != nil {
		return []byte{}, err
	}
	defer out.Close()

	output, err := io.ReadAll(out.Reader)
	return output, err
}

// FetchContainerLogs returns the logs of a container by its name
// returns the entire log as a string
func FetchContainerLogs(container_name string) (string, error) {
	ctx := context.Background()
	out, err := cli.ContainerLogs(ctx, container_name, container.LogsOptions{ShowStdout: true, ShowStderr: true})
	if err != nil {
		return "", err
	}
	defer out.Close()

	logs, err := io.ReadAll(out)
	return string(logs), err
}

func CopyContainerResource(container_name string, source string, destination string) error {
	ctx := context.Background()

	reader, _, err := cli.CopyFromContainer(ctx, container_name, source)
	if err != nil {
		return err
	}
	defer reader.Close()

	return fileutils.ExtractTarToDestination(reader, destination)
}

func GetContainerEnvVariable(container_name string, variable string) (string, error) {
	ctx := context.Background()

	inspect, err := cli.ContainerInspect(ctx, container_name)
	if err != nil {
		return "", err
	}

	for _, env := range inspect.Config.Env {
		if env == variable {
			return env, nil
		}
	}
	return "", nil
}
