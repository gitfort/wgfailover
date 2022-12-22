package cmd

import (
	"context"
	"log"
	"time"

	"github.com/gitfort/wgfailover/internal/pkg/config"
	"github.com/gitfort/wgfailover/pkg/slices"
	"github.com/gitfort/wgfailover/pkg/wireguard"
	"github.com/robfig/cron/v3"
	"github.com/spf13/cobra"
)

const (
	HealthCheckTimeout        = 30 * time.Second
	WireguardHandshakeTimeout = 3 * time.Minute
	NetworkReviewCron         = "*/3 * * * *"
)

var (
	InProgress = false
	reviewCmd  = &cobra.Command{
		Use:   "review",
		Short: "Review interfaces and automatic switch between them",
		Args:  cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load(configFile)
			if err != nil {
				return err
			}
			if err = NetworkReview(cmd.Context(), cfg); err != nil {
				log.Print(err)
			}
			scheduler := cron.New()
			if _, err = scheduler.AddFunc(NetworkReviewCron, func() {
				if InProgress {
					return
				}
				InProgress = true
				defer func() {
					InProgress = false
				}()
				if err = NetworkReview(cmd.Context(), cfg); err != nil {
					log.Print(err)
				}
			}); err != nil {
				return err
			}
			scheduler.Run()
			return nil
		},
	}
)

func init() {
	rootCmd.AddCommand(reviewCmd)
}

func NetworkReview(ctx context.Context, cfg *config.Config) error {
	activeDevices, err := wireguard.ActiveDevices(ctx)
	if err != nil {
		return err
	}

	var activeDeviceIndex int = -1
	for index, device := range cfg.Devices {
		if slices.Find(device.Name, activeDevices) >= 0 {
			if activeDeviceIndex >= 0 {
				log.Printf("deactivating %v due to conflict", device)
				if err = wireguard.DownDevice(ctx, device.Name); err != nil {
					return err
				}
				continue
			}
			activeDeviceIndex = index
		}
	}

	if activeDeviceIndex < 0 {
		activeDeviceIndex = 0
		log.Printf("activating %v as default device", cfg.Devices[activeDeviceIndex].Name)
		if err := wireguard.UpDevice(ctx, cfg.Devices[activeDeviceIndex].Name); err != nil {
			return err
		}
	}

	var latestHandshake time.Time
	latestHandshake, err = wireguard.LatestHandshake(ctx, cfg.Devices[activeDeviceIndex].Name)
	if err != nil {
		return err
	}

	if time.Since(latestHandshake) <= cfg.Devices[activeDeviceIndex].GetLatestHandshakeTimeout() {
		if hc := cfg.Devices[activeDeviceIndex].HealthCheck; hc == nil {
			log.Print("everything looks good")
			return nil
		} else if err = hc.Do(ctx); err == nil {
			log.Print("everything looks good")
			return nil
		} else {
			log.Printf("deactivating %v because %v", cfg.Devices[activeDeviceIndex].Name, err)
		}
	} else {
		log.Printf("deactivating %v because %v has passed since its latest handshake", cfg.Devices[activeDeviceIndex].Name, cfg.Devices[activeDeviceIndex].GetLatestHandshakeTimeout())
	}

	if err = wireguard.DownDevice(ctx, cfg.Devices[activeDeviceIndex].Name); err != nil {
		return err
	}

	nextActiveDeviceIndex := activeDeviceIndex + 1
	if nextActiveDeviceIndex >= len(cfg.Devices) {
		nextActiveDeviceIndex = 0
	}
	log.Printf("activating %v as next device", cfg.Devices[nextActiveDeviceIndex].Name)
	if err = wireguard.UpDevice(ctx, cfg.Devices[nextActiveDeviceIndex].Name); err != nil {
		return err
	}

	return NetworkReview(ctx, cfg)
}
