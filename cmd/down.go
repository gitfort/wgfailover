package cmd

import (
	"context"
	"log"

	"github.com/gitfort/wgfailover/internal/pkg/config"
	"github.com/gitfort/wgfailover/pkg/slices"
	"github.com/gitfort/wgfailover/pkg/wireguard"
	"github.com/spf13/cobra"
)

var downCmd = &cobra.Command{
	Use:   "down",
	Args:  cobra.ExactArgs(0),
	Short: "Deactivate all failover interfaces",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load(configFile)
		if err != nil {
			return err
		}
		return NetworkDown(cmd.Context(), cfg)
	},
}

func init() {
	rootCmd.AddCommand(downCmd)
}

func NetworkDown(ctx context.Context, cfg *config.Config) error {
	activeDevices, err := wireguard.ActiveDevices(ctx)
	if err != nil {
		return err
	}
	for _, device := range cfg.Devices {
		if slices.Find(device.Name, activeDevices) >= 0 {
			log.Printf("deactivating %v", device.Name)
			if err = wireguard.DownDevice(ctx, device.Name); err != nil {
				return err
			}
		}
	}

	return nil
}
