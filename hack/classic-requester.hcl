signals = ["sigterm"]

consul {
  port = 8500
}

echo {
  classic {
    client {
      enable = true
      bind = "127.0.0.1"
      port = 6110
    }
  }
}
