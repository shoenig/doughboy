package classic

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/hashicorp/consul/api"
	clean "github.com/hashicorp/go-cleanhttp"
	"github.com/pkg/errors"

	"gophers.dev/cmds/doughboy/config"
	"gophers.dev/cmds/doughboy/connect/common"
	"gophers.dev/pkgs/ignore"
	"gophers.dev/pkgs/loggy"
)

type Requester struct {
	name   string
	config config.Server
	consul *api.Client
	log    loggy.Logger

	lock      sync.Mutex
	server    *http.Server // serve HC
	shutdownC chan bool
}

func NewRequester(config config.Server, consul *api.Client) *Requester {
	return &Requester{
		name:   "doughboy-classic-requester",
		consul: consul,
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

	client := clean.DefaultClient()
	client.Timeout = 1 * time.Second

	r.shutdownC = make(chan bool)

	go func() {
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

	return nil
}

func (r *Requester) lookupUpstream(destination string) (string, error) {
	sidecar := r.name + "-sidecar-proxy"

	services, _, err := r.consul.Catalog().Service(sidecar, "", nil)
	if err != nil {
		return "", err
	}

	// we assume this doughboy-classic-requester is the only one registered
	// a more complete implementation needs to reference the specific service ID
	// of this running doughboy instance
	if len(services) != 1 {
		return "", errors.Errorf("failed to lookup %q", r.name)
	}

	fmt.Printf("service[0]: %#v\n", services[0])
	fmt.Printf("service[0].ServiceProxy: %#v\n", services[0].ServiceProxy)

	for _, upstream := range services[0].ServiceProxy.Upstreams {
		fmt.Printf("upstream: %#v\n", upstream)

		if upstream.DestinationName == destination {
			return fmt.Sprintf("http://%s:%d/classic/responder/poke",
				services[0].ServiceProxy.LocalServiceAddress,
				services[0].ServiceProxy.Upstreams[0].LocalBindPort,
			), nil
		}
	}

	return "", errors.Errorf("no upstream of name %q", destination)
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
