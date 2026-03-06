package main

import (
	"fmt"
	"os"

	"scriberr/internal/commands"
	"scriberr/internal/config"
	"scriberr/internal/logger"

	"github.com/urfave/cli/v2"
)

func main() {
	// Initialize logger
	log := logger.New()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Warn("Using default config", "error", err)
		cfg = config.Default()
	}

	app := &cli.App{
		Name:                  "scriberr",
		Usage:                 "A brief description of your application",
		Version:               "0.1.0",
		EnableBashCompletion: true,
		Suggest:               true,
		HideVersion:           false,
		HideHelp:              false,
		HideHelpCommand:       false,
		Flags:                 configFlags(cfg),
		Commands:              commands.All(log, cfg),
		CommandNotFound:       commandNotFound,
		OnUsageError:          usageError,
		Before:                beforeHook(log),
		After:                 afterHook(log),
		Action:                defaultAction(log, cfg),
	}

	if err := app.Run(os.Args); err != nil {
		log.Error("Application error", "error", err)
		os.Exit(1)
	}
}

func configFlags(cfg *config.Config) []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:        "config",
			Aliases:     []string{"c"},
			Usage:       "Path to config file",
			DefaultText: "config.yaml",
			EnvVars:     []string{"SCIBERR_CONFIG"},
		},
		&cli.StringFlag{
			Name:        "log-level",
			Aliases:     []string{"l"},
			Usage:       "Log level (debug, info, warn, error)",
			DefaultText: "info",
			EnvVars:     []string{"SCIBERR_LOG_LEVEL"},
		},
	}
}

func commandNotFound(cCtx *cli.Context, command string) {
	fmt.Fprintf(cCtx.App.Writer, "Error: '%s' is not a valid command.\n", command)
	fmt.Fprintln(cCtx.App.Writer, "Run 'scriberr --help' for usage.")
}

func usageError(cCtx *cli.Context, err error, isSubcommand bool) error {
	return cli.Exit(fmt.Sprintf("Usage error: %v", err), 1)
}

func beforeHook(log *logger.Logger) cli.BeforeFunc {
	return func(cCtx *cli.Context) error {
		log.Debug("Starting application", "command", cCtx.Command.Name)
		return nil
	}
}

func afterHook(log *logger.Logger) cli.AfterFunc {
	return func(cCtx *cli.Context) error {
		log.Debug("Finished command", "command", cCtx.Command.Name)
		return nil
	}
}

func defaultAction(log *logger.Logger, cfg *config.Config) cli.ActionFunc {
	return func(cCtx *cli.Context) error {
		fmt.Println("scriberr - version", cCtx.App.Version)
		fmt.Println("Run 'scriberr --help' for usage.")
		return nil
	}
}
