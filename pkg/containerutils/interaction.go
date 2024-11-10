package containerutils

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/boxboxjason/svcadm/pkg/fileutils"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/pkg/stdcopy"
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

	resp, err := cli.ContainerExecCreate(ctx, container_name, container.ExecOptions{
		Cmd:          command,
		AttachStderr: true,
	})
	if err != nil {
		return err
	}

	err = cli.ContainerExecStart(ctx, resp.ID, container.ExecStartOptions{})
	if err != nil {
		return err
	}

	inspect_resp, err := cli.ContainerExecInspect(ctx, resp.ID)
	if err != nil {
		return err
	}

	if inspect_resp.ExitCode != 0 {
		return fmt.Errorf("command %v in container %s exited with code %d", command, container_name, inspect_resp.ExitCode)
	}

	return nil
}

func RunContainerCommandWithOutput(containerName string, command ...string) ([]byte, error) {
	ctx := context.Background()

	// Create the exec instance with stdout and stderr attached
	resp, err := cli.ContainerExecCreate(ctx, containerName, container.ExecOptions{
		Cmd:          command,
		AttachStdout: true,
		AttachStderr: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create exec instance: %w", err)
	}

	// Attach to the exec instance to capture output
	exec_start_resp, err := cli.ContainerExecAttach(ctx, resp.ID, container.ExecStartOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to attach to exec instance: %w", err)
	}
	defer exec_start_resp.Close()

	// Buffers to capture stdout and stderr
	var stdoutBuf, stderrBuf bytes.Buffer

	// Copy the output from exec_start_resp.Reader to the buffers
	outputDone := make(chan error)
	go func() {
		_, err := stdcopy.StdCopy(&stdoutBuf, &stderrBuf, exec_start_resp.Reader)
		outputDone <- err
	}()

	// Wait for the output copying to complete
	if err := <-outputDone; err != nil {
		return nil, fmt.Errorf("failed to capture output: %w", err)
	}

	// Inspect the exec instance result to check the exit code
	inspect_resp, err := cli.ContainerExecInspect(ctx, resp.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to inspect exec instance: %w", err)
	}

	// Check for non-zero exit code and return an error if found
	if inspect_resp.ExitCode != 0 {
		return stdoutBuf.Bytes(), fmt.Errorf("command %v in container %s exited with code %d: %s", command, containerName, inspect_resp.ExitCode, stderrBuf.String())
	}

	return stdoutBuf.Bytes(), nil
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
		split_env := strings.Split(env, "=")
		if split_env[0] == variable {
			return strings.Join(split_env[1:], "="), nil
		}
	}
	return "", fmt.Errorf("environment variable %s not found in container %s", variable, container_name)
}
