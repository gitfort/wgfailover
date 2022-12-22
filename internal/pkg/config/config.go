package config

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/gitfort/wgfailover/pkg/slices"
	"github.com/gitfort/wgfailover/pkg/wireguard"
	"gopkg.in/yaml.v3"
)

const (
	HealthCheckTimeout = 30 * time.Second
)

func Load(filepath string) (*Config, error) {
	var config *Config
	bin, err := ioutil.ReadFile(filepath)
	if err != nil {
		return nil, err
	}
	if err := yaml.Unmarshal(bin, &config); err != nil {
		return nil, err
	}
	return config, nil
}

type Config struct {
	Devices []Device `yaml:"devices"`
}

type Device struct {
	Name                   string         `yaml:"name"`
	LatestHandshakeTimeout *time.Duration `yaml:"latest_handshake_timeout"`
	HealthCheck            *HealthCheck   `yaml:"health_check"`
}

func (d *Device) GetLatestHandshakeTimeout() time.Duration {
	if d.LatestHandshakeTimeout != nil {
		return *d.LatestHandshakeTimeout
	}
	return wireguard.HandshakeTimeout
}

type HealthCheck struct {
	URL     string         `yaml:"url"`
	Timeout *time.Duration `yaml:"timeout"`
	Values  []string       `yaml:"values"`
}

func (h *HealthCheck) GetTimeout() time.Duration {
	if h.Timeout != nil {
		return *h.Timeout
	}
	return HealthCheckTimeout
}

func (hc *HealthCheck) Do(ctx context.Context) error {
	hcCTX, hcCancel := context.WithTimeout(ctx, hc.GetTimeout())
	defer hcCancel()
	var (
		err  error
		data []byte
		req  *http.Request
		res  *http.Response
	)
	req, err = http.NewRequestWithContext(hcCTX, http.MethodGet, hc.URL, nil)
	if err != nil {
		return err
	}
	res, err = http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer func() {
		_ = res.Body.Close()
	}()
	data, err = io.ReadAll(res.Body)
	if err != nil {
		return err
	}
	if slices.Find(string(data), hc.Values) >= 0 {
		return nil
	}
	return fmt.Errorf("health check returns %s", data)
}
