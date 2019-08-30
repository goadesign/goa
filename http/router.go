package http

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"runtime"
)

type route struct {
	method string
	path   string
}

var nullLogger *log.Logger
var loadTestHandler = false

type mockResponseWriter struct{}

func (m *mockResponseWriter) Header() (h http.Header) {
	return http.Header{}
}

func (m *mockResponseWriter) Write(p []byte) (n int, err error) {
	return len(p), nil
}

func (m *mockResponseWriter) WriteString(s string) (n int, err error) {
	return len(s), nil
}
func (m *mockResponseWriter) WriteHeader(int) {}
func init() {

	runtime.GOMAXPROCS(1)

	// makes logging 'webscale' (ignores them)
	log.SetOutput(new(mockResponseWriter))
	nullLogger = log.New(new(mockResponseWriter), "", 0)

	// initBeego()
	// initGin()
	// initRevel()
}

// func goaHandler() http.HandleFunc {
// 	return nil
// }
// var mux Muxer

func httpHandlerFunc(w http.ResponseWriter, r *http.Request) {
}
func httpHandlerWrite(w http.ResponseWriter, r *http.Request) {
	// var (
	mux := NewMuxer()
	params := mux.Vars(r)
	// )
	// params := Muxer.Vars(r)
	fmt.Fprintf(w, params["name"])
}
func httpHandlerFuncTest(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, r.RequestURI)
}
func main() {
	fmt.Println("Usage: go test -bench=. -timeout=20m")
	os.Exit(1)
}
func loadGoa(routes []route) http.Handler {
	mux := NewMuxer()
	for _, route := range routes {
		mux.Handle(route.method, route.path, httpHandlerFunc)
	}
	return mux
}
func loadGoaSingle(method, path string, handler http.HandlerFunc) http.Handler {
	mux := NewMuxer()
	// for _, route := range routes {
	mux.Handle(method, path, handler)
	// }
	return mux
}

// var mux Muxer
// {
// 	mux = Muxer.NewMuxer()
// }
// // func goa() *Muxer{
// // 		return
// // }
