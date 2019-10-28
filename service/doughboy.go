package service

import (
	"os"

	"github.com/hashicorp/consul/api"
	"github.com/pkg/errors"

	"gophers.dev/cmds/doughboy/config"
	"gophers.dev/pkgs/loggy"
)

type DoughBoy struct {
	configuration *config.Configuration
	log           loggy.Logger

	consul *api.Client
}

func New(args []string) (*DoughBoy, error) {
	if len(args) != 1 {
		return nil, errors.New("config file required")
	}

	logger := loggy.New("doughboy")

	c, err := config.LoadHCL(os.Args[1])
	if err != nil {
		return nil, errors.Wrap(err, "malformed config file")
	}

	c.Log(logger)

	return &DoughBoy{
		configuration: c,
		log:           logger,
	}, nil
}

func (db *DoughBoy) Start() error {
	db.log.Infof("starting up, pid: %d", os.Getpid())

	for _, f := range []initializer{
		initSigs,
		initConsul,
		initNativeResponder,
		initNativeRequester,
		initClassicResponder,
		initClassicRequester,
	} {
		if err := f(db); err != nil {
			return errors.Wrap(err, "initialization failed")
		}
	}

	return nil
}

func (db *DoughBoy) Wait() {
	select {
	// wait forever
	// we could add a shutdown signal perhaps
	}
}
