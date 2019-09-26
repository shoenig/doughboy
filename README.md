doughboy
========

Responds to pokes of signal interrupts, http, etc.

[![Go Report Card](https://goreportcard.com/badge/gophers.dev/cmds/doughboy)](https://goreportcard.com/report/gophers.dev/cmds/doughboy)
[![Build Status](https://travis-ci.com/shoenig/doughboy.svg?branch=master)](https://travis-ci.com/shoenig/doughboy)
[![GoDoc](https://godoc.org/gophers.dev/cmds/doughboy?status.svg)](https://godoc.org/gophers.dev/cmds/doughboy)
[![NetflixOSS Lifecycle](https://img.shields.io/osslifecycle/shoenig/doughboy.svg)](OSSMETADATA)
[![GitHub](https://img.shields.io/github/license/shoenig/doughboy.svg)](LICENSE)

# Project Overview

Module `gophers.dev/cmds/doughboy` responds to a configurable set of signal
interrupts. In the future it could respond to other things like HTTP requests
in configurable ways.

# Getting Started

The `doughboy` command can be installed by running
```bash
$ go get gophers.dev/cmds/doughboy
```

# Example Usage

Build and run the basic example
```bash
$ go build
$ ./doughboy hack/example.hcl
2019/09/25 21:52:19 INFO  [doughboy] service.address: 127.0.0.1
2019/09/25 21:52:19 INFO  [doughboy] service.port: 1234
2019/09/25 21:52:19 INFO  [doughboy] signals: [sigterm sigusr1]
2019/09/25 21:52:19 INFO  [doughboy] starting up!
2019/09/25 21:52:19 INFO  [doughboy] will listen at 127.0.0.1:1234
2019/09/25 21:52:19 INFO  [doughboy] pid: 50444
```

Poke it with a signal
```bash
$ pkill -SIGTERM doughboy
```

Notice the `doughboy` process intercepts the interrupt without dying
```
2019/09/25 21:53:14 INFO  [signals] received <terminated> @ 21:53:14.039
```

# Configuration

See [example.hcl](hack/example.hcl) for an example configuration file.

# Contributing

The `gophers.dev/cmds/doughboy` module is always improving with new features
and error corrections. For contributing bug fixes and new features please file an issue.

# License

The `gophers.dev/cmds/doughboy` module is open source under the [BSD-3-Clause](LICENSE) license.
