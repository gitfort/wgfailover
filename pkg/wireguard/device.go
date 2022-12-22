package wireguard

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/gitfort/wgfailover/pkg/execext"
)

const (
	DeviceWait       = 30 * time.Second
	HandshakeTimeout = 180 * time.Second
)

func UpDevice(ctx context.Context, device string) error {
	up, upCancel := context.WithTimeout(ctx, DeviceWait)
	defer upCancel()
	defer time.Sleep(DeviceWait)
	if _, err := execext.CommandContextStream(up, "wg-quick", "up", device); err != nil {
		if downErr := DownDevice(ctx, device); downErr != nil {
			return downErr
		}
		return err
	}
	return nil
}

func DownDevice(ctx context.Context, device string) error {
	down, upCancel := context.WithTimeout(ctx, DeviceWait)
	defer upCancel()
	defer time.Sleep(DeviceWait)
	if _, err := execext.CommandContextStream(down, "wg-quick", "down", device); err != nil {
		if !strings.Contains(err.Error(), "No such file or directory") {
			return err
		}
		if _, err = execext.CommandContextStream(ctx, "ip", "link", "delete", "dev", device); err != nil {
			return err
		}
		if _, err = execext.CommandContextStream(ctx, "systemctl", "restart", "systemd-networkd"); err != nil {
			return err
		}
	}
	return nil
}

func LatestHandshake(ctx context.Context, device string) (time.Time, error) {
	var dump string
	for i := 0; i < 3 && dump == ""; i++ {
		var err error
		dump, err = execext.CommandContextStream(ctx, "wg", "show", device, "latest-handshakes")
		if err != nil {
			return time.Time{}, err
		}
	}
	result := strings.Split(dump, "=")
	if len(result) != 2 {
		return time.Time{}, nil
	}
	value, err := strconv.ParseInt(strings.TrimSpace(result[1]), 10, 64)
	if err != nil {
		return time.Time{}, err
	}
	return time.Unix(value, 0), nil
}

func ActiveDevices(ctx context.Context) ([]string, error) {
	var dump string
	for i := 0; i < 3 && dump == ""; i++ {
		var err error
		dump, err = execext.CommandContextStream(ctx, "wg", "show", "interfaces")
		if err != nil {
			return nil, err
		}
	}
	return strings.Split(dump, " "), nil
}
