package httpmiddleware

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"openmars/internal/delivery/dto"
	"openmars/internal/logger"
)

// Recovery middleware recovers from panics and returns a 500 error
func Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log := logger.Get()
				log.Error("Panic in HTTP handler",
					slog.Any("error", err),
					slog.String("method", r.Method),
					slog.String("path", r.URL.Path),
					slog.String("remote_addr", r.RemoteAddr))

				// Send error response
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)

				errorResponse := dto.ErrorPayload{
					Message: "Internal server error",
				}

				if err := json.NewEncoder(w).Encode(errorResponse); err != nil {
					log.Warn("Failed to encode error response", slog.Any("error", err))
				}
			}
		}()

		next.ServeHTTP(w, r)
	})
}
