package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/FreeNetLabs/tunn/internal/tunnel"
	"github.com/FreeNetLabs/tunn/pkg/config"

	"github.com/spf13/cobra"
)

type contextKey string

const configKey contextKey = "cfg"

var rootCmd = &cobra.Command{
	Use:     "tunn",
	Short:   "A powerful tunnel tool for secure connections",
	Version: "v0.1.2",

	PreRunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadConfig(configFile)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		cmd.SetContext(context.WithValue(cmd.Context(), configKey, cfg))
		return nil
	},

	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, ok := cmd.Context().Value(configKey).(*config.Config)
		if !ok {
			return fmt.Errorf("failed to retrieve config from context")
		}

		fmt.Printf("Mode: %s\n\n", cfg.Mode)

		manager := tunnel.NewManager(cfg)
		if err := manager.Start(); err != nil {
			return fmt.Errorf("failed to start tunnel: %w", err)
		}
		return nil
	},
}

var configFile string

func init() {
	rootCmd.PersistentFlags().StringVarP(&configFile, "config", "c", "config.json", "config file path")
	rootCmd.CompletionOptions.DisableDefaultCmd = true
	rootCmd.SetHelpCommand(&cobra.Command{Use: "no-help", Hidden: true})
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
