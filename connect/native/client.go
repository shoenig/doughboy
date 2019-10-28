package native

import (
	"io/ioutil"
	"net/http"
	"sync"
	"time"

	"github.com/hashicorp/consul/api"
	"github.com/hashicorp/consul/connect"
	"github.com/pkg/errors"

	"gophers.dev/pkgs/ignore"
	"gophers.dev/pkgs/loggy"
)

type Requester struct {
	consul *api.Client
	log    loggy.Logger

	lock      sync.Mutex
	service   *connect.Service
	shutdownC chan bool
}

func NewRequester(consul *api.Client) *Requester {
	return &Requester{
		consul: consul,
		log:    loggy.New("native-requester"),
	}
}

func (r *Requester) Open() error {
	r.lock.Lock()
	defer r.lock.Unlock()

	if r.service != nil {
		return errors.New("native requester already created")
	}

	service, err := connect.NewService(ConsulServiceName, r.consul)
	if err != nil {
		return errors.Wrap(err, "cannot create connect service")
	}
	r.service = service

	client := r.service.HTTPClient()
	r.shutdownC = make(chan bool)

	go func() {
		for {
			select {
			case <-r.shutdownC:
				r.log.Infof("shutting down")
				return
			case <-time.After(3 * time.Second):
				r.doRequest(client)
			}

		}
	}()

	return nil
}

func (r *Requester) doRequest(client *http.Client) {
	response, err := client.Get("https://doughboy-native-responder.service.consul/native/responder/poke")
	if err != nil {
		r.log.Errorf("client request failed: %v", err)
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

	if r.service == nil {
		return errors.New("native client not started")
	}

	r.shutdownC <- true

	r.service = nil
	r.shutdownC = nil

	return nil
}
