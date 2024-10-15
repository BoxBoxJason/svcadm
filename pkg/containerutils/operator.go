package containerutils

import (
	"fmt"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/boxboxjason/svcadm/pkg/logger"
)

const (
	docker = "docker"
	podman = "podman"
)

var (
	container_operator    string
	mu_CONTAINER_OPERATOR sync.RWMutex
	containers_network    string
	mu_SVCADM_NETWORK     sync.RWMutex
)

// SetContainerOperator sets the container operator to be used
func SetContainerOperator(operator string) error {
	mu_CONTAINER_OPERATOR.Lock()
	defer mu_CONTAINER_OPERATOR.Unlock()
	if operator != docker && operator != podman {
		return fmt.Errorf("invalid container operator %s", operator)
	}
	container_operator = operator
	return nil
}

// GetContainerOperator gets the container operator to be used
func GetContainerOperator() string {
	mu_CONTAINER_OPERATOR.RLock()
	defer mu_CONTAINER_OPERATOR.RUnlock()
	if container_operator == "" {
		container_operator = docker
	}
	return container_operator
}

// SetContainersNetwork sets the network to be used by the containers
func SetContainersNetwork(network string) {
	mu_SVCADM_NETWORK.Lock()
	defer mu_SVCADM_NETWORK.Unlock()
	containers_network = network
}

// GetContainersNetwork gets the network to be used by the containers
func GetContainersNetwork() string {
	mu_SVCADM_NETWORK.RLock()
	defer mu_SVCADM_NETWORK.RUnlock()
	return containers_network
}

// CheckOperatorInstalled checks if the container operator is installed
func CheckOperatorInstalled() bool {
	operator := GetContainerOperator()
	_, err := exec.LookPath(operator)
	return err == nil
}

// CheckOperatorRunning checks if the container operator is running
func CheckOperatorRunning() bool {
	operator := GetContainerOperator()
	return exec.Command(operator, "info").Run() == nil
}

// StartContainer starts a container using the container operator with the specified parameters
func StartContainer(container string, image string, volumes map[string]string, ports map[int]int, env map[string]string, restart_policy string, container_args []string, command []string) error {
	args := []string{
		"run",
		"-d",
		"--name", container,
		"--pull=missing",
		"--network", containers_network,
		fmt.Sprintf("--restart=%s", restart_policy),
	}

	// Attach container arguments
	args = append(args, container_args...)

	// Attach volumes
	for volume, path := range volumes {
		args = append(args, "-v", fmt.Sprintf("%s:%s", volume, path)) // Each flag and value separated
	}

	// Attach ports
	for host_port, container_port := range ports {
		args = append(args, "-p", fmt.Sprintf("%d:%d", host_port, container_port))
	}

	// Attach environment variables
	for key, value := range env {
		args = append(args, "-e", fmt.Sprintf("%s=%s", key, value)) // Separate -e and key=value
	}

	// Add the image and optional command
	args = append(args, image)
	if len(command) > 0 {
		args = append(args, command...) // Safely split the command
	}

	operator := GetContainerOperator()

	return exec.Command(operator, args...).Run()
}

// StopContainer stops a container using the container operator
func StopContainer(container string) error {
	operator := GetContainerOperator()
	return exec.Command(operator, "stop", container).Run()
}

// RemoveContainer removes a container using the container operator
func RemoveContainer(container string) error {
	operator := GetContainerOperator()
	return exec.Command(operator, "rm", "-f", container).Run()
}

// FetchContainerStatus fetches the status of a container using the container operator
func FetchContainerStatus(container string) (string, error) {
	operator := GetContainerOperator()
	cmd := exec.Command(operator, "inspect", container, "--format", "{{.State.Status}}")
	output, err := cmd.Output()
	return string(output), err
}

// WaitForContainerReadiness waits for a container to be ready using the container operator
func WaitForContainerReadiness(container string, retry_interval int, max_retries int) error {
	for max_retries > 0 {
		status, err := FetchContainerStatus(container)
		if err != nil {
			return err
		}
		if strings.Contains(status, "running") {
			logger.Info(container, "container is ready")
			return nil
		}
		logger.Debug(container, "container is not ready, retrying in", retry_interval, "seconds")
		max_retries--
		time.Sleep(time.Duration(retry_interval) * time.Second)
	}
	return fmt.Errorf("timed out waiting for container %s to be ready", container)
}

// FetchContainerLogs fetches the logs of a container using the container operator
func FetchContainerLogs(container string) (string, error) {
	operator := GetContainerOperator()
	cmd := exec.Command(operator, "logs", container)
	output, err := cmd.Output()
	return string(output), err
}

// CreateVolume creates a volume using the container operator
func CreateVolume(volume_name string) error {
	operator := GetContainerOperator()
	return exec.Command(operator, "volume", "create", volume_name).Run()
}

// RemoveVolume removes a volume using the container operator
func RemoveVolume(volume_name string) error {
	operator := GetContainerOperator()
	cmd := exec.Command(operator, "volume", "rm", volume_name)
	return cmd.Run()
}

// RunContainerCommand runs a command in an existing container using the container operator
func RunContainerCommand(container string, command ...string) error {
	operator := GetContainerOperator()
	command = append([]string{"exec", container}, command...)
	cmd := exec.Command(operator, command...)

	// TODO: REMOVE THIS DEBUG PRINTER
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("Command error: %v\n", err)
		fmt.Printf("Command output: %s\n", string(output))
	}
	return err
}

// CopyContainerFile copies runs a copy command from a container towards the host machine
func CopyContainerFile(container string, source string, destination string) error {
	operator := GetContainerOperator()
	return exec.Command(operator, "cp", fmt.Sprintf("%s:%s", container, source), destination).Run()
}

// RunContainerCommandWithOutput runs a command in an existing container using the container operator and returns the output as a string
func RunContainerCommandWithOutput(container string, command ...string) ([]byte, error) {
	operator := GetContainerOperator()
	command = append([]string{"exec", container}, command...)
	cmd := exec.Command(operator, command...)

	// TODO: REMOVE THIS DEBUG PRINTER
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("Command error: %v\n", err)
		fmt.Printf("Command output: %s\n", string(output))
	}
	return output, err
}

// CheckVolumeExists checks if a container operator volume exists
func CheckVolumeExists(volume string) bool {
	operator := GetContainerOperator()
	return exec.Command(operator, "volume", "inspect", volume).Run() == nil
}

// CheckContainerExists checks if a container exists
func CheckContainerExists(container string) bool {
	operator := GetContainerOperator()
	return exec.Command(operator, "inspect", container).Run() == nil
}

// ContainerNetworkExists checks if a container network exists
func ContainerNetworkExists(network string) bool {
	operator := GetContainerOperator()
	return exec.Command(operator, "network", "inspect", network).Run() == nil
}

// CreateNetwork creates a container network
func CreateNetwork(network string, driver string) error {
	if !ContainerNetworkExists(network) {
		operator := GetContainerOperator()
		logger.Debug("Creating network ", network)
		return exec.Command(operator, "network", "create", "--driver", driver, network).Run()
	}
	return nil
}

// RemoveNetwork removes a container network
func RemoveNetwork(network string) error {
	operator := GetContainerOperator()
	return exec.Command(operator, "network", "rm", network).Run()
}

// ResumeContainer resumes a container using the container operator
func ResumeContainer(container string) error {
	operator := GetContainerOperator()
	return exec.Command(operator, "start", container).Run()
}

// GetContainerEnvVariable gets an environment variable from a container
func GetContainerEnvVariable(container string, variable string) (string, error) {
	output, err := RunContainerCommandWithOutput(container, "printenv", variable)
	return strings.Trim(string(output), "\n"), err
}
