signals = ["sigterm"]

consul {
  port = 8500
}

echo {
  classic {
    server {
      enable = true
      bind = "127.0.0.1"
      port = 6100
    }
  }
}
