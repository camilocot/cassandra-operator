package probe

import (
	"net/http"
	"sync"
)

const (
	// HTTPReadyzEndpoint is the ready endpoint path
	HTTPReadyzEndpoint = "/readyz"
)

var (
	mu    sync.Mutex
	ready = false
)

// SetReady set ready to true using a exclusion lock
func SetReady() {
	mu.Lock()
	ready = true
	mu.Unlock()
}

// ReadyzHandler writes back the HTTP status code 200 if the operator is ready, and 500 otherwise
func ReadyzHandler(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	isReady := ready
	mu.Unlock()
	if isReady {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusInternalServerError)
	}
}
