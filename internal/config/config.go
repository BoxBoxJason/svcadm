package config

type Configuration struct {
	Services []Service `yaml:"services"`
	General  General   `yaml:"general"`
}

type General struct {
	Access            Access            `yaml:"access"`
	ContainerOperator ContainerOperator `yaml:"operator"`
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
}

type Image struct {
	Repository string `yaml:"repository"`
	Tag        string `yaml:"tag"`
}

type Container struct {
	Name    string            `yaml:"name"`
	Ports   []string          `yaml:"ports"`
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
