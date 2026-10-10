package http

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"openmars/internal/delivery/dto"
	"openmars/internal/logger"
)

// BaseHandler provides common functionality for all HTTP handlers
type BaseHandler struct {
	logger *slog.Logger
}

// NewBaseHandler creates a new base handler
func NewBaseHandler() *BaseHandler {
	return &BaseHandler{
		logger: logger.Get(),
	}
}

// WriteJSONResponse writes a JSON response to the client
func (h *BaseHandler) WriteJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.logger.Error("Failed to encode JSON response", slog.Any("error", err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// WriteErrorResponse writes an error response to the client
func (h *BaseHandler) WriteErrorResponse(w http.ResponseWriter, statusCode int, message string) {
	errorResponse := dto.ErrorPayload{
		Message: message,
	}
	h.WriteJSONResponse(w, statusCode, errorResponse)
}

// ParseJSONRequest parses a JSON request body into the provided struct
func (h *BaseHandler) ParseJSONRequest(r *http.Request, dest interface{}) error {
	if err := json.NewDecoder(r.Body).Decode(dest); err != nil {
		h.logger.Error("Failed to parse JSON request", slog.Any("error", err))
		return err
	}
	return nil
}
