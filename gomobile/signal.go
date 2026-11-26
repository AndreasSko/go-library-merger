package gomobile

import (
	"os/signal"
	"syscall"
)

// On iOS, Go ignores SIGPIPE (a write to a closed connection just fails), but
// does so in its own signal handler. Sentry reports a SIGPIPE as a crash
// unless the signal is set to SIG_IGN when Sentry starts. The app sets SIG_IGN
// before starting Sentry, but the Go runtime starts on its own thread and can
// install its handler in between. Ignoring the signal here, after the runtime
// installed its handlers, makes SIG_IGN stick in any order.
func init() {
	signal.Ignore(syscall.SIGPIPE)
}
