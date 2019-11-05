package main

import (
	"os"

	"gophers.dev/cmds/doughboy/service"

	"gophers.dev/pkgs/loggy"
)

func main() {
	log := loggy.New("main")

	doughboy, err := service.New(os.Args[1:])
	if err != nil {
		log.Errorf("unable to start: %v", err)
		os.Exit(1)
	}

	if err := doughboy.Start(); err != nil {
		log.Errorf("program stopped: %v", err)
		os.Exit(1)
	}

	doughboy.Wait()
}
