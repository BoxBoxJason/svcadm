package services

import (
	"fmt"
	"sync"

	"github.com/boxboxjason/svcadm/internal/config"
	"github.com/boxboxjason/svcadm/internal/services/gitlabadm"
	"github.com/boxboxjason/svcadm/internal/services/mattermostadm"
	"github.com/boxboxjason/svcadm/internal/services/minioadm"
	"github.com/boxboxjason/svcadm/internal/services/nginxadm"
	"github.com/boxboxjason/svcadm/internal/services/psqladm"
	"github.com/boxboxjason/svcadm/internal/services/sonaradm"
	"github.com/boxboxjason/svcadm/internal/services/svcadm"
	"github.com/boxboxjason/svcadm/internal/services/trivyadm"
	"github.com/boxboxjason/svcadm/internal/services/vaultadm"
	"github.com/boxboxjason/svcadm/pkg/containerutils"
	"github.com/boxboxjason/svcadm/pkg/fileutils"
	"github.com/boxboxjason/svcadm/pkg/logger"
)

func BackupServices() error {
	for _, service := range config.GetConfiguration().Services {
		if service.Enabled {
			err := Backup(service.Name, service.Backup.Location)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// Backup creates a backup of the service
func Backup(service_name string, backup_path string) error {
	service := config.GetService(service_name)
	if service.Name != service_name {
		return fmt.Errorf("service %s not found", service_name)
	}
	service_adm, err := getServiceAdm(&service)
	if err != nil {
		return err
	}
	return service_adm.Backup(backup_path)
}

// getServiceAdm returns the service adm for the given service
// Can input either a service name, a service struct or a pointer to a service struct
func getServiceAdm(service_interface interface{}) (svcadm.ServiceAdm, error) {
	var service_name string
	var service *config.Service
	if service_interface == nil {
		return nil, fmt.Errorf("service not found")
	} else if service_attempt, ok := service_interface.(config.Service); ok {
		service_name = service_attempt.Name
		service = &service_attempt
	} else if service_attempt, ok := service_interface.(*config.Service); ok {
		service_name = service_attempt.Name
		service = service_attempt
	} else if service_name_string, ok := service_interface.(string); ok {
		service_name = service_name_string
		tmp_service := config.GetService(service_name)
		service = &tmp_service
	} else {
		return nil, fmt.Errorf("service not found")
	}

	var serviceAdm svcadm.ServiceAdm
	switch service_name {
	case "sonarqube":
		serviceAdm = &sonaradm.SonarAdm{Service: *service}
	case "postgresql":
		serviceAdm = &psqladm.PsqlAdm{Service: *service}
	case "gitlab":
		serviceAdm = &gitlabadm.GitLabAdm{Service: *service}
	case "minio":
		serviceAdm = &minioadm.MinioAdm{Service: *service}
	case "mattermost":
		serviceAdm = &mattermostadm.MattermostAdm{Service: *service}
	case "nginx":
		serviceAdm = &nginxadm.NginxAdm{Service: *service}
	case "vault":
		serviceAdm = &vaultadm.VaultAdm{Service: *service}
	case "trivy":
		serviceAdm = &trivyadm.TrivyAdm{Service: *service}
	default:
		return nil, fmt.Errorf("service %s not found", service.Name)
	}
	return serviceAdm, nil
}

// CleanupService deletes the service container, its volumes and all ressources associated with it, except the backups
func CleanupService(service *config.Service) error {

	service_adm, err := getServiceAdm(service)
	if err != nil {
		logger.Error("could not get service adm for", service.Name)
		return err
	}

	SERVICE_ADM_PREFIX := service_adm.GetServiceAdmName() + ":"

	container_name := service.Container.Name
	default_volumes, default_directories := service_adm.Cleanup()
	volumes := config.GetServiceVolumes(service)

	for _, volume := range default_volumes {
		volumes[volume] = ""
	}

	container_exists, err := containerutils.CheckContainerExists(container_name)
	if err != nil {
		logger.Error(SERVICE_ADM_PREFIX, "could not check if container exists", container_name)
	} else if container_exists {
		logger.Debug(SERVICE_ADM_PREFIX, "stopping and deleting container", container_name)
		_ = containerutils.StopContainer(container_name)
		err := containerutils.RemoveContainer(container_name, true)
		if err != nil {
			logger.Error(SERVICE_ADM_PREFIX, "could not remove container", container_name)
		}
	}

	for volume := range volumes {
		if containerutils.CheckVolumeExists(volume) {
			logger.Debug(SERVICE_ADM_PREFIX, "deleting volume", volume)
			err := containerutils.RemoveVolume(volume, true)
			if err != nil {
				logger.Error(SERVICE_ADM_PREFIX, "could not remove volume", volume)
			}
		}
	}

	for _, path := range default_directories {
		logger.Debug(SERVICE_ADM_PREFIX, "deleting", path)
		err := fileutils.DeleteDirectory(path)
		if err != nil {
			logger.Error(SERVICE_ADM_PREFIX, "could not remove directory", path)
		}
	}

	return nil
}

// CleanupServices deletes all services containers, their volumes and all ressources associated with them, except the backups
func CleanupServices() {
	var services_cleanup sync.WaitGroup

	for _, service := range config.GetConfiguration().Services {
		if service.Enabled {
			logger.Info("cleaning up", service.Name)
			services_cleanup.Add(1)
			go func(service config.Service) {
				defer services_cleanup.Done()

				err := CleanupService(&service)
				if err != nil {
					logger.Error("something went wrong while cleaning up", service.Name, "check the logs for more information")
				}
			}(service)
		}
	}

	services_cleanup.Wait()
}

// FetchServiceStatus returns the status of a service (from its container)
func FetchServiceStatus(service *config.Service) (string, error) {
	return containerutils.FetchContainerStatus(service.Container.Name)
}

func FetchServicesStatus() {
	var statuses sync.WaitGroup

	for _, service := range config.GetConfiguration().Services {
		if service.Enabled {
			statuses.Add(1)
			go func(service config.Service) {
				defer statuses.Done()

				status, err := FetchServiceStatus(&service)
				if err != nil {
					logger.Error("could not fetch status for", service.Name)
				}
				fmt.Printf("%s: %s\n", service.Name, status)
			}(service)
		}
	}

	statuses.Wait()
}

// GetServiceLogs returns the logs of a service (from its container)
func GetServiceLogs(service *config.Service) (string, error) {
	return containerutils.FetchContainerLogs(service.Container.Name)
}

// PauseService pauses the service by stopping the service container
func PauseService(service *config.Service) error {
	return containerutils.StopContainer(service.Container.Name)
}

// ResumeService resumes the service by starting the service container
func ResumeService(service *config.Service) error {
	return containerutils.ResumeContainer(service.Container.Name)
}

// StartService runs
func StartService(service_adm svcadm.ServiceAdm) error {
	logger.Debug(service_adm.GetServiceAdmName()+":", "service pre-init")
	additional_env, additional_volumes, cap_adds, entrypoint, err := service_adm.PreInit()
	if err != nil {
		return err
	}
	service := service_adm.GetService()

	container_env := make(map[string]string)
	for variable, value := range service.Container.Env {
		container_env[variable] = value
	}
	for variable, value := range additional_env {
		container_env[variable] = value
	}

	container_volumes := make(map[string]string)
	if service.Persistence.Enabled {
		for volume, path := range service.Persistence.Volumes {
			container_volumes[volume] = path
		}
	}
	for volume, path := range additional_volumes {
		container_volumes[volume] = path
	}

	logger.Debug(service_adm.GetServiceAdmName()+":", "service init")
	err = containerutils.CreateContainer(service.Container.Name, fmt.Sprintf("%s:%s", service.Image.Repository, service.Image.Tag), serviceLabels(service_adm), container_volumes, service.Container.Ports, container_env, service.Container.Restart, cap_adds, entrypoint)
	if err != nil {
		return err
	}

	logger.Debug(service_adm.GetServiceAdmName()+":", "service post-init")
	return service_adm.PostInit(additional_env)
}

// StartServices starts all services that are enabled in the configuration file
func StartServices() error {
	configuration := config.GetConfiguration()
	service_order := servicesStartOrder(&configuration.Services)

	// Start the services by batches asynchroniously
	for i, batch := range service_order {
		logger.Debug("starting batch", i+1)
		var services_start sync.WaitGroup
		for _, service_adm := range batch {
			services_start.Add(1)
			go func(service_adm svcadm.ServiceAdm) {
				defer services_start.Done()
				err := StartService(service_adm)
				if err != nil {
					logger.Error("could not start", service_adm.GetServiceAdmName())
				}
			}(service_adm)
		}
		services_start.Wait()
	}
	logger.Info("all services started")
	return nil
}
