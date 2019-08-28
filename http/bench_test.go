package http

import (
	"net/http"
	"os"
	"regexp"
	"runtime"
	"strings"
	"testing"
)

var (
	goahttp http.Handler
)
var benchRe *regexp.Regexp

const fiveColon = "/:a/:b/:c/:d/:e"
const fiveBrace = "/{a}/{b}/{c}/{d}/{e}"
const fiveRoute = "/test/test/test/test/test"

const twentyColon = "/:a/:b/:c/:d/:e/:f/:g/:h/:i/:j/:k/:l/:m/:n/:o/:p/:q/:r/:s/:t"
const twentyBrace = "/{a}/{b}/{c}/{d}/{e}/{f}/{g}/{h}/{i}/{j}/{k}/{l}/{m}/{n}/{o}/{p}/{q}/{r}/{s}/{t}"
const twentyRoute = "/a/b/c/d/e/f/g/h/i/j/k/l/m/n/o/p/q/r/s/t"

func isTested(name string) bool {
	if benchRe == nil {
		// Get -test.bench flag value (not accessible via flag package)
		bench := ""
		for _, arg := range os.Args {
			if strings.HasPrefix(arg, "-test.bench=") {
				// ignore the benchmark name after an underscore
				bench = strings.SplitN(arg[12:], "_", 2)[0]
				break
			}
		}

		// Compile RegExp to match Benchmark names
		var err error
		benchRe, err = regexp.Compile(bench)
		if err != nil {
			panic(err.Error())
		}
	}
	return benchRe.MatchString(name)
}

func calcMem(name string, load func()) {
	if !isTested(name) {
		return
	}

	m := new(runtime.MemStats)

	// before
	runtime.GC()
	runtime.ReadMemStats(m)
	before := m.HeapAlloc

	load()

	// after
	runtime.GC()
	runtime.ReadMemStats(m)
	after := m.HeapAlloc
	println("   "+name+":", after-before, "Bytes")
}

func benchRequest(b *testing.B, router http.Handler, r *http.Request) {
	w := new(mockResponseWriter)
	u := r.URL
	rq := u.RawQuery
	r.RequestURI = u.RequestURI()

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		u.RawQuery = rq
		router.ServeHTTP(w, r)
	}
}
func benchRoutes(b *testing.B, router http.Handler, routes []route) {
	w := new(mockResponseWriter)
	r, _ := http.NewRequest("GET", "/", nil)
	u := r.URL
	rq := u.RawQuery

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		for _, route := range routes {
			r.Method = route.method
			r.RequestURI = route.path
			u.Path = route.path
			u.RawQuery = rq
			router.ServeHTTP(w, r)
		}
	}
}
func BenchmarkGoa_Param(b *testing.B) {
	goahttp = loadGoaSingle("GET", "/user/:name", httpHandlerFunc)
	r, _ := http.NewRequest("GET", "/user/gordon", nil)
	benchRequest(b, goahttp, r)
}
func BenchmarkGoa_Param5(b *testing.B) {
	router := loadGoaSingle("GET", fiveColon, httpHandlerFunc)

	r, _ := http.NewRequest("GET", fiveRoute, nil)
	benchRequest(b, router, r)
}
func BenchmarkGoa_Param20(b *testing.B) {
	router := loadGoaSingle("GET", twentyColon, httpHandlerFunc)

	r, _ := http.NewRequest("GET", twentyRoute, nil)
	benchRequest(b, router, r)
}

func BenchmarkGoa_ParamWrite(b *testing.B) {
	router := loadGoaSingle("GET", "/user/:name", httpHandlerWrite)

	r, _ := http.NewRequest("GET", "/user/gordon", nil)
	benchRequest(b, router, r)
}
