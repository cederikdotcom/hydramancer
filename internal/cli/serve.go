package cli

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/cederikdotcom/hydramancer/internal/api"
	"github.com/cederikdotcom/hydramancer/internal/config"
	"github.com/cederikdotcom/hydrarelease/pkg/updater"
	"github.com/cederikdotcom/hydraserve"
	"github.com/spf13/cobra"
)

func newServeCmd() *cobra.Command {
	var (
		addr string
		dev  bool
	)

	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Start the creator onboarding portal",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load(configPath)
			if err != nil {
				return fmt.Errorf("loading config: %w", err)
			}

			if !dev {
				u := updater.NewProductionUpdater("hydramancer", Version)
				u.SetServiceName("hydramancer")
				u.StartAutoCheck(6*time.Hour, true)
				log.Printf("Auto-update: enabled (every 6h)")
			}

			ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
			defer cancel()

			srv := api.NewServer()

			listen := addr
			if dev {
				if listen == "" {
					listen = ":8080"
				}
			} else if listen == "" && cfg.Server.Domain != "" {
				listen = ""
			} else if listen == "" {
				listen = ":8087"
			}

			if listen != "" {
				log.Printf("hydramancer %s listening on %s", Version, listen)
			} else {
				log.Printf("hydramancer %s listening on :443 (TLS) for %s", Version, cfg.Server.Domain)
			}

			go func() {
				<-ctx.Done()
				log.Println("shutting down...")
				os.Exit(0)
			}()

			return hydraserve.ListenAndServe(hydraserve.Config{
				Handler: srv.Handler(),
				Domain:  cfg.Server.Domain,
				CertDir: cfg.GetCertCache(),
				Listen:  listen,
			})
		},
	}

	cmd.Flags().StringVar(&addr, "addr", "", "listen address (overrides config)")
	cmd.Flags().BoolVar(&dev, "dev", false, "run in development mode (plain HTTP)")

	return cmd
}
