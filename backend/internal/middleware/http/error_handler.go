package httpmiddleware

import (
	"encoding/json"
	"net/http"
	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/logger"

	"go.uber.org/zap"
)

// Recovery middleware recovers from panics and returns a 500 error
func Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log := logger.Get()
				log.Error("Panic in HTTP handler",
					zap.Any("error", err),
					zap.String("method", r.Method),
					zap.String("path", r.URL.Path),
					zap.String("remote_addr", r.RemoteAddr))

				// Send error response
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)

				errorResponse := dto.ErrorPayload{
					Message: "Internal server error",
				}

				if err := json.NewEncoder(w).Encode(errorResponse); err != nil {
					log.Warn("Failed to encode error response", zap.Error(err))
				}
			}
		}()

		next.ServeHTTP(w, r)
	})
}
