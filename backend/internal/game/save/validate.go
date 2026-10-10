package save

import (
	"errors"
	"fmt"
	"reflect"
	"slices"
	"strings"
	"unicode/utf8"

	"openmars/internal/game"
	"openmars/internal/game/award"
	"openmars/internal/game/board"
	"openmars/internal/game/cards"
	"openmars/internal/game/colony"
	"openmars/internal/game/datastore"
	"openmars/internal/game/milestone"
	"openmars/internal/game/projectfunding"
	"openmars/internal/game/shared"
	"openmars/internal/game/standardproject"
)

// Catalog supplies immutable definitions used to validate a portable position.
type Catalog struct {
	StandardProjects standardproject.StandardProjectRegistry
	Cards            cards.CardRegistry
	Maps             *board.MapRegistry
	Colonies         colony.ColonyRegistry
	Awards           award.AwardRegistry
	Milestones       milestone.MilestoneRegistry
	Projects         projectfunding.ProjectFundingRegistry
}

// Fingerprint includes all loaded rule assets with deterministic project ordering.
func (c Catalog) Fingerprint() (string, error) {
	maps := []*board.MapDefinition{}
	for _, info := range c.Maps.ListMaps() {
		m, _ := c.Maps.GetMap(info.ID)
		maps = append(maps, m)
	}
	projects := c.Projects.GetAll()
	slices.SortFunc(projects, func(a, b projectfunding.ProjectDefinition) int { return strings.Compare(a.ID, b.ID) })
	return Fingerprint(struct {
		Cards            []cards.Card
		Maps             []*board.MapDefinition
		Colonies         []colony.ColonyDefinition
		Awards           []award.AwardDefinition
		Milestones       []milestone.MilestoneDefinition
		Projects         []projectfunding.ProjectDefinition
		StandardProjects []standardproject.StandardProjectDefinition
	}{c.Cards.GetAll(), maps, c.Colonies.GetAll(), c.Awards.GetAll(), c.Milestones.GetAll(), projects, c.StandardProjects.GetAll()})
}

// Validate checks current gameplay invariants and the complete historical archive.
func Validate(doc *Document, catalog Catalog, fingerprint string) (err error) {
	defer func() {
		if err != nil {
			var classified *Error
			if !errors.As(err, &classified) {
				err = Failure("invalid_save", "This save contains invalid game data. Select another export.", err)
			}
		}
	}()
	if doc == nil || doc.State == nil {
		return fmt.Errorf("state: required")
	}
	if doc.FormatVersion != FormatVersion || doc.RulesVersion != RulesVersion {
		return Failure("incompatible_save", "This save uses a different game version. Open it on a matching server.", nil)
	}
	if doc.ContentFingerprint != fingerprint {
		return Failure("incompatible_save", "This save uses different game content. Open it on a matching server.", nil)
	}
	if doc.SavedAt.IsZero() {
		return fmt.Errorf("savedAt: required")
	}
	if doc.State.Status != shared.GameStatusActive {
		return fmt.Errorf("only in-progress games can be loaded")
	}
	if err := catalog.validateState(doc.State); err != nil {
		return fmt.Errorf("state: %w", err)
	}
	if len(doc.History) > MaxEntries || len(doc.Log) > MaxEntries {
		return fmt.Errorf("too many history entries")
	}
	// Archive snapshots include intermediate mutations (for example, assigning a
	// host before inserting that player). Resolve their references across the archive.
	archive := *doc.State
	archive.Players = map[string]*datastore.PlayerState{}
	archive.Settings = doc.State.Settings
	archive.Settings.CardPacks = []string{shared.PackBaseGame, shared.PackPrelude, shared.PackVenus, shared.PackExperimental, shared.PackColonies, shared.PackProjectFunding}
	for id, p := range doc.State.Players {
		archive.Players[id] = p
	}
	for _, h := range doc.History {
		if h != nil && h.State != nil {
			for id, p := range h.State.Players {
				archive.Players[id] = p
			}
		}
	}
	var seq int64
	for i, h := range doc.History {
		if h == nil || h.State == nil || h.GameID != doc.State.ID || h.State.ID != doc.State.ID || h.Sequence <= seq || h.Timestamp.IsZero() {
			return fmt.Errorf("history[%d]: invalid entry or sequence", i)
		}
		for id, p := range h.State.Players {
			if p == nil || p.ID != id {
				return fmt.Errorf("history[%d]: invalid player", i)
			}
		}
		if err := catalog.walk(reflect.ValueOf(h.State), &archive, "history"); err != nil {
			return fmt.Errorf("history[%d]: %w", i, err)
		}
		seq = h.Sequence
	}
	seq = 0
	for i, entry := range doc.Log {
		if entry.GameID != doc.State.ID || entry.SequenceNumber <= seq || entry.Timestamp.IsZero() {
			return fmt.Errorf("log[%d]: invalid entry or sequence", i)
		}
		if entry.PlayerID != "" && archive.Players[entry.PlayerID] == nil {
			return fmt.Errorf("log[%d]: unknown player", i)
		}
		if err := catalog.walk(reflect.ValueOf(entry), &archive, "log"); err != nil {
			return err
		}
		seq = entry.SequenceNumber
	}
	if len(doc.Log) > 0 && doc.LogBaseline == nil {
		return fmt.Errorf("logBaseline: required when log is present")
	}
	if doc.LogBaseline != nil {
		for id, p := range doc.LogBaseline.Players {
			if p == nil || archive.Players[id] == nil {
				return fmt.Errorf("logBaseline: invalid player")
			}
		}
		if err := catalog.walk(reflect.ValueOf(doc.LogBaseline), &archive, "logBaseline"); err != nil {
			return err
		}
	}
	return nil
}

func (c Catalog) validateState(s *datastore.GameState) error {
	if s.ID == "" || len(s.ID) > 128 {
		return fmt.Errorf("id: required and at most 128 characters")
	}
	if s.CreatedAt.IsZero() || s.UpdatedAt.IsZero() {
		return fmt.Errorf("timestamps: required")
	}
	if s.Generation < 1 || s.Generation > 10000 {
		return fmt.Errorf("generation: out of range")
	}
	if s.Settings.MaxPlayers < 1 || s.Settings.MaxPlayers > 10 || len(s.Players) > s.Settings.MaxPlayers {
		return fmt.Errorf("players: invalid player count")
	}
	if s.Players == nil {
		return fmt.Errorf("players: required")
	}
	if len(s.Players) == 0 || s.Players[s.HostPlayerID] == nil {
		return fmt.Errorf("hostPlayerId: unknown player")
	}
	def, ok := c.Maps.GetMap(s.Settings.MapID)
	if !ok {
		return fmt.Errorf("settings.mapId: unknown map")
	}
	packs := []string{shared.PackBaseGame, shared.PackPrelude, shared.PackVenus, shared.PackExperimental, shared.PackColonies, shared.PackProjectFunding}
	if !slices.Contains(s.Settings.CardPacks, shared.PackBaseGame) {
		return fmt.Errorf("settings.cardPacks: base game required")
	}
	for _, p := range s.Settings.CardPacks {
		if !slices.Contains(packs, p) {
			return fmt.Errorf("settings.cardPacks: unknown pack %q", p)
		}
	}
	if s.Settings.VenusNextEnabled && !slices.Contains(s.Settings.CardPacks, shared.PackVenus) {
		return fmt.Errorf("settings: Venus requires its card pack")
	}
	phases := []shared.GamePhase{shared.GamePhaseWaitingForGameStart, shared.GamePhaseStartingSelection, shared.GamePhaseInitApplyCorp, shared.GamePhaseInitApplyPrelude, shared.GamePhaseAction, shared.GamePhaseProductionAndCardDraw, shared.GamePhaseFinalPhase, shared.GamePhaseComplete}
	if !slices.Contains(phases, s.CurrentPhase) {
		return fmt.Errorf("currentPhase: unknown phase")
	}
	if !slices.Contains([]shared.GameStatus{shared.GameStatusLobby, shared.GameStatusActive, shared.GameStatusCompleted}, s.Status) {
		return fmt.Errorf("status: unknown status")
	}
	if s.CurrentPhase == shared.GamePhaseWaitingForGameStart || s.CurrentPhase == shared.GamePhaseComplete {
		return fmt.Errorf("currentPhase: not an active phase")
	}
	if s.Temperature < -30 || s.Temperature > 8 || s.Temperature%2 != 0 || s.Oxygen < 0 || s.Oxygen > 14 || s.Oceans < 0 || s.Oceans > s.MaxOceans || s.MaxOceans > 9 || s.Venus < 0 || s.Venus > 30 || s.Venus%2 != 0 {
		return fmt.Errorf("global parameters: out of range")
	}
	if s.CurrentTurnActions < -1 || s.CurrentTurnTotalActions < -1 || s.GlobalActionCounter < 0 || s.ShuffleCount < 0 || s.DrawnCardCount < 0 {
		return fmt.Errorf("turn/deck counters: out of range")
	}
	if err := checkOrder(s.PlayerOrder, s.Players); err != nil {
		return fmt.Errorf("playerOrder: %w", err)
	}
	if err := checkOrder(s.TurnOrder, s.Players); err != nil {
		return fmt.Errorf("turnOrder: %w", err)
	}
	if s.Players[s.CurrentTurnPlayerID] == nil {
		return fmt.Errorf("currentTurnPlayerId: unknown player")
	}
	expected := board.GenerateBoardFromMap(def, s.Settings.VenusNextEnabled)
	hexes := map[string]board.Tile{}
	for _, tile := range expected {
		hexes[tile.Coordinates.String()] = tile
	}
	seen := map[string]bool{}
	if len(s.Tiles) != len(expected) {
		return fmt.Errorf("tiles: board does not match map")
	}
	for _, tile := range s.Tiles {
		hex := tile.Coordinates.String()
		expectedTile, known := hexes[hex]
		if !known || seen[hex] {
			return fmt.Errorf("tiles: unknown or duplicate hex %q", hex)
		}
		seen[hex] = true
		if tile.Type != expectedTile.Type || tile.Location != expectedTile.Location {
			return fmt.Errorf("tiles: terrain does not match map at %s", hex)
		}
		if tile.OccupiedBy != nil && !board.ValidPlaceableTileType(strings.TrimSuffix(string(tile.OccupiedBy.Type), "-tile")) {
			return fmt.Errorf("tiles: unknown tile type %q", tile.OccupiedBy.Type)
		}
	}
	locations := map[string]string{}
	register := func(ids []string, where string) error {
		for _, id := range ids {
			if id == "" {
				return fmt.Errorf("%s: empty card ID", where)
			}
			if prev, ok := locations[id]; ok {
				return fmt.Errorf("card %s occurs in both %s and %s", id, prev, where)
			}
			locations[id] = where
		}
		return nil
	}
	for _, zone := range []struct {
		ids  []string
		name string
	}{{s.ProjectCards, "deck"}, {s.Corporations, "corporations"}, {s.PreludeCards, "preludes"}, {s.DiscardPile, "discard"}, {s.RemovedCards, "removed"}} {
		if err := register(zone.ids, zone.name); err != nil {
			return err
		}
	}
	for _, phase := range s.SelectCorporationPhases {
		if phase != nil {
			if err := register(phase.AvailableCorporations, "corporation selection"); err != nil {
				return err
			}
		}
	}
	for _, phase := range s.SelectStartingCardsPhases {
		if phase != nil && !phase.SelectionComplete {
			if err := register(phase.AvailableCards, "starting cards"); err != nil {
				return err
			}
		}
	}
	for _, phase := range s.SelectPreludeCardsPhases {
		if phase != nil {
			if err := register(phase.AvailablePreludes, "prelude selection"); err != nil {
				return err
			}
		}
	}
	for _, phase := range s.ProductionPhases {
		if phase != nil && !phase.SelectionComplete {
			if err := register(phase.AvailableCards, "production cards"); err != nil {
				return err
			}
		}
	}
	for _, choices := range s.DeferredStartingChoices {
		if choices == nil {
			continue
		}
		if !choices.CorpApplied {
			if err := register(choices.CardIDs, "chosen starting cards"); err != nil {
				return err
			}
		}
		if choices.PreludesAppliedCount >= 0 && choices.PreludesAppliedCount <= len(choices.PreludeIDs) {
			if err := register(choices.PreludeIDs[choices.PreludesAppliedCount:], "chosen preludes"); err != nil {
				return err
			}
		}
	}

	for id, p := range s.Players {
		if p == nil || p.ID != id || id == "" {
			return fmt.Errorf("players: invalid identity")
		}
		if strings.TrimSpace(p.Name) == "" || utf8.RuneCountInString(p.Name) > game.MaxPlayerNameLength {
			return fmt.Errorf("players.%s.name: invalid name", id)
		}
		if !slices.Contains(shared.PlayerColors, p.Color) {
			return fmt.Errorf("players.%s.color: unknown color", id)
		}
		if p.TerraformRating < 0 {
			return fmt.Errorf("players.%s.terraformRating: negative rating", id)
		}
		if p.PlayerType != "human" && p.PlayerType != "bot" {
			return fmt.Errorf("players.%s.playerType: invalid", id)
		}
		if pending := p.PendingCardDrawSelection; pending != nil {
			if err := register(pending.AvailableCards, "pending draw "+id); err != nil {
				return err
			}
		}
		if err := register(p.HandCardIDs, "hand "+id); err != nil {
			return err
		}
		if err := register(p.PlayedCardIDs, "played "+id); err != nil {
			return err
		}
		if p.CorporationID != "" {
			card, err := c.Cards.GetByID(p.CorporationID)
			if err != nil || card.Type != cards.CardTypeCorporation {
				return fmt.Errorf("players.%s.corporationId: invalid corporation", id)
			}
			if err := register([]string{p.CorporationID}, "corporation "+id); err != nil {
				return err
			}
		}
		if p.Resources.Credits < 0 || p.Resources.Steel < 0 || p.Resources.Titanium < 0 || p.Resources.Plants < 0 || p.Resources.Energy < 0 || p.Resources.Heat < 0 {
			return fmt.Errorf("players.%s.resources: negative resource", id)
		}
		if p.Production.Credits < -5 || p.Production.Steel < 0 || p.Production.Titanium < 0 || p.Production.Plants < 0 || p.Production.Energy < 0 || p.Production.Heat < 0 {
			return fmt.Errorf("players.%s.production: out of range", id)
		}
		for _, a := range p.Actions {
			if a.TimesUsedThisTurn < 0 || a.TimesUsedThisGeneration < 0 {
				return fmt.Errorf("players.%s.actions: negative usage", id)
			}
		}
	}
	for id, f := range s.TradeFleets {
		if s.Players[id] == nil || f.Used < 0 || f.Capacity < 0 || f.Used > f.Capacity {
			return fmt.Errorf("tradeFleets: invalid fleet")
		}
	}
	for _, state := range s.ColonyStates {
		if state == nil {
			return fmt.Errorf("colonyStates: null colony")
		}
		def, err := c.Colonies.GetByID(state.DefinitionID)
		if err != nil || state.MarkerPosition < -1 || state.MarkerPosition >= len(def.Steps) || len(state.PlayerColonies) > len(def.Colonies) {
			return fmt.Errorf("colonyStates: invalid colony")
		}
	}
	for _, id := range s.SelectedAwards {
		if _, err := c.Awards.GetByID(id); err != nil {
			return err
		}
	}
	for _, id := range s.SelectedMilestones {
		if _, err := c.Milestones.GetByID(id); err != nil {
			return err
		}
	}
	for id, first := range s.ForcedFirstActions {
		if s.Players[id] == nil || first == nil || !slices.Contains([]string{"queued", "resolving"}, first.State) {
			return fmt.Errorf("forcedFirstActions: invalid continuation")
		}
		card, err := c.Cards.GetByID(first.CorporationID)
		if err != nil {
			return err
		}
		for _, index := range first.BehaviorIndices {
			if index < 0 || index >= len(card.Behaviors) {
				return fmt.Errorf("forcedFirstActions: invalid behavior index")
			}
		}
	}
	if err := c.validatePosition(s); err != nil {
		return err
	}
	return c.walk(reflect.ValueOf(s), s, "state")
}

func checkOrder(order []string, players map[string]*datastore.PlayerState) error {
	seen := map[string]bool{}
	for _, id := range order {
		if seen[id] || players[id] == nil {
			return fmt.Errorf("unknown or duplicate player")
		}
		seen[id] = true
	}
	if len(seen) != len(players) {
		return fmt.Errorf("missing player")
	}
	return nil
}

func (c Catalog) cardID(id string, s *datastore.GameState) error {
	if id == "" {
		return nil
	}
	card, err := c.Cards.GetByID(id)
	if err != nil {
		return err
	}
	if !s.Settings.EnabledPacks()[card.Pack] {
		return fmt.Errorf("card %s belongs to a disabled pack", id)
	}
	return nil
}

func (c Catalog) walk(v reflect.Value, s *datastore.GameState, path string) error {
	if !v.IsValid() {
		return nil
	}
	if v.CanInterface() {
		if condition, ok := v.Interface().(shared.BehaviorCondition); ok && !v.IsNil() {
			if !slices.Contains(shared.AllResourceTypes, condition.GetResourceType()) {
				return fmt.Errorf("%s: unknown resource type", path)
			}
			if condition.GetAmount() < -10000 || condition.GetAmount() > 10000 {
				return fmt.Errorf("%s: effect amount out of range", path)
			}
			if problems := shared.ValidateResourceCondition(condition, strings.Contains(path, ".Inputs[")); len(problems) > 0 {
				return fmt.Errorf("%s: %s", path, strings.Join(problems, "; "))
			}
		}
	}
	for v.Kind() == reflect.Pointer || v.Kind() == reflect.Interface {
		if v.IsNil() {
			return nil
		}
		v = v.Elem()
	}
	if v.CanInterface() {
		if per, ok := v.Interface().(shared.PerCondition); ok && per.Amount <= 0 {
			return fmt.Errorf("%s: per divisor must be positive", path)
		}
		if callback, ok := v.Interface().(shared.TileCompletionCallback); ok {
			if !slices.Contains([]string{"convert-plants-to-greenery", "standard-project-greenery", "standard-project-aquifer", "adjacent-removal"}, callback.Type) {
				return fmt.Errorf("%s: unknown tile callback", path)
			}
			if callback.Type == "adjacent-removal" && callback.Output == nil {
				return fmt.Errorf("%s: missing removal output", path)
			}
		}
	}
	switch v.Kind() {
	case reflect.Struct:
		t := v.Type()
		cardID := ""
		for _, name := range []string{"CardID", "SourceCardID"} {
			if f := v.FieldByName(name); f.IsValid() && f.Kind() == reflect.String && f.String() != "" {
				cardID = f.String()
				break
			}
		}
		for i := 0; i < v.NumField(); i++ {
			field := t.Field(i)
			if !field.IsExported() || field.Tag.Get("json") == "-" {
				continue
			}
			value := v.Field(i)
			if field.Name == "BehaviorIndex" || field.Name == "SourceBehaviorIndex" {
				if cardID != "" {
					card, err := c.Cards.GetByID(cardID)
					if err != nil {
						return err
					}
					if value.Int() < 0 || int(value.Int()) >= len(card.Behaviors) {
						return fmt.Errorf("%s.%s: invalid behavior index", path, field.Name)
					}
				}
			}
			if err := c.reference(value, field.Name, s, path); err != nil {
				return err
			}
			if err := c.walk(value, s, path+"."+field.Name); err != nil {
				return err
			}
		}
	case reflect.Map:
		iter := v.MapRange()
		for iter.Next() {
			entry := iter.Value()
			if (entry.Kind() == reflect.Pointer || entry.Kind() == reflect.Interface) && entry.IsNil() {
				return fmt.Errorf("%s: null map entry", path)
			}
			if err := c.walk(entry, s, path); err != nil {
				return err
			}
		}
	case reflect.Slice, reflect.Array:
		if v.Len() > MaxEntries {
			return fmt.Errorf("%s: too many entries", path)
		}
		for i := 0; i < v.Len(); i++ {
			entry := v.Index(i)
			if (entry.Kind() == reflect.Pointer || entry.Kind() == reflect.Interface) && entry.IsNil() {
				return fmt.Errorf("%s[%d]: null entry", path, i)
			}
			if err := c.walk(entry, s, fmt.Sprintf("%s[%d]", path, i)); err != nil {
				return err
			}
		}
	case reflect.Int, reflect.Int64:
		if v.Int() < -1000000000 || v.Int() > 1000000000 {
			return fmt.Errorf("%s: numeric value out of range", path)
		}
	case reflect.String:
		if len(v.String()) > MaxStringBytes {
			return fmt.Errorf("%s: string too long", path)
		}
	}
	return nil
}

func (c Catalog) reference(v reflect.Value, name string, s *datastore.GameState, path string) error {
	if v.Kind() == reflect.Pointer {
		if v.IsNil() {
			return nil
		}
		v = v.Elem()
	}
	if v.Kind() == reflect.Slice {
		for i := 0; i < v.Len(); i++ {
			if err := c.reference(v.Index(i), strings.TrimSuffix(name, "s"), s, path); err != nil {
				return err
			}
		}
		return nil
	}
	if v.Kind() != reflect.String || v.String() == "" {
		return nil
	}
	id := v.String()
	switch name {
	case "CardID", "CorporationID", "SourceCardID", "TriggeringCardID", "HandCardID", "PlayedCardID", "ProjectCard", "Corporation", "PreludeCard", "PreludeID", "Card", "DiscardPile", "RemovedCard", "AvailableCard", "AvailableCorporation", "AvailablePrelude", "Cards":
		if err := c.cardID(id, s); err != nil {
			return fmt.Errorf("%s.%s: %w", path, name, err)
		}
	case "PlayerID", "HostPlayerID", "CurrentTurnPlayerID", "CurrentTurn", "TargetPlayerID", "TriggeringPlayerID", "FundedByPlayer", "OwnerID", "ReservedBy", "TraderID", "EligiblePlayerID", "PlayerColonie", "SeatOwner":
		if s.Players[id] == nil {
			return fmt.Errorf("%s.%s: unknown player %q", path, name, id)
		}
	}
	return nil
}
