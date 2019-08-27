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
	ginQuery   http.Handler
	echoQuery  http.Handler
	beegoQuery http.Handler
	bmuxQuery  http.Handler
)
var benchRe *regexp.Regexp

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
func init() {

}
func benchRequest(b *testing.B, router Muxer, r *http.Request) {
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
func BenchmarkGoa_Param(b *testing.B) {
	// f, ok := http.HandleFunc.(http.HandlerFunc)
	// if !ok {
	// 	f = func(w http.ResponseWriter, r *http.Request) {
	// 		http.HandleFunc.ServeHTTP(w, r)
	// 	}
	// }
	// mux.Handle("GET", "/add/{a}/{b}", f)
	mux := NewMuxer()
	// mux.Handle("GET", "/add/{a}/{b}", nil)
	r, _ := http.NewRequest("GET", "/add/{a}/{b}", nil)
	benchRequest(b, mux, r)
}
