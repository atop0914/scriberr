package commands

import (
	"fmt"
	"strings"

	"scriberr/internal/config"
	"scriberr/internal/logger"
	"scriberr/internal/model"

	"github.com/urfave/cli/v2"
)

// ModelCommands returns all model-related CLI commands
func ModelCommands(log *logger.Logger, cfg *config.Config) []*cli.Command {
	return []*cli.Command{
		ListModelsCommand(log, cfg),
		DownloadModelCommand(log, cfg),
		DeleteModelCommand(log, cfg),
		ClearCacheCommand(log, cfg),
		ModelStatusCommand(log, cfg),
	}
}

// ListModelsCommand lists all available models
func ListModelsCommand(log *logger.Logger, cfg *config.Config) *cli.Command {
	return &cli.Command{
		Name:        "list-models",
		Aliases:     []string{"lm"},
		Usage:       "List all available Whisper models",
		Description: "Shows all available Whisper models with their sizes and download status",
		Category:    "Model",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  "cached",
				Usage: "Show only cached models",
			},
			&cli.BoolFlag{
				Name:  "available",
				Usage: "Show only available models (not cached)",
			},
		},
		Action: func(cCtx *cli.Context) error {
			showCached := cCtx.Bool("cached")
			showAvailable := cCtx.Bool("available")

			// Default: show all
			showAll := !showCached && !showAvailable

			fmt.Println("Available Whisper Models:")
			fmt.Println(strings.Repeat("=", 70))

			available := model.AvailableModels()
			cached := model.ListCachedModels(log)
			cachedMap := make(map[string]bool)
			for _, c := range cached {
				cachedMap[string(c.Size)] = true
			}

			for _, m := range available {
				isCached := cachedMap[string(m.Size)]

				// Filter based on flags
				if !showAll {
					if showCached && !isCached {
						continue
					}
					if showAvailable && isCached {
						continue
					}
				}

				status := "[ ]"
				if isCached {
					status = "[✓]"
				}

				fmt.Printf("%s %-10s %-8s %-10s - %s\n",
					status,
					string(m.Size),
					m.Params,
					m.FileSize,
					m.Description,
				)
			}

			fmt.Println(strings.Repeat("=", 70))
			fmt.Printf("Total: %d models\n", len(available))

			return nil
		},
	}
}

// DownloadModelCommand downloads a model
func DownloadModelCommand(log *logger.Logger, cfg *config.Config) *cli.Command {
	return &cli.Command{
		Name:        "download-model",
		Aliases:     []string{"dlm"},
		Usage:       "Download a Whisper model",
		Description: "Downloads and caches the specified Whisper model",
		Category:    "Model",
		ArgsUsage:   "<model-size>",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  "force",
				Aliases: []string{"f"},
				Usage: "Force re-download even if cached",
			},
		},
		Action: func(cCtx *cli.Context) error {
			sizeStr := cCtx.Args().First()
			if sizeStr == "" {
				return fmt.Errorf("model size required (tiny, base, small, medium, large, large-v2, large-v3)")
			}

			size := model.ModelSize(sizeStr)
			if !model.IsValidModelSize(sizeStr) {
				return fmt.Errorf("invalid model size: %s", sizeStr)
			}

			// Check if cached (unless force flag)
			if !cCtx.Bool("force") && model.IsModelCached(size) {
				path, _ := model.ModelPath(size)
				fmt.Printf("Model '%s' already cached at: %s\n", size, path)
				return nil
			}

			downloader := model.NewDownloader(log)
			downloader.SetProgressCallback(func(downloaded, total int64) {
				percent := float64(downloaded) / float64(total) * 100
				fmt.Printf("\rDownloading: %.1f%% (%s / %s)",
					percent,
					formatSize(downloaded),
					formatSize(total))
			})

			fmt.Printf("Downloading model: %s\n", size)
			path, err := downloader.Download(size)
			if err != nil {
				return fmt.Errorf("download failed: %w", err)
			}

			fmt.Printf("\n✓ Model downloaded successfully!\n")
			fmt.Printf("  Path: %s\n", path)

			return nil
		},
	}
}

// DeleteModelCommand deletes a cached model
func DeleteModelCommand(log *logger.Logger, cfg *config.Config) *cli.Command {
	return &cli.Command{
		Name:        "delete-model",
		Aliases:     []string{"rmm"},
		Usage:       "Delete a cached model",
		Description: "Removes a cached Whisper model from the local cache",
		Category:    "Model",
		ArgsUsage:   "<model-size>",
		Action: func(cCtx *cli.Context) error {
			sizeStr := cCtx.Args().First()
			if sizeStr == "" {
				return fmt.Errorf("model size required")
			}

			size := model.ModelSize(sizeStr)
			if !model.IsValidModelSize(sizeStr) {
				return fmt.Errorf("invalid model size: %s", sizeStr)
			}

			if !model.IsModelCached(size) {
				return fmt.Errorf("model not cached: %s", sizeStr)
			}

			downloader := model.NewDownloader(log)
			if err := downloader.DeleteModel(size); err != nil {
				return fmt.Errorf("failed to delete model: %w", err)
			}

			fmt.Printf("✓ Model '%s' deleted from cache\n", size)
			return nil
		},
	}
}

// ClearCacheCommand clears all cached models
func ClearCacheCommand(log *logger.Logger, cfg *config.Config) *cli.Command {
	return &cli.Command{
		Name:        "clear-cache",
		Aliases:     []string{"cc"},
		Usage:       "Clear all cached models",
		Description: "Removes all cached Whisper models from the local cache",
		Category:    "Model",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  "force",
				Aliases: []string{"f"},
				Usage: "Skip confirmation prompt",
			},
		},
		Action: func(cCtx *cli.Context) error {
			// Get cache stats before
			_, totalSize, _ := model.CacheStats(log)

			if !cCtx.Bool("force") {
				fmt.Printf("This will delete all cached models (%s). Continue? [y/N]: ",
					formatSize(totalSize))
				var response string
				fmt.Scanln(&response)
				if strings.ToLower(response) != "y" {
					fmt.Println("Cancelled.")
					return nil
				}
			}

			downloader := model.NewDownloader(log)
			if err := downloader.ClearCache(); err != nil {
				return fmt.Errorf("failed to clear cache: %w", err)
			}

			fmt.Printf("✓ Cache cleared (%s freed)\n", formatSize(totalSize))
			return nil
		},
	}
}

// ModelStatusCommand shows model cache status
func ModelStatusCommand(log *logger.Logger, cfg *config.Config) *cli.Command {
	return &cli.Command{
		Name:        "model-status",
		Aliases:     []string{"ms"},
		Usage:       "Show model cache status",
		Description: "Displays information about the model cache",
		Category:    "Model",
		Action: func(cCtx *cli.Context) error {
			cacheDir, err := model.EnsureCacheDir()
			if err != nil {
				return fmt.Errorf("failed to get cache directory: %w", err)
			}

			totalModels, totalSize, err := model.CacheStats(log)
			if err != nil {
				return fmt.Errorf("failed to get cache stats: %w", err)
			}

			cached := model.ListCachedModels(log)

			fmt.Println("Model Cache Status:")
			fmt.Println(strings.Repeat("=", 50))
			fmt.Printf("Cache Directory: %s\n", cacheDir)
			fmt.Printf("Cached Models: %d\n", totalModels)
			fmt.Printf("Total Size: %s\n", formatSize(totalSize))
			fmt.Println(strings.Repeat("=", 50))

			if len(cached) > 0 {
				fmt.Println("Cached Models:")
				for _, m := range cached {
					fmt.Printf("  • %s (%s)\n", m.Size, m.FileSize)
				}
			} else {
				fmt.Println("No models cached.")
			}

			// Show default model from config if set
			if cfg.Model.Size != "" {
				isCached := model.IsModelCached(model.ModelSize(cfg.Model.Size))
				status := "Not cached"
				if isCached {
					status = "Cached"
				}
				fmt.Printf("\nDefault Model (%s): %s\n", cfg.Model.Size, status)
			}

			return nil
		},
	}
}

// Helper function to format file sizes
func formatSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
