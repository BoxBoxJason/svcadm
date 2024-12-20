package trivyadm

import (
	"fmt"

	"github.com/boxboxjason/svcadm/internal/config"
	"github.com/boxboxjason/svcadm/pkg/containerutils"
)

const (
	TRIVYADM            = "trivyadm"
	TRIVYADM_LOG_PREFIX = "trivyadm:"
)

type TrivyAdm struct {
	Service config.Service
}

func (t *TrivyAdm) PreInit() (map[string]string, map[string]string, map[int]int, []string, []string, error) {
	return nil, nil, nil, nil, []string{"server", "--listen", "0.0.0.0:4954"}, nil
}

func (t *TrivyAdm) CreateUser(user *config.User) error {
	return nil
}

func (t *TrivyAdm) CreateAdminUser(user *config.User) error {
	return nil
}

func (t *TrivyAdm) PostInit() error {
	return t.WaitFor()
}

func (t *TrivyAdm) Backup(backup_path string) error {
	return nil
}

func (t *TrivyAdm) WaitFor() error {
	return containerutils.WaitForContainerReadiness(t.Service.Container.Name, 5, 30)
}

func (t *TrivyAdm) GenerateNginxConf() string {
	return fmt.Sprintf(`# Trivy
location /%s/ {
	proxy_pass http://%s:4954/;
	proxy_http_version 1.1;
	proxy_set_header Host $host;
	proxy_set_header Upgrade $http_upgrade;
	proxy_set_header Connection "upgrade";
	proxy_set_header X-Real-IP $remote_addr;
	proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
}`, t.Service.Name, t.Service.Container.Name)
}

// GetService returns the service configuration
func (t *TrivyAdm) GetService() config.Service {
	return t.Service
}

func (t *TrivyAdm) GetServiceName() string {
	return t.Service.Name
}

func (t *TrivyAdm) GetServiceAdmName() string {
	return TRIVYADM
}

func (t *TrivyAdm) Cleanup() ([]string, []string) {
	return []string{}, []string{}
}
