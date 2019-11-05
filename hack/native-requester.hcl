signals = ["sigterm"]

consul {
  port = 8500
}

echo {
  native {
    client {
      enable = true
    }
  }
}
