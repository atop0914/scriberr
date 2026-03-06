package commands

import (
	"fmt"

	"scriberr/internal/config"
	"scriberr/internal/logger"

	"github.com/urfave/cli/v2"
)

// All returns all CLI commands
func All(log *logger.Logger, cfg *config.Config) []*cli.Command {
	cmds := []*cli.Command{
		ServerCommand(log, cfg),
		InitCommand(log, cfg),
		VersionCommand(log, cfg),
	}
	// Add model commands
	cmds = append(cmds, ModelCommands(log, cfg)...)
	return cmds
}

// ServerCommand starts the server
func ServerCommand(log *logger.Logger, cfg *config.Config) *cli.Command {
	return &cli.Command{
		Name:        "server",
		Aliases:     []string{"s"},
		Usage:       "Start the API server",
		Description: "Starts the scriberr API server",
		Category:    "Server",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "host",
				Usage: "Server host",
				Value: cfg.Server.Host,
			},
			&cli.IntFlag{
				Name:  "port",
				Usage: "Server port",
				Value: cfg.Server.Port,
			},
		},
		Action: func(cCtx *cli.Context) error {
			host := cCtx.String("host")
			port := cCtx.Int("port")
			log.Info("Starting server", "host", host, "port", port)
			fmt.Printf("Server starting on %s:%d\n", host, port)
			// TODO: Implement server startup
			return nil
		},
	}
}

// InitCommand initializes a new project
func InitCommand(log *logger.Logger, cfg *config.Config) *cli.Command {
	return &cli.Command{
		Name:        "init",
		Aliases:     []string{"i"},
		Usage:       "Initialize a new project",
		Description: "Creates a new scriberr project with default configuration",
		Category:    "Project",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "name",
				Usage: "Project name",
			},
			&cli.BoolFlag{
				Name:  "force",
				Usage: "Overwrite existing files",
			},
		},
		Action: func(cCtx *cli.Context) error {
			name := cCtx.String("name")
			if name == "" {
				name = "my-project"
			}
			log.Info("Initializing project", "name", name)
			fmt.Printf("Initializing project: %s\n", name)
			// TODO: Implement project initialization
			return nil
		},
	}
}

// VersionCommand shows version info
func VersionCommand(log *logger.Logger, cfg *config.Config) *cli.Command {
	return &cli.Command{
		Name:        "version",
		Aliases:     []string{"v"},
		Usage:       "Show version information",
		Description: "Display version and build information",
		Category:    "Info",
		Action: func(cCtx *cli.Context) error {
			fmt.Printf("scriberr %s\n", cfg.App.Version)
			fmt.Printf("Environment: %s\n", cfg.App.Environment)
			return nil
		},
	}
}
