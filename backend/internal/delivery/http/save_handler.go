package http

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
	gameaction "openmars/internal/action/game"
	"openmars/internal/delivery/dto"
	"openmars/internal/delivery/websocket/core"
	"openmars/internal/game/save"
)

// SaveHandler transports portable saves; all live game access goes through the hub.
type SaveHandler struct {
	action *gameaction.SaveGameAction
	hub    *core.Hub
}

func NewSaveHandler(action *gameaction.SaveGameAction, hub *core.Hub) *SaveHandler {
	return &SaveHandler{action: action, hub: hub}
}

func saveError(w http.ResponseWriter, err error) {
	public := save.PublicError(err, "The game operation couldn’t be completed. Try again. If it keeps failing, report a bug.")
	status := http.StatusBadRequest
	switch public.Code {
	case "save_too_large":
		status = http.StatusRequestEntityTooLarge
	case "forbidden":
		status = http.StatusForbidden
	case "game_not_found":
		status = http.StatusNotFound
	case "incompatible_save", "seat_unavailable", "seat_taken", "not_ready":
		status = http.StatusConflict
	case "internal_error":
		status = http.StatusInternalServerError
	}
	if status >= 500 {
		slog.Error("Game save operation failed", "code", public.Code, "error", err)
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(dto.GameSaveErrorResponse{Code: public.Code, Message: public.Message})
}

func readSaveBody(w http.ResponseWriter, r *http.Request) ([]byte, error) {
	tooLarge := func(err error) error {
		return save.Failure("save_too_large", "This save exceeds the 128 MiB limit.", err)
	}
	if r.ContentLength > save.MaxBytes {
		return nil, tooLarge(nil)
	}
	r.Body = http.MaxBytesReader(w, r.Body, save.MaxBytes)
	data, err := io.ReadAll(r.Body)
	var limit *http.MaxBytesError
	if errors.As(err, &limit) {
		return nil, tooLarge(err)
	}
	if err != nil {
		return nil, save.Failure("invalid_save", "The save couldn’t be read. Select the file again.", err)
	}
	return data, nil
}

// Validate previews a save without creating or altering a game.
func (h *SaveHandler) Validate(w http.ResponseWriter, r *http.Request) {
	data, err := readSaveBody(w, r)
	if err != nil {
		saveError(w, err)
		return
	}
	doc, err := h.action.Validate(data)
	if err != nil {
		saveError(w, err)
		return
	}
	summary := dto.GameSaveSummaryDto{Generation: doc.State.Generation, Phase: dto.GamePhase(doc.State.CurrentPhase), MapID: doc.State.Settings.MapID, SavedAt: doc.SavedAt.Format(time.RFC3339), HistoryEntries: len(doc.History), LogEntries: len(doc.Log), Seats: []dto.ResumeSeatDto{}}
	for _, id := range doc.State.PlayerOrder {
		p := doc.State.Players[id]
		summary.Seats = append(summary.Seats, dto.ResumeSeatDto{ID: id, Name: p.Name, SavedName: p.Name, CorporationID: p.CorporationID, Color: p.Color, PlayerType: p.PlayerType, Exited: p.HasExited})
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(summary)
}

// Import accepts raw JSON; seat/name are query parameters and bot credentials a header
// so JavaScript never needs to deserialize the authoritative document.
func (h *SaveHandler) Import(w http.ResponseWriter, r *http.Request) {
	data, err := readSaveBody(w, r)
	if err != nil {
		saveError(w, err)
		return
	}
	doc, err := h.action.Validate(data)
	if err != nil {
		saveError(w, err)
		return
	}
	seatID := r.URL.Query().Get("seatId")
	name := r.URL.Query().Get("playerName")
	prepared, err := h.action.PrepareImport(doc, seatID, name, r.Header.Get("X-Bot-Token"))
	if err != nil {
		saveError(w, err)
		return
	}
	var response dto.ImportGameSaveResponse
	actionErr := fmt.Errorf("save operation did not complete")
	err = h.hub.Do(r.Context(), func() {
		g, loadErr := h.action.Import(r.Context(), prepared)
		actionErr = loadErr
		if loadErr == nil {
			response = dto.ImportGameSaveResponse{GameID: g.ID(), PlayerID: seatID, PlayerName: strings.TrimSpace(name)}
		}
	})
	if err != nil {
		saveError(w, err)
		return
	}
	if actionErr != nil {
		saveError(w, actionErr)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(response)
}

// Export snapshots a host's game and encodes it outside the command executor.
func (h *SaveHandler) Export(w http.ResponseWriter, r *http.Request) {
	var doc *save.Document
	actionErr := fmt.Errorf("save operation did not complete")
	err := h.hub.Do(r.Context(), func() {
		doc, actionErr = h.action.Capture(r.Context(), mux.Vars(r)["gameId"], r.URL.Query().Get("playerId"))
	})
	if err != nil {
		saveError(w, err)
		return
	}
	if actionErr != nil {
		saveError(w, actionErr)
		return
	}
	data, err := h.action.Encode(doc)
	if err != nil {
		saveError(w, err)
		return
	}
	if _, err := h.action.Validate(data); err != nil {
		saveError(w, save.Failure("internal_error", "The game couldn’t be saved. Try again. If it keeps failing, report a bug.", err))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="openmars-gen%d-%s.json"`, doc.State.Generation, doc.SavedAt.Format("2006-01-02-150405")))
	_, _ = w.Write(data)
}
