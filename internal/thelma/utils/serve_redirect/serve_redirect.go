package serve_redirect

import (
	"errors"
	"fmt"
	"github.com/rs/zerolog/log"
	"net/http"
	"sync"
)

var localServerMutex sync.Mutex

// ServeRedirect starts a local HTTP server to receive an OAuth2 authorization code at the given localhost port and
// path. It accepts a callback to store the received and validated code.
//
// This function returns a state that must be included in the initial OAuth2 request and a function that must be called
// to close the server. If an error is returned, no server will have been started and the close function will be nil.
func ServeRedirect(port int, path string, callback func(code string)) (state string, closeFunc func(), err error) {
	state, err = generateState()
	if err != nil {
		return "", nil, err
	}

	mux := http.NewServeMux()
	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("code") == "" {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = fmt.Fprintf(w, "%d - no code in request", http.StatusBadRequest)
		} else {
			log.Debug().Msg("Received redirect with authorization code")
			if r.URL.Query().Get("state") != state {
				w.WriteHeader(http.StatusConflict)
				_, _ = fmt.Fprintf(w, "%d - bad state", http.StatusConflict)
			} else {
				w.WriteHeader(http.StatusOK)
				_, _ = fmt.Fprintf(w, "Success! You can close this window.")
				callback(r.URL.Query().Get("code"))
			}
		}
	})
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: mux,
	}
	localServerMutex.Lock()
	go func() {
		log.Debug().Msgf("Starting authorization code server on :%d%s", port, path)
		if srvErr := server.ListenAndServe(); !errors.Is(srvErr, http.ErrServerClosed) {
			log.Error().Err(srvErr).Msg("Failed to start server")
		}
	}()

	return state, func() {
		if err := server.Close(); err != nil {
			log.Error().Err(err).Msg("Failed to close server")
		} else {
			log.Debug().Msgf("Shutting down authorization code server on :%d%s", port, path)
			localServerMutex.Unlock()
		}
	}, nil
}
