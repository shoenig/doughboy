package config

import (
	"fmt"
	"io/ioutil"

	"github.com/hashicorp/hcl"

	"gophers.dev/pkgs/loggy"
)

type Consul struct {
	Port int `hcl:"port"`
}

type Server struct {
	Enable bool   `hcl:"enable"`
	Bind   string `hcl:"bind"`
	Port   int    `hcl:"port"`
}

func (s Server) String() string {
	return fmt.Sprintf(
		"(enabled: %t, bind: %s, port %d)",
		s.Enable, s.Bind, s.Port,
	)
}

type Client struct {
	Enable bool `hcl:"enable"`
}

func (c Client) String() string {
	return fmt.Sprintf(
		"(enabled: %t)",
		c.Enable,
	)
}

type Echo struct {
	Native struct {
		Server Server `hcl:"server"`
		Client Client `hcl:"client"` // todo: also listen for HCs
	} `hcl:"native"`
	Classic struct {
		Server Server `hcl:"server"`
		Client Server `hcl:"client"`
	} `hcl:"classic"`
}

type Configuration struct {
	Signals []string `hcl:"signals"`
	Consul  Consul   `hcl:"consul"`
	Echo    Echo     `hcl:"echo"`
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
	logger.Tracef("config | signals: %v", c.Signals)
	logger.Tracef("config | consul agent port: %d", c.Consul.Port)
	logger.Tracef("config | echo native server: %s", c.Echo.Native.Server)
	logger.Tracef("config | echo native client: %s", c.Echo.Native.Client)
	logger.Tracef("config | echo classic server: %s", c.Echo.Classic.Server)
	logger.Tracef("config | echo classic client: %s", c.Echo.Classic.Client)
}
