package google

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
)

// loggingTransport an http transport that logs requests/responses to stderr
// maybe one day Thelma will have a real logging transport for its http clients. Until then, this is helpful.
//
// adapted in hacky fashion from https://github.com/googleapis/google-api-go-client/blob/9f186713659dc989b56a801e556a563954ecb673/examples/debug.go
type loggingTransport struct {
	rt http.RoundTripper
}

func (t *loggingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	var buf bytes.Buffer

	_, _ = os.Stderr.Write([]byte("\n[request]\n"))
	if req.Body != nil {
		req.Body = io.NopCloser(&readButCopy{req.Body, &buf})
	}
	_ = req.Write(os.Stdout)
	if req.Body != nil {
		req.Body = io.NopCloser(&buf)
	}
	_, _ = os.Stderr.Write([]byte("\n[/request]\n"))

	res, err := t.rt.RoundTrip(req)

	fmt.Printf("[response]\n")
	if err != nil {
		fmt.Printf("ERROR: %v", err)
	} else {
		body := res.Body
		res.Body = nil
		_ = res.Write(os.Stdout)
		if body != nil {
			res.Body = io.NopCloser(&echoAsRead{body})
		}
	}

	return res, err
}

type echoAsRead struct {
	src io.Reader
}

func (r *echoAsRead) Read(p []byte) (int, error) {
	n, err := r.src.Read(p)
	if n > 0 {
		_, _ = os.Stdout.Write(p[:n])
	}
	if err == io.EOF {
		fmt.Printf("\n[/response]\n")
	}
	return n, err
}

type readButCopy struct {
	src io.Reader
	dst io.Writer
}

func (r *readButCopy) Read(p []byte) (int, error) {
	n, err := r.src.Read(p)
	if n > 0 {
		_, _ = r.dst.Write(p[:n])
	}
	return n, err
}
