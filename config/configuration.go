package config

import (
	"io/ioutil"

	"github.com/hashicorp/hcl"

	"gophers.dev/pkgs/loggy"
)

type Service struct {
	Address string `hcl:"address"`
	Port    int    `hcl:"port"`
}

type Configuration struct {
	Service Service  `hcl:"service"`
	Signals []string `hcl:"signals"`
}

func LoadHCL(filename string) (*Configuration, error) {
	bs, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var c Configuration
	if err := hcl.Decode(&c, string(bs)); err != nil {
		return nil, err
	}

	return &c, nil
}

func (c *Configuration) Log(logger loggy.Logger) {
	logger.Infof("service.address: %s", c.Service.Address)
	logger.Infof("service.port: %d", c.Service.Port)
	logger.Infof("signals: %v", c.Signals)
}
