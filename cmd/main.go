package main

import (
	"path/filepath"

	"github.com/boxboxjason/svcadm/internal/config"
	"github.com/boxboxjason/svcadm/internal/constants"
	"github.com/boxboxjason/svcadm/internal/services"
	"github.com/boxboxjason/svcadm/pkg/containerutils"
	"github.com/boxboxjason/svcadm/pkg/logger"
	"github.com/spf13/cobra"
)

var (
	config_file string
	log_level   string
)

func init() {
	root_cmd.PersistentFlags().StringVarP(&config_file, "config", "c", filepath.Join(constants.SVCADM_HOME, "svcadm.yaml"), "config file to use")
	root_cmd.PersistentFlags().StringVarP(&log_level, "loglevel", "l", "debug", "log level to use, one of: debug, info, warn, error, fatal")
}

// Entrypoint for the application & subcommands definition
func main() {
	root_cmd.AddCommand(setup_cmd, status_cmd, config_cmd, cleanup_cmd, backup_cmd)
	err := root_cmd.Execute()
	if err != nil {
		logger.Fatal(err)
	}
}

// Define the root command
var root_cmd = &cobra.Command{
	Use:   "svcadm",
	Short: "svcadm is a service manager for development environments",
	Long:  `svcadm is a service manager for development environments. Providing teams with a way to easily manage the services they need for their development environment.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Set the log level
		logger.SetupLogger(constants.LOG_DIR, log_level)

		// Set and validate the configuration file
		config.SetConfiguration(config_file)
		config.ValidateConfiguration()

		configuration := config.GetConfiguration()
		err := containerutils.SetContainerEngine(configuration.General.ContainerOperator.Name)
		if err != nil {
			logger.Fatal(err)
		}
		err = containerutils.SetContainersNetwork(configuration.General.ContainerOperator.Network.Name)
		if err != nil {
			logger.Fatal(err)
		}
	},
}

// Define the setup subcommand
var setup_cmd = &cobra.Command{
	Use:   "setup",
	Short: "Setup all services",
	Long:  "Setup all services defined in the configuration file",
	PreRun: func(cmd *cobra.Command, args []string) {
		// Set and validate the users
		configuration := config.GetConfiguration()
		config.SetUsers(configuration.General.Access.Logins)
		config.ValidateUsers()
	},
	Run: func(cmd *cobra.Command, args []string) {
		logger.Debug("setup requested")

		err := services.StartServices()
		if err != nil {
			logger.Fatal(err)
		}
	},
}

// Define the status subcommand
var status_cmd = &cobra.Command{
	Use:   "status",
	Short: "Get the status of services",
	Long:  "Get the status of services defined in the configuration file. Can request multiple services at once or all enabled services if no service is specified",
	Run: func(cmd *cobra.Command, args []string) {
		services.FetchServicesStatus()
	},
}

// Define the config subcommand
var config_cmd = &cobra.Command{
	Use:   "config",
	Short: "Manage the configuration file",
	Long:  "Manage the configuration file. Can be used to create a new configuration file, check the validity of an existring one, edit an existing one, or view the current configuration",
}

// Define the cleanup subcommand
var cleanup_cmd = &cobra.Command{
	Use:   "cleanup",
	Short: "Cleanup all services",
	Long:  "Cleanup all services defined in the configuration file",
	Run: func(cmd *cobra.Command, args []string) {
		logger.Debug("Cleanup requested")
		services.CleanupServices()
		logger.Info("all services cleaned up")
	},
}

// Define the backup subcommand
var backup_cmd = &cobra.Command{
	Use:   "backup",
	Short: "Backup services",
	Long:  "Start a manual backup of one or more services",
	Run: func(cmd *cobra.Command, args []string) {
		logger.Debug("Backup requested")

		err := services.BackupServices()
		if err != nil {
			logger.Fatal(err)
		}
		logger.Info("all services backed up")
	},
}
