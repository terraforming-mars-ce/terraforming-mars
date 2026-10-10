package http

import (
	"net/http"
)

// HealthHandler handles HTTP health check requests
type HealthHandler struct {
	*BaseHandler
}

// NewHealthHandler creates a new health handler
func NewHealthHandler() *HealthHandler {
	return &HealthHandler{
		BaseHandler: NewBaseHandler(),
	}
}

// HealthCheck returns the health status of the service
func (h *HealthHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	response := map[string]string{
		"status":  "healthy",
		"service": "openmars",
	}

	h.WriteJSONResponse(w, http.StatusOK, response)
}
