package gin

import (
	"bufio"
	"net"
	"net/http"
)

const (
	noWritten     = -1
	defaultStatus = http.StatusOK
)

type responseWriter struct {
	http.ResponseWriter
	size   int
	status int
}

func (w *responseWriter) Unwrap() http.ResponseWriter {
	return w.ResponseWriter
}

// reset resets the responseWriter to its initial state.
// 把 responseWriter 重置为初始状态
func (w *responseWriter) reset(writer http.ResponseWriter) {
	w.ResponseWriter = writer
	w.size = noWritten
	w.status = defaultStatus
}

// CloseNotify implements ResponseWriter.
func (r *responseWriter) CloseNotify() <-chan bool {
	panic("unimplemented")
}

// Flush implements ResponseWriter.
func (r *responseWriter) Flush() {
	panic("unimplemented")
}

// Header implements ResponseWriter.
// Subtle: this method shadows the method (ResponseWriter).Header of responseWriter.ResponseWriter.
func (r *responseWriter) Header() http.Header {
	panic("unimplemented")
}

// Hijack implements ResponseWriter.
func (r *responseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	panic("unimplemented")
}

// Pusher implements ResponseWriter.
func (r *responseWriter) Pusher() http.Pusher {
	panic("unimplemented")
}

// Size implements ResponseWriter.
func (r *responseWriter) Size() int {
	panic("unimplemented")
}

// Status implements ResponseWriter.
func (r *responseWriter) Status() int {
	panic("unimplemented")
}

// Write implements ResponseWriter.
// Subtle: this method shadows the method (ResponseWriter).Write of responseWriter.ResponseWriter.
func (r *responseWriter) Write([]byte) (int, error) {
	panic("unimplemented")
}

// WriteHeader implements ResponseWriter.
// Subtle: this method shadows the method (ResponseWriter).WriteHeader of responseWriter.ResponseWriter.
func (r *responseWriter) WriteHeader(statusCode int) {
	panic("unimplemented")
}

// WriteHeaderNow implements ResponseWriter.
func (r *responseWriter) WriteHeaderNow() {
	panic("unimplemented")
}

// WriteString implements ResponseWriter.
func (r *responseWriter) WriteString(string) (int, error) {
	panic("unimplemented")
}

// Written implements ResponseWriter.
func (r *responseWriter) Written() bool {
	panic("unimplemented")
}

// ResponseWriter ...
type ResponseWriter interface {
	http.ResponseWriter
	http.Hijacker
	http.Flusher
	http.CloseNotifier

	// Status returns the HTTP response status code of the current request.
	Status() int

	// Size returns the number of bytes already written into the response http body.
	// See Written()
	Size() int

	// WriteString writes the string into the response body.
	WriteString(string) (int, error)

	// Written returns true if the response body was already written.
	Written() bool

	// WriteHeaderNow forces to write the http header (status code + headers).
	WriteHeaderNow()

	// Pusher get the http.Pusher for server push
	Pusher() http.Pusher
}
