signals = ["sigterm"]

consul {
  port = 8500
}

echo {
  classic {
    client {
      server {
        enable = true
        bind = "127.0.0.1" # inside the network namespace
        port = 6110 # listens http, not connect enabled
      }
      upstream_port = 6111 # as configured in nomad job connect stanza
    }
  }
}
