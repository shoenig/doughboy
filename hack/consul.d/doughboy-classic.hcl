services {
  name = "doughboy-classic-responder"
  port = 6100
  connect {
    native = false
    sidecar_service {
      // See https://www.consul.io/docs/connect/registration/sidecar-service.html#minimal-example
      // for more information about the implicit sidecar consul service definition.
    }
  }
}

services {
  name = "doughboy-classic-requester"
  port = 6110
  connect {
    native = false
    sidecar_service {
      proxy {
        upstreams = [{
          destination_name = "doughboy-classic-responder"
          local_bind_port = 6111 // sidecar binds to, for this upstream
        }]
      }
    }
  }
}

