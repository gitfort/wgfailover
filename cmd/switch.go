package cmd

import (
	"context"
	"log"

	"github.com/gitfort/wgfailover/internal/pkg/config"
	"github.com/gitfort/wgfailover/pkg/slices"
	"github.com/gitfort/wgfailover/pkg/wireguard"
	"github.com/spf13/cobra"
)

var switchCmd = &cobra.Command{
	Use:   "switch",
	Short: "Switch to the selected interface between failovers",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load(configFile)
		if err != nil {
			return err
		}
		return NetworkSwitch(cmd.Context(), cfg, args[0])
	},
}

func init() {
	rootCmd.AddCommand(switchCmd)
}

func NetworkSwitch(ctx context.Context, cfg *config.Config, network string) error {
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

	log.Printf("activating %v", network)
	if err = wireguard.UpDevice(ctx, network); err != nil {
		return err
	}
	return nil
}
