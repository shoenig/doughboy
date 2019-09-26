package service

import (
	"fmt"
	"net/http"
	"os"

	"github.com/pkg/errors"

	"gophers.dev/cmds/doughboy/config"
	"gophers.dev/cmds/doughboy/sigs"

	"gophers.dev/pkgs/loggy"

	"oss.indeed.com/go/libtime"
)

type DoughboyService struct {
	configuration *config.Configuration
	log           loggy.Logger
}

func New(args []string) (*DoughboyService, error) {
	if len(args) != 1 {
		return nil, errors.New("config file required")
	}

	log := loggy.New("doughboy")

	c, err := config.LoadHCL(os.Args[1])
	if err != nil {
		return nil, errors.Wrap(err, "malformed config file")
	}

	c.Log(log)

	return &DoughboyService{
		configuration: c,
		log:           log,
	}, nil
}

func (ds *DoughboyService) Start() error {
	ds.log.Infof("starting up!")

	address := fmt.Sprintf(
		"%s:%d",
		ds.configuration.Service.Address,
		ds.configuration.Service.Port,
	)
	ds.log.Infof("will listen at %s", address)

	pid := os.Getpid()
	ds.log.Infof("pid: %d", pid)

	signals, err := sigs.Lookup(ds.configuration.Signals...)
	if err != nil {
		return err
	}

	sigWatcher := sigs.New(libtime.SystemClock())
	go sigWatcher.Watch(signals...)

	if err := http.ListenAndServe(address, makeAPI(loggy.New("api"))); err != nil {
		ds.log.Errorf("failed to listen and serve: %v", err)
		return errors.Wrap(err, "unable to listen and serve")
	}

	return errors.New("stopped listening (!)")
}

func makeAPI(log loggy.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		msg := fmt.Sprintf("request from %s for %s", r.RemoteAddr, r.RequestURI)
		log.Infof(msg)
		_, _ = fmt.Fprintf(w, msg+"\n")
	}
}
