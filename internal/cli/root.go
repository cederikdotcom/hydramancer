package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var configPath string

func NewRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "hydramancer",
		Short: "Creator onboarding portal for Experiencenet",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Printf("hydramancer %s\n", Version)
			return cmd.Help()
		},
	}

	cmd.PersistentFlags().StringVar(&configPath, "config", "", "config file path (default ~/.hydramancer/config.yaml)")

	cmd.AddCommand(newServeCmd())
	cmd.AddCommand(newVersionCmd())

	return cmd
}

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("hydramancer %s\n", Version)
		},
	}
}
