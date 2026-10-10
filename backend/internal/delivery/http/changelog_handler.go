package http

import (
	"net/http"

	"openmars/internal/delivery/dto"
)

// ChangelogHandler serves the player release notes this server was built with
type ChangelogHandler struct {
	*BaseHandler
	changelog dto.ChangelogResponse
}

// NewChangelogHandler creates a new changelog handler
func NewChangelogHandler(changelog dto.ChangelogResponse) *ChangelogHandler {
	return &ChangelogHandler{
		BaseHandler: NewBaseHandler(),
		changelog:   changelog,
	}
}

// Changelog returns the player release notes of every published version, newest first
func (h *ChangelogHandler) Changelog(w http.ResponseWriter, r *http.Request) {
	h.WriteJSONResponse(w, http.StatusOK, h.changelog)
}
