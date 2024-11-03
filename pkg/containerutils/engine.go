package containerutils

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"

	"github.com/boxboxjason/svcadm/pkg/logger"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
)

const (
	docker = "docker"
	podman = "podman"
)

var (
	// Socket paths
	podman_socket_path = fmt.Sprintf("/run/user/%d/podman/podman.sock", os.Getuid())
	docker_socket_path = "/var/run/docker.sock"
	cli                *client.Client
	mu_CLI             = &sync.RWMutex{}
	// Container engine
	container_engine    string
	mu_CONTAINER_ENGINE = &sync.RWMutex{}
	// Container network
	container_network string
	mu_NETWORK        = &sync.RWMutex{}
)

// ========== Container engine ==========
func SetContainerEngine(op string) error {
	mu_CONTAINER_ENGINE.Lock()
	if op == docker || op == podman {
		container_engine = op
		logger.Info("Container engine set to", container_engine)
	} else {
		return errors.New("invalid container engine " + op)
	}
	mu_CONTAINER_ENGINE.Unlock()
	return initContainerClient()
}

func GetContainerEngine() string {
	mu_CONTAINER_ENGINE.RLock()
	defer mu_CONTAINER_ENGINE.RUnlock()
	return container_engine
}

// ========== Container network ==========
func SetContainersNetwork(network string) error {
	// Lock the network and cli mutexes
	mu_NETWORK.Lock()
	defer mu_NETWORK.Unlock()
	mu_CLI.RLock()
	defer mu_CLI.RUnlock()

	// Check if the network name is valid
	if !VALID_NETWORK_NAME.MatchString(network) {
		return errors.New("invalid network name " + network)
	} else {
		err := CreateContainerNetwork(container_network, "bridge")
		if err != nil {
			return err
		}
		container_network = network
		logger.Info("Container network set to", container_network)
	}

	return nil
}

func GetContainersNetwork() string {
	mu_NETWORK.RLock()
	defer mu_NETWORK.RUnlock()
	return container_network
}

// Check if the network exists
func networkExists(network_name string) (bool, error) {
	ctx := context.Background()
	filter := filters.NewArgs()
	filter.Add("name", network_name)

	networks, err := cli.NetworkList(ctx, network.ListOptions{Filters: filter})
	if err != nil {
		return false, err
	}

	for _, net := range networks {
		if net.Name == network_name {
			return true, nil
		}
	}
	return false, nil
}

// CreateContainerNetwork creates a network if it doesn't exist
func CreateContainerNetwork(network_name, driver string) error {
	exists, err := networkExists(network_name)
	if err != nil {
		return fmt.Errorf("error checking network existence: %w", err)
	}

	if !exists {
		ctx := context.Background()
		_, err := cli.NetworkCreate(ctx, network_name, network.CreateOptions{
			Driver: driver,
		})
		if err != nil {
			return fmt.Errorf("error creating network %s: %w", network_name, err)
		}
		logger.Info("Network", network_name, "created successfully")
	} else {
		logger.Info("Network", network_name, "already exists, using existing network")
	}

	return nil
}

// ========== Container client ==========
func initContainerClient() error {
	mu_CLI.Lock()
	defer mu_CLI.Unlock()
	engine := GetContainerEngine()

	// Check if DOCKER_HOST is set and if not set it to the default socket path
	DOCKER_HOST := os.Getenv("DOCKER_HOST")
	if DOCKER_HOST == "" {
		if engine == podman {
			if !checkPodmanService() {
				return fmt.Errorf("no socket found at %s please run 'podman system service --time=0 unix://%s'", podman_socket_path, podman_socket_path)
			}
			os.Setenv("DOCKER_HOST", "unix://"+podman_socket_path)
		} else {
			os.Setenv("DOCKER_HOST", "unix://"+docker_socket_path)
		}
	}

	var err error
	cli, err = client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	return err
}

func checkPodmanService() bool {
	_, err := os.Stat(podman_socket_path)
	return err == nil
}
