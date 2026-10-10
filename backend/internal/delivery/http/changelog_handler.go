package http

import (
	"net/http"
	"os"
	"path/filepath"

	"openmars/internal/delivery/dto"

	"github.com/gorilla/mux"
)

// ChangelogHandler serves the player release notes this server was built with,
// and the images they reference from changelog/<version>/
type ChangelogHandler struct {
	*BaseHandler
	dir       string
	changelog dto.ChangelogResponse
	images    map[string]struct{}
}

// NewChangelogHandler creates a new changelog handler. Only images named in
// the notes are served.
func NewChangelogHandler(dir string, changelog dto.ChangelogResponse) *ChangelogHandler {
	images := make(map[string]struct{})
	for _, entry := range changelog.Entries {
		for _, section := range entry.Sections {
			if section.Image != nil {
				images[entry.Version+"/"+section.Image.File] = struct{}{}
			}
		}
	}
	return &ChangelogHandler{
		BaseHandler: NewBaseHandler(),
		dir:         dir,
		changelog:   changelog,
		images:      images,
	}
}

// Changelog returns the player release notes of every published version, newest first
func (h *ChangelogHandler) Changelog(w http.ResponseWriter, r *http.Request) {
	h.WriteJSONResponse(w, http.StatusOK, h.changelog)
}

// Image serves an image referenced by a version's player release notes
func (h *ChangelogHandler) Image(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	version, file := vars["version"], vars["file"]
	if _, ok := h.images[version+"/"+file]; !ok {
		h.WriteErrorResponse(w, http.StatusNotFound, "Image not found")
		return
	}

	image, err := os.Open(filepath.Join(h.dir, version, file))
	if err != nil {
		h.WriteErrorResponse(w, http.StatusNotFound, "Image not found")
		return
	}
	defer func() { _ = image.Close() }()
	info, err := image.Stat()
	if err != nil {
		h.WriteErrorResponse(w, http.StatusInternalServerError, "Failed to read image")
		return
	}

	w.Header().Set("Cache-Control", "public, max-age=86400")
	http.ServeContent(w, r, file, info.ModTime(), image)
}
