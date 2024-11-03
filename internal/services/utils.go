package services

import (
	"github.com/boxboxjason/svcadm/internal/config"
	"github.com/boxboxjason/svcadm/internal/constants"
	"github.com/boxboxjason/svcadm/internal/services/svcadm"
	"github.com/boxboxjason/svcadm/pkg/logger"
)

func serviceLabels(service_adm svcadm.ServiceAdm) map[string]string {
	service := service_adm.GetService()
	labels := constants.SVCADM_LABELS
	labels["service"] = service.Name
	return labels
}

func servicesStartOrder(services *[]config.Service) [][]svcadm.ServiceAdm {
	var services_order [][]svcadm.ServiceAdm
	in_degree := make(map[string]int)
	adjacency_list := make(map[string][]string)

	// Initialize in-degree and adjacency list
	for _, service := range *services {
		if service.Enabled {
			service_name := service.Name
			// Initialize the in-degree to 0 initially
			in_degree[service_name] = len(config.SERVICE_NEEDS[service_name])
			adjacency_list[service_name] = []string{}
		}
	}

	// Populate adjacency list and calculate in-degree
	for _, service := range *services {
		if service.Enabled {
			service_name := service.Name
			for _, dependency := range config.SERVICE_NEEDS[service_name] {
				// Populate adjacency list: dependency -> service (dependency must come first)
				adjacency_list[dependency] = append(adjacency_list[dependency], service_name)
			}
		}
	}

	// Add services with in-degree 0 to the queue (services with no dependencies)
	queue := []string{}
	for service_name, degree := range in_degree {
		if degree == 0 {
			queue = append(queue, service_name)
		}
	}

	// Process services in topological order
	for len(queue) > 0 {
		current_batch := []svcadm.ServiceAdm{}
		next_queue := []string{}

		// Convert queue services to ServiceAdm and add to current_batch
		for _, service_name := range queue {
			serviceAdm, err := getServiceAdm(service_name)
			if err != nil {
				logger.Error(err)
				return nil
			}
			current_batch = append(current_batch, serviceAdm)

			// Process services dependent on the current service
			for _, dependent := range adjacency_list[service_name] {
				in_degree[dependent]--
				if in_degree[dependent] == 0 {
					next_queue = append(next_queue, dependent)
				}
			}
		}

		// Add current batch to services order and update the queue
		services_order = append(services_order, current_batch)
		queue = next_queue
	}

	// Check for cycles (if any service still has in-degree > 0)
	for _, degree := range in_degree {
		if degree > 0 {
			logger.Error("Service dependency cycle detected: no valid start order possible")
			logger.Debug("in-degree:", in_degree)
			logger.Debug("adjacency list:", adjacency_list)
			return nil
		}
	}

	return services_order
}
