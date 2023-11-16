package server

import (
	"os"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

type TestSettings struct {
	IperfSettings *IperfSettings `yaml:"iperf_settings"`
}

type IperfSettings struct {
	Servers  []string `yaml:"servers" json:"servers"`
	User     string   `yaml:"user" json:"user"`
	Password string   `yaml:"password" json:"password"`
	Threads  int      `yaml:"iperf_threads" json:"iperf_threads"`
	Time     int      `yaml:"time" json:"time"`
}

func ParseSettings(cfgFile string) (*TestSettings, error) {
	out, err := os.ReadFile(cfgFile)
	if err != nil {
		return nil, err
	}
	settings := TestSettings{}
	err = yaml.Unmarshal(out, &settings)
	if err != nil {
		return nil, err
	}
	if settings.IperfSettings == nil {
		return nil, errors.Errorf("iperf parameters should be specified")
	}
	if len(settings.IperfSettings.Servers) < 2 {
		return nil, errors.Errorf("iperf servers should be specified")
	}
	return &settings, nil
}
