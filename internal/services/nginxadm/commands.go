package nginxadm

import (
	"fmt"
	"path"

	"github.com/boxboxjason/svcadm/internal/config"
	"github.com/boxboxjason/svcadm/internal/constants"
	"github.com/boxboxjason/svcadm/internal/services/gitlabadm"
	"github.com/boxboxjason/svcadm/internal/services/mattermostadm"
	"github.com/boxboxjason/svcadm/internal/services/minioadm"
	"github.com/boxboxjason/svcadm/internal/services/psqladm"
	"github.com/boxboxjason/svcadm/internal/services/sonaradm"
	"github.com/boxboxjason/svcadm/internal/services/svcadm"
	"github.com/boxboxjason/svcadm/internal/services/trivyadm"
	"github.com/boxboxjason/svcadm/internal/services/vaultadm"
	"github.com/boxboxjason/svcadm/pkg/containerutils"
	"github.com/boxboxjason/svcadm/pkg/fileutils"
	"github.com/boxboxjason/svcadm/pkg/formatutils"
	"github.com/boxboxjason/svcadm/pkg/logger"
	"github.com/boxboxjason/svcadm/pkg/utils"
)

const (
	NGINXADM            = "nginxadm"
	NGINXADM_LOG_PREFIX = "nginxadm:"
)

var (
	NGINXADM_PATH   = path.Join(constants.SVCADM_HOME, "nginxadm")
	NGINX_CONF_PATH = path.Join(NGINXADM_PATH, "nginx.conf")
)

const NGINX_CONF = `server {
	listen 80;
	listen [::]:80;
	server_name %s;
	return 301 https://$host$request_uri;
}

map $http_upgrade $connection_upgrade {
    default upgrade;
    ''      "";
}

server {
	listen 443 ssl;
	listen [::]:443 ssl;
	server_name %s;
	ssl_certificate %s;
	ssl_certificate_key %s;
	ssl_protocols TLSv1.2 TLSv1.3;
	ssl_ciphers HIGH:!aNULL:!MD5;

%s
}
`

type NginxAdm struct {
	Service config.Service
}

// PreInit sets up the nginx service by generating the nginx configuration files with each service location
func (n *NginxAdm) PreInit() (map[string]string, map[string]string, map[int]int, []string, []string, error) {
	// Generate the nginx configuration file
	nginx_locations := ""
	configuration := config.GetConfiguration()
	for _, service := range configuration.Services {
		if service.Enabled && service.Nginx {
			logger.Debug(NGINXADM_LOG_PREFIX, "adding location for", service.Name)
			nginx_locations += formatutils.IndentMultilineString(getServiceNginx(&service), 4) + "\n"
		}
	}
	hostname := utils.GetHostname()
	nginx_conf := fmt.Sprintf(NGINX_CONF, hostname, hostname, "/etc/ssl/certs/svcadm.crt", "/etc/ssl/private/svcadm.key", nginx_locations)
	err := fileutils.WriteToFile(NGINX_CONF_PATH, nginx_conf)

	return nil, map[string]string{NGINX_CONF_PATH: "/etc/nginx/conf.d/default.conf:Z"}, nil, nil, nil, err
}

// PostInit sets up the nginx service after the configuration files have been generated (empty because nginx does not require post init)
func (n *NginxAdm) PostInit() error {
	return n.WaitFor()
}

// CreateUser creates a user for the nginx service (empty because nginx does not require user)
func (n *NginxAdm) CreateUser(user *config.User) error {
	return nil
}

// CreateAdminUser creates an admin user for the nginx service (empty because nginx does not require admin user)
func (n *NginxAdm) CreateAdminUser(user *config.User) error {
	return nil
}

// Backup creates a backup of the nginx configuration files (empty because nginx does not require backup)
func (n *NginxAdm) Backup(backup_path string) error {
	return nil
}

// WaitFor waits for the nginx service to be ready, using a curl request to the nginx server
func (n *NginxAdm) WaitFor() error {
	return containerutils.WaitForContainerReadiness(n.Service.Container.Name, 5, 60)
}

// GenerateNginxConf generates the nginx configuration file (empty because nginx is not proxified)
func (n *NginxAdm) GenerateNginxConf() string {
	return ""
}

// getServiceAdm returns the service adm for a service
func getServiceNginx(service *config.Service) string {
	var serviceAdm svcadm.ServiceAdm
	switch service.Name {
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
		serviceAdm = &NginxAdm{Service: *service}
	case "vault":
		serviceAdm = &vaultadm.VaultAdm{Service: *service}
	case "trivy":
		serviceAdm = &trivyadm.TrivyAdm{Service: *service}
	default:
		return ""
	}
	return serviceAdm.GenerateNginxConf()
}

// GetService returns the service object from the configuration
func (n *NginxAdm) GetService() config.Service {
	return n.Service
}

func (n *NginxAdm) GetServiceName() string {
	return n.Service.Name
}

func (n *NginxAdm) GetServiceAdmName() string {
	return NGINXADM
}

func (n *NginxAdm) Cleanup() ([]string, []string) {
	return []string{}, []string{NGINXADM_PATH}
}
