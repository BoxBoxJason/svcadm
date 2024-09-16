package utils

import (
	"fmt"
	"os/exec"
)

// CheckOperatorInstalled checks if the container operator is installed
func CheckOperatorInstalled(operator string) bool {
	_, err := exec.LookPath(operator)
	return err == nil
}

// CheckOperatorRunning checks if the container operator is running
func CheckOperatorRunning(operator string) bool {
	cmd := exec.Command(operator, "info")
	return cmd.Run() == nil
}

// StartContainer starts a container using the container operator with the specified parameters
func StartContainer(operator string, container string, image string, network string, volumes map[string]string, ports []string, env map[string]string, command string) error {
	volumes_string := ""
	for volume, path := range volumes {
		volumes_string += fmt.Sprintf("-v %s:%s ", volume, path)
	}
	ports_string := ""
	for _, port := range ports {
		ports_string += fmt.Sprintf("-p %s ", port)
	}
	env_string := ""
	for key, value := range env {
		env_string += fmt.Sprintf("-e %s=%s ", key, value)
	}
	cmd := exec.Command(operator, "run", "-d", "--name", container, "--network", network, volumes_string, ports_string, env_string, image, command)
	return cmd.Run()
}

// StopContainer stops a container using the container operator
func StopContainer(operator string, container string) error {
	cmd := exec.Command(operator, "stop", container)
	return cmd.Run()
}

// RemoveContainer removes a container using the container operator
func RemoveContainer(operator string, container string) error {
	cmd := exec.Command(operator, "rm", "-f", container)
	return cmd.Run()
}

// FetchContainerStatus fetches the status of a container using the container operator
func FetchContainerStatus(operator string, container string) (string, error) {
	cmd := exec.Command(operator, "inspect", container, "--format", "{{.State.Status}}")
	output, err := cmd.Output()
	return string(output), err
}

// FetchContainerLogs fetches the logs of a container using the container operator
func FetchContainerLogs(operator string, container string) (string, error) {
	cmd := exec.Command(operator, "logs", container)
	output, err := cmd.Output()
	return string(output), err
}

// CreateVolume creates a volume using the container operator
func CreateVolume(operator string, volume_name string) error {
	cmd := exec.Command(operator, "volume", "create", volume_name)
	return cmd.Run()
}

// RemoveVolume removes a volume using the container operator
func RemoveVolume(operator string, volume_name string) error {
	cmd := exec.Command(operator, "volume", "rm", volume_name)
	return cmd.Run()
}

// RunContainerCommand runs a command in an existing container using the container operator
func RunContainerCommand(operator string, container string, command string) error {
	cmd := exec.Command(operator, "exec", container, command)
	return cmd.Run()
}
