package cli

import (
	"fmt"

	"github.com/cederikdotcom/hydrarelease/pkg/updater"
	"github.com/spf13/cobra"
)

// newUpdateCmd returns the standard hydra self-update command.
//
// Note on the containerised deployment: hydramancer normally runs from an OCI
// image, where the image is the unit of update and there is no systemd to
// restart, so `serve` leaves auto-update off (see newServeCmd). These commands
// exist for host installs and for forcing a binary update over
// `hydracluster exec` when redeploying the image is not practical. After using
// --force inside a container, restart the container so the new binary is the
// running process.
func newUpdateCmd() *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update to the latest release",
		RunE: func(cmd *cobra.Command, args []string) error {
			u := updater.NewProductionUpdater("hydramancer", Version)
			u.SetServiceName("hydramancer")
			return u.RunInteractiveUpdate(force)
		},
	}

	cmd.Flags().BoolVar(&force, "force", false, "skip confirmation prompt")

	return cmd
}

func newCheckUpdateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "check-update",
		Short: "Check if a new version is available",
		RunE: func(cmd *cobra.Command, args []string) error {
			u := updater.NewProductionUpdater("hydramancer", Version)

			info, err := u.CheckForUpdate()
			if err != nil {
				return fmt.Errorf("checking for update: %w", err)
			}

			fmt.Printf("Current version: %s\n", info.CurrentVersion)
			fmt.Printf("Latest version:  %s\n", info.LatestVersion)

			if info.Available {
				fmt.Println("\nA new version is available. Run 'hydramancer update' to install it.")
			} else {
				fmt.Println("\nAlready up to date.")
			}

			return nil
		},
	}
}
