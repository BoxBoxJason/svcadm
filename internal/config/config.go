package config

import (
	"path"
	"sync"

	"github.com/boxboxjason/svcadm/internal/constants"
	"github.com/boxboxjason/svcadm/pkg/logger"
)

var (
	configuration      Configuration
	mu_CONFIGURATION   sync.RWMutex
	users              Users
	mu_USERS           sync.RWMutex
	CONFIGURATION_PATH = path.Join(constants.SVCADM_HOME, "svcadm.yaml")
)

// SetConfiguration sets the configuration to be used
// Warning: if you do not pass one of the supported types, the configuration will be set as empty
func SetConfiguration(config interface{}) {
	mu_CONFIGURATION.Lock()
	defer mu_CONFIGURATION.Unlock()
	if c, ok := config.(Configuration); ok {
		configuration = c
	} else if c, ok := config.(*Configuration); ok {
		configuration = *c
	} else if c, ok := config.(string); ok {
		conf, _, err := ParseConfiguration(c)
		if err != nil {
			logger.Fatal(err)
		} else {
			configuration = conf
		}
	} else {
		logger.Fatal("unsupported type for configuration")
		configuration = Configuration{}
	}
	logger.Debug("configuration set")
}

// GetConfiguration gets the configuration to be used
func GetConfiguration() Configuration {
	mu_CONFIGURATION.RLock()
	defer mu_CONFIGURATION.RUnlock()
	return configuration
}

// SetUsers sets the users to be used
// Warning: if you do not pass one of the supported types, the users will be set as empty
func SetUsers(u interface{}) {
	mu_USERS.Lock()
	defer mu_USERS.Unlock()
	if c, ok := u.(Users); ok {
		users = c
	} else if c, ok := u.(*Users); ok {
		users = *c
	} else if c, ok := u.(string); ok {
		usrs, err := ParseUsers(c)
		if err != nil {
			logger.Fatal(err)
		} else {
			users = usrs
		}
	} else {
		logger.Fatal("unsupported type for users")
		users = Users{}
	}
}

// GetUsers gets the users to be used
func GetUsers() Users {
	mu_USERS.RLock()
	defer mu_USERS.RUnlock()
	return users
}

type Configuration struct {
	Services []Service `yaml:"services"`
	General  General   `yaml:"general"`
}

type General struct {
	Access            Access            `yaml:"access"`
	ContainerOperator ContainerOperator `yaml:"operator"`
	Containers        Containers        `yaml:"containers"`
}

type ContainerOperator struct {
	Name    string  `yaml:"name"`
	Network Network `yaml:"network"`
}

type Access struct {
	Logins     string     `yaml:"logins"`
	Encryption Encryption `yaml:"encryption"`
}

type Network struct {
	Name   string `yaml:"name"`
	Driver string `yaml:"driver"`
}

type Service struct {
	Name        string      `yaml:"name"`
	Enabled     bool        `yaml:"enabled"`
	Image       Image       `yaml:"image"`
	Container   Container   `yaml:"container"`
	Persistence Persistence `yaml:"persistence"`
	Backup      Backup      `yaml:"backup"`
	Nginx       bool        `yaml:"nginx"`
}

type Image struct {
	Repository string `yaml:"repository"`
	Tag        string `yaml:"tag"`
}

type Container struct {
	Name    string            `yaml:"name"`
	Ports   map[int]int       `yaml:"ports"`
	Restart string            `yaml:"restart"`
	Env     map[string]string `yaml:"env"`
}

type Persistence struct {
	Enabled bool              `yaml:"enabled"`
	Volumes map[string]string `yaml:"volumes"`
}

type Backup struct {
	Enabled   bool   `yaml:"enabled"`
	Frequency string `yaml:"frequency"`
	Retention int    `yaml:"retention"`
	Location  string `yaml:"location"`
}

type User struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Email    string `yaml:"email"`
}

type Users struct {
	Admins []User `yaml:"admins"`
	Users  []User `yaml:"users"`
}

type Encryption struct {
	Enabled bool   `yaml:"enabled"`
	Key     string `yaml:"key"`
	Salt    string `yaml:"salt"`
}

type Containers struct {
	Labels map[string]string `yaml:"labels"`
}
