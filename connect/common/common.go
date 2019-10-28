package common

import (
	"io"
	"net/http"

	"gophers.dev/pkgs/loggy"
)

func HealthCheck(logger loggy.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger.Tracef("health-check requested from %s", r.RemoteAddr)

		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, "ok\n")
	}
}
