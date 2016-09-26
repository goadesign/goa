// +build windows

package codegen

import (
	"os"
	"syscall"
)

var defaultSignals = []os.Signal{
	syscall.SIGHUP,
	syscall.SIGINT,
	syscall.SIGQUIT,
	syscall.SIGTERM,
}
