// +build !windows,!appengine

package utils

import (
	"os"
	"syscall"
)

var defaultSignals = []os.Signal{
	syscall.SIGHUP,
	syscall.SIGINT,
	syscall.SIGQUIT,
	syscall.SIGTERM,
	syscall.SIGUSR1,
	syscall.SIGUSR2,
}
