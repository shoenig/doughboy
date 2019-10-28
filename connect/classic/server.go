package classic

import (
	"errors"
	"fmt"
	"net/http"
	"sync"

	"github.com/gorilla/mux"

	"gophers.dev/cmds/doughboy/config"
	"gophers.dev/cmds/doughboy/connect/common"
	"gophers.dev/pkgs/loggy"
)

type Responder struct {
	config config.Server
	log    loggy.Logger

	lock      sync.Mutex
	server    *http.Server
	shutdownC chan bool
}

func NewResponder(config config.Server) *Responder {
	return &Responder{
		config: config,
		log:    loggy.New("classic-responder"),
	}
}

func (r *Responder) Open() error {
	r.lock.Lock()
	defer r.lock.Unlock()

	if r.server != nil {
		return errors.New("classic responder already started")
	}

	r.server = &http.Server{
		Addr:      fmt.Sprintf("%s:%d", r.config.Bind, r.config.Port),
		TLSConfig: nil, // not TLS aware
		Handler:   r.mux(),
	}
	r.shutdownC = make(chan bool)

	go func() {
		if err := r.server.ListenAndServe(); err != nil {
			panic(err)
		}
	}()

	go func() {
		for {
			// respond to requests forever until shutdown

			select {

			case <-r.shutdownC:
				r.log.Infof("shutting down")
				_ = r.server.Close()
				r.server = nil
				r.shutdownC = nil
				return
			}
		}
	}()

	return nil
}

func (r *Responder) Close() error {
	r.lock.Lock()
	defer r.lock.Unlock()

	if r.server == nil {
		return errors.New("classic responder not started")
	}

	r.shutdownC <- true

	r.server = nil
	r.shutdownC = nil

	return nil
}

func (r *Responder) mux() http.Handler {
	router := mux.NewRouter()
	router.Handle("/classic/responder/poke", r.handler())
	router.Handle("/classic/responder/health", common.HealthCheck(r.log))
	return router
}

func (r *Responder) handler() http.HandlerFunc {
	return func(w http.ResponseWriter, rq *http.Request) {
		r.log.Infof("got a request from %s", rq.RemoteAddr)
		_, _ = w.Write([]byte("this is a response\n"))
	}
}
