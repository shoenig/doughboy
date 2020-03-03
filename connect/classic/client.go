package classic

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/mux"
	clean "github.com/hashicorp/go-cleanhttp"
	"github.com/pkg/errors"

	"gophers.dev/cmds/doughboy/config"
	"gophers.dev/cmds/doughboy/connect/common"
	"gophers.dev/pkgs/ignore"
	"gophers.dev/pkgs/loggy"
)

type Requester struct {
	config config.ClassicServer
	log    loggy.Logger

	lock      sync.Mutex
	server    *http.Server // serve http HC
	shutdownC chan bool
}

func NewRequester(config config.ClassicServer) *Requester {
	return &Requester{
		config: config,
		log:    loggy.New("classic-requester"),
	}
}

func (r *Requester) Open() error {
	r.lock.Lock()
	defer r.lock.Unlock()

	if r.server != nil {
		return errors.New("classic requester already created")
	}

	r.server = &http.Server{
		Addr:      fmt.Sprintf("%s:%d", r.config.Bind, r.config.Port),
		TLSConfig: nil, // not TLS aware
		Handler:   r.mux(),
	}
	r.shutdownC = make(chan bool)

	client := clean.DefaultClient() // used to query upstream via connect
	client.Timeout = 1 * time.Second

	go func() {
		r.log.Infof("listen and serve beginning now ...")
		if err := r.server.ListenAndServe(); err != nil {
			panic(err)
		}
	}()

	go func() {
		for {
			// do requests forever until shutdown

			select {

			case <-r.shutdownC:
				r.log.Infof("shutting down")
				_ = r.server.Close()
				r.server = nil
				r.shutdownC = nil
				return

			case <-time.After(3 * time.Second):
				r.doRequest(client)
			}
		}
	}()

	r.log.Infof("requester is running ...")
	return nil
}

// Using Consul for service discovery of the upstream is not going to work from
// inside the network namespace that we create for ourselves. Instead we will just
// have to pass in the local bind port, as configured in the nomad job file.
//
//func (r *Requester) lookupUpstream(destination string) (string, error) {
//	sidecar := r.name + "-sidecar-proxy"
//
//	services, _, err := r.consul.Catalog().Service(sidecar, "", nil)
//	if err != nil {
//		return "", err
//	}
//
//	// we assume this doughboy-classic-requester is the only one registered
//	// a more complete implementation needs to reference the specific service ID
//	// of this running doughboy instance
//	if len(services) != 1 {
//		return "", errors.Errorf("failed to lookup %q", r.name)
//	}
//
//	for _, upstream := range services[0].ServiceProxy.Upstreams {
//		if upstream.DestinationName == destination {
//			return fmt.Sprintf("http://%s:%d/classic/responder/poke",
//				services[0].ServiceProxy.LocalServiceAddress,
//				services[0].ServiceProxy.Upstreams[0].LocalBindPort,
//			), nil
//		}
//	}
//
//	return "", errors.Errorf("no upstream of name %q", destination)
//}

// Just use the local bind port passed in through configuration, which is also
// configured in the nomad job file.
func (r *Requester) lookupUpstream(_ string) (string, error) {
	return fmt.Sprintf(
		"http://127.0.0.1:%d/classic/responder/poke",
		r.config.UpstreamPort,
	), nil
}

func (r *Requester) doRequest(client *http.Client) {
	upstreamURL, err := r.lookupUpstream("doughboy-classic-responder")
	if err != nil {
		r.log.Errorf("request failed: %v", err)
		return
	}

	response, err := client.Get(upstreamURL)
	if err != nil {
		r.log.Errorf("do GET request failed: %v", err)
		return
	}
	defer ignore.Drain(response.Body)

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		r.log.Errorf("unable to read response: %v", err)
		return
	}

	r.log.Infof("got response: %s", string(body))
}

func (r *Requester) Close() error {
	r.lock.Lock()
	defer r.lock.Unlock()

	if r.shutdownC == nil {
		return errors.New("classic client not started")
	}

	r.shutdownC <- true
	r.shutdownC = nil

	return nil
}

func (r *Requester) mux() http.Handler {
	router := mux.NewRouter()
	router.Handle("/classic/requester/health", common.HealthCheck(r.log))
	return router
}
