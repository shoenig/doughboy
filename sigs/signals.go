package sigs

import (
	"errors"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"gophers.dev/pkgs/loggy"

	"oss.indeed.com/go/libtime"
)

type Watcher struct {
	clock libtime.Clock
	log   loggy.Logger
}

var conversions = map[string]os.Signal{
	//1) SIGHUP	     2) SIGINT	 3) SIGQUIT	 4) SIGILL
	//5) SIGTRAP	 6) SIGABRT	 7) SIGEMT	 8) SIGFPE
	//9) SIGKILL	10) SIGBUS	11) SIGSEGV	12) SIGSYS
	//13) SIGPIPE	14) SIGALRM	15) SIGTERM	16) SIGURG
	//17) SIGSTOP	18) SIGTSTP	19) SIGCONT	20) SIGCHLD
	//21) SIGTTIN	22) SIGTTOU	23) SIGIO	24) SIGXCPU
	//25) SIGXFSZ	26) SIGVTALRM	27) SIGPROF	28) SIGWINCH
	//29) SIGINFO	30) SIGUSR1	31) SIGUSR2

	"sigterm": syscall.SIGTERM,
	"sigusr1": syscall.SIGUSR1,
}

func lookup(name string) (os.Signal, error) {
	lower := strings.ToLower(name)
	s, ok := conversions[lower]
	if !ok {
		return syscall.Signal(0), errors.New("not a known signal")
	}
	return s, nil
}

func Lookup(names ...string) ([]os.Signal, error) {
	signals := make([]os.Signal, 0, len(names))
	for _, name := range names {
		s, err := lookup(name)
		if err != nil {
			return nil, err
		}
		signals = append(signals, s)
	}
	return signals, nil
}

func New(clock libtime.Clock) *Watcher {
	return &Watcher{
		clock: clock,
		log:   loggy.New("signals"),
	}
}

func (sw *Watcher) Watch(signals ...os.Signal) {
	sigC := make(chan os.Signal, 1)
	signal.Notify(sigC, signals...)

	for {
		rxSignal := <-sigC
		now := sw.clock.Now().Format("15:04:05.000")
		sw.log.Infof("received <%s> @ %s", rxSignal, now)
	}
}
