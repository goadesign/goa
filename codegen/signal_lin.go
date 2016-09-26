// +build !windows

package codegen

// Avoid importing syscall to play nice with e.g. Appengine
const (
	SIGHUP  = 0x1
	SIGINT  = 0x2
	SIGQUIT = 0x3
	SIGTERM = 0xf
	SIGUSR1 = 0xa
	SIGUSR2 = 0xc
)
