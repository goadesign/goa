package goa

import (
	"io"
	"sync"
	"sync/atomic"
)

// SkipResponseWriter converts an io.WriterTo into a io.ReadCloser.
// The Read/Close methods this function returns will pipe the Write calls that wt makes, to implement a Reader that has the written bytes.
// If Read is called Close must also be called to avoid leaking memory.
// The returned value implements io.WriterTo as well, so the generated handler will call that instead of the Read method.
//
// Server handlers that use SkipResponseBodyEncodeDecode() io.ReadCloser as a return type.
func SkipResponseWriter(wt io.WriterTo) io.ReadCloser {
	return &writerToReaderAdapter{WriterTo: wt}
}

type writerToReaderAdapter struct {
	io.WriterTo
	prOnce sync.Once
	pr     *io.PipeReader
}

func (a *writerToReaderAdapter) initPipe() {
	r, w := io.Pipe()
	go func() {
		_, err := a.WriteTo(w)
		w.CloseWithError(err)
	}()
	a.pr = r
}

func (a *writerToReaderAdapter) Read(b []byte) (n int, err error) {
	a.prOnce.Do(a.initPipe)
	return a.pr.Read(b)
}

func (a *writerToReaderAdapter) Close() error {
	a.prOnce.Do(a.initPipe)
	return a.pr.Close()
}

type writeCounter struct {
	io.Writer
	n atomic.Int64
}

func (wc *writeCounter) Count() int64 { return wc.n.Load() }
func (wc *writeCounter) Write(b []byte) (n int, err error) {
	n, err = wc.Writer.Write(b)
	wc.n.Add(int64(n))
	return
}

// WriterToFunc implements [io.WriterTo]. The io.Writer passed to the function will be wrapped.
type WriterToFunc func(w io.Writer) (err error)

// WriteTo writes to w.
//
// The value in w is wrapped when passed to fn keeping track of how bytes are written by fn.
func (fn WriterToFunc) WriteTo(w io.Writer) (n int64, err error) {
	wc := writeCounter{Writer: w}
	err = fn(&wc)
	return wc.Count(), err
}
