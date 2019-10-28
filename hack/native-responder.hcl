signals = ["sigterm"]

consul {
  port = 8500
}

echo {
  native {
    server {
      enable = true
      bind = "127.0.0.1"
      port = 5000
    }
  }
}
