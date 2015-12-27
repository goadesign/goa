package utils

import (
	"os"
	"os/signal"
)

// Catch signals and invoke then callback
func Catch(signals []os.Signal, then func()) {
	c := make(chan os.Signal)
	if signals == nil {
		signals = defaultSignals
	}
	signal.Notify(c, signals...)
	<-c
	if then != nil {
		then()
	}
}
