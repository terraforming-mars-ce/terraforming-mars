package http

import (
	"net/http"

	"openmars/internal/delivery/dto"
)

// MetaHandler serves the identity a gateway uses to list and verify this server
type MetaHandler struct {
	*BaseHandler
	meta dto.MetaResponse
}

// NewMetaHandler creates a new meta handler
func NewMetaHandler(meta dto.MetaResponse) *MetaHandler {
	return &MetaHandler{
		BaseHandler: NewBaseHandler(),
		meta:        meta,
	}
}

// Meta returns this server's alias, display name and version
func (h *MetaHandler) Meta(w http.ResponseWriter, r *http.Request) {
	h.WriteJSONResponse(w, http.StatusOK, h.meta)
}
