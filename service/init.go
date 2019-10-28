package service

import (
	"fmt"

	"github.com/hashicorp/consul/api"
	"github.com/pkg/errors"

	"gophers.dev/cmds/doughboy/connect/classic"
	"gophers.dev/cmds/doughboy/connect/native"
	"gophers.dev/cmds/doughboy/sigs"

	"oss.indeed.com/go/libtime"
)

type initializer func(db *DoughBoy) error

func initSigs(db *DoughBoy) error {
	db.log.Tracef("initializing signal watchers")
	signals, err := sigs.Lookup(db.configuration.Signals...)
	if err != nil {
		return err
	}
	db.log.Tracef("will watch signals: %v", signals)

	sigWatcher := sigs.New(libtime.SystemClock())
	go sigWatcher.Watch(signals...)
	return nil
}

func initConsul(db *DoughBoy) error {
	db.log.Tracef("initializing consul client")

	port := db.configuration.Consul.Port
	consulConfig := api.DefaultConfig()
	consulConfig.Address = fmt.Sprintf("127.0.0.1:%d", port)

	consulClient, err := api.NewClient(consulConfig)
	if err != nil {
		return errors.Wrap(err, "cannot create consul client")
	}

	db.consul = consulClient
	return nil
}

func initNativeResponder(db *DoughBoy) error {
	if db.configuration.Echo.Native.Server.Enable {
		db.log.Tracef("initializing native echo service responder")
		r := native.NewResponder(
			db.configuration.Echo.Native.Server,
			db.consul,
		)
		if err := r.Open(); err != nil {
			return errors.Wrap(err, "cannot start native responder")
		}
	}
	return nil
}

func initNativeRequester(db *DoughBoy) error {
	if db.configuration.Echo.Native.Client.Enable {
		db.log.Tracef("initializing native echo service requester")
		requester := native.NewRequester(db.consul)
		if err := requester.Open(); err != nil {
			return errors.Wrap(err, "cannot start native requester")
		}
	}

	return nil
}

func initClassicResponder(db *DoughBoy) error {
	if db.configuration.Echo.Classic.Server.Enable {
		db.log.Tracef("initializing classic echo service responder")
		r := classic.NewResponder(db.configuration.Echo.Classic.Server)
		if err := r.Open(); err != nil {
			return errors.Wrap(err, "cannot start classic responder")
		}
	}
	return nil
}

func initClassicRequester(db *DoughBoy) error {
	if db.configuration.Echo.Classic.Client.Enable {
		db.log.Tracef("initializing classic echo service requester")
		r := classic.NewRequester(
			db.configuration.Echo.Classic.Client,
			db.consul,
		)
		if err := r.Open(); err != nil {
			return errors.Wrap(err, "cannot start classic requester")
		}
	}
	return nil
}
