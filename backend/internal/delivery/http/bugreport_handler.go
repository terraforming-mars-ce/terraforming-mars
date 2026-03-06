package http

import (
	"net/http"

	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/service/bugreport"

	"github.com/gorilla/mux"
)

// BugReportHandler handles HTTP requests for bug reports.
type BugReportHandler struct {
	*BaseHandler
	service *bugreport.Service
}

// NewBugReportHandler creates a new bug report handler.
func NewBugReportHandler(service *bugreport.Service) *BugReportHandler {
	return &BugReportHandler{
		BaseHandler: NewBaseHandler(),
		service:     service,
	}
}

// GetStatus handles GET /api/v1/bugs/status.
func (h *BugReportHandler) GetStatus(w http.ResponseWriter, r *http.Request) {
	h.logger.Debug("GET /api/v1/bugs/status")

	caps := h.service.Capabilities()
	resp := dto.BugReportStatusResponse{
		Available: h.service.IsAvailable(),
		Claude:    caps.Claude,
	}
	if !resp.Available {
		resp.Reason = "GitHub App not configured"
	}
	h.WriteJSONResponse(w, http.StatusOK, resp)
}

// SubmitBugReport handles POST /api/v1/bugs.
func (h *BugReportHandler) SubmitBugReport(w http.ResponseWriter, r *http.Request) {
	h.logger.Debug("POST /api/v1/bugs")

	if !h.service.IsAvailable() {
		h.WriteErrorResponse(w, http.StatusServiceUnavailable, "Bug reporting is not available")
		return
	}

	var req dto.BugReportRequest
	if err := h.ParseJSONRequest(r, &req); err != nil {
		h.WriteErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if len(req.Description) == 0 {
		h.WriteErrorResponse(w, http.StatusBadRequest, "Description is required")
		return
	}

	if len(req.Description) > 30000 {
		h.WriteErrorResponse(w, http.StatusBadRequest, "Description exceeds maximum length of 30000 characters")
		return
	}

	author := req.Author
	if len(author) > 100 {
		author = author[:100]
	}

	id := h.service.SubmitBugReport(req.Description, author, req.Screenshot, req.GameState)
	report := h.service.GetReport(id)

	h.WriteJSONResponse(w, http.StatusAccepted, dto.BugReportResponse{
		Report: toReportDto(report),
	})
}

// GetReport handles GET /api/v1/bugs/{id}.
func (h *BugReportHandler) GetReport(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	report := h.service.GetReport(id)
	if report == nil {
		h.WriteErrorResponse(w, http.StatusNotFound, "Bug report not found")
		return
	}

	h.WriteJSONResponse(w, http.StatusOK, dto.BugReportResponse{
		Report: toReportDto(report),
	})
}

func toReportDto(r *bugreport.Report) dto.BugReportDto {
	return dto.BugReportDto{
		ID:            r.ID,
		Status:        r.Status,
		StatusMessage: r.StatusMessage,
		IssueURL:      r.IssueURL,
	}
}
