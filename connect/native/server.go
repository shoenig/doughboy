package native

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/gorilla/mux"
	"github.com/hashicorp/consul/api"
	"github.com/hashicorp/consul/connect"
	"github.com/pkg/errors"

	"gophers.dev/cmds/doughboy/config"
	"gophers.dev/cmds/doughboy/connect/common"
	"gophers.dev/pkgs/loggy"
)

const ConsulServiceName = "doughboy-native-responder"

type Responder struct {
	consul *api.Client
	config config.Server
	log    loggy.Logger

	lock       sync.Mutex
	service    *connect.Service
	httpServer *http.Server
}

func NewResponder(config config.Server, consul *api.Client) *Responder {
	return &Responder{
		consul: consul,
		config: config,
		log:    loggy.New("native-responder"),
	}
}

func (r *Responder) Open() error {
	r.lock.Lock()
	defer r.lock.Unlock()

	if r.service != nil {
		return errors.New("echo server already started")
	}

	service, err := connect.NewService(ConsulServiceName, r.consul)
	if err != nil {
		return errors.Wrap(err, "cannot create connect service")
	}
	r.service = service

	httpServer := &http.Server{
		Addr:      fmt.Sprintf(":%d", r.config.Port),
		TLSConfig: r.service.ServerTLSConfig(),
		Handler:   r.mux(),
	}

	go func() {
		if err := httpServer.ListenAndServeTLS("", ""); err != nil {
			panic(err)
		}
	}()

	return nil
}

func (r *Responder) mux() http.Handler {
	router := mux.NewRouter()
	router.Handle("/native/responder/poke", r.handler())
	router.Handle("/native/responder/health", common.HealthCheck(r.log))
	return router
}

func (r *Responder) handler() http.HandlerFunc {
	return func(w http.ResponseWriter, rq *http.Request) {
		r.log.Infof("got a request from %s", rq.RemoteAddr)
		_, _ = w.Write([]byte("this is a response\n"))
	}
}

func (r *Responder) Close() error {
	r.lock.Lock()
	defer r.lock.Unlock()

	if r.httpServer == nil {
		return errors.New("echo server not started")
	}

	if err := r.httpServer.Close(); err != nil {
		return errors.Wrap(err, "could not close connect http server")
	}

	if err := r.service.Close(); err != nil {
		return errors.Wrap(err, "could not close connect service")
	}

	r.service = nil

	return nil
}
