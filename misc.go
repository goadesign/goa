package goa

import (
	"fmt"
	"os"

	log "gopkg.in/inconshreveable/log15.v2"
)

// ToLogCtx converts the given map into a map of string to interface{} suitable to be passed
// to a log method as context.
func ToLogCtx(m map[string]string) log.Ctx {
	res := make(log.Ctx, len(m))
	for k, v := range m {
		res[k] = interface{}(v)
	}
	return res
}

// ToLogCtxA converts the given map into a map of string to interface{} suitable to be passed
// to a log method as context.
func ToLogCtxA(m map[string][]string) log.Ctx {
	res := make(log.Ctx, len(m))
	for k, v := range m {
		res[k] = interface{}(v)
	}
	return res
}

// fatalf prints the given error message and exits process with status 1.
func fatalf(format string, val ...interface{}) {
	fmt.Printf("goa: "+format, val)
	os.Exit(1)
}
