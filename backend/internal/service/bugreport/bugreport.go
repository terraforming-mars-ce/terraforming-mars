package bugreport

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/bradleyfalzon/ghinstallation/v2"
	"github.com/google/go-github/v75/github"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Config holds configuration for the bug report service.
type Config struct {
	GitHubAppID          int64
	GitHubInstallationID int64
	GitHubPrivateKeyPath string
	GitHubRepoOwner      string
	GitHubRepoName       string
	RepoPath             string
}

// Report represents an in-progress or completed bug report.
type Report struct {
	ID            string `json:"id"`
	Status        string `json:"status"`
	StatusMessage string `json:"statusMessage"`
	IssueURL      string `json:"issueUrl,omitempty"`
}

// Capabilities tracks which external services are available.
type Capabilities struct {
	GitHubApp bool
	Claude    bool
}

// Service handles bug report submission via GitHub Issues.
type Service struct {
	logger       *zap.Logger
	config       Config
	ghClient     *github.Client
	capabilities Capabilities

	mu      sync.Mutex
	reports map[string]*Report
}

// NewService creates and initializes a bug report service.
func NewService(logger *zap.Logger) *Service {
	s := &Service{
		logger:  logger,
		reports: make(map[string]*Report),
	}
	s.initialize()
	return s
}

func (s *Service) initialize() {
	cfg := loadConfig()
	s.config = cfg

	s.capabilities.GitHubApp = s.initGitHub(cfg)

	// Check Claude availability in the background to avoid blocking startup
	go func() {
		result := s.initClaude(cfg)
		s.mu.Lock()
		s.capabilities.Claude = result
		s.mu.Unlock()
		s.logger.Debug("Bug report service initialized",
			zap.Bool("github_app", s.capabilities.GitHubApp),
			zap.Bool("claude", result))
	}()
}

func (s *Service) initGitHub(cfg Config) bool {
	if cfg.GitHubInstallationID == 0 {
		s.logger.Debug("Bug report: GitHub App disabled (GITHUB_INSTALLATION_ID not set)")
		return false
	}

	if _, err := os.Stat(cfg.GitHubPrivateKeyPath); os.IsNotExist(err) {
		s.logger.Debug("Bug report: GitHub App disabled (private key not found: " + cfg.GitHubPrivateKeyPath + ")")
		return false
	}

	itr, err := ghinstallation.NewKeyFromFile(
		http.DefaultTransport,
		cfg.GitHubAppID,
		cfg.GitHubInstallationID,
		cfg.GitHubPrivateKeyPath,
	)
	if err != nil {
		s.logger.Error("Bug report: GitHub App disabled (failed to create transport)", zap.Error(err))
		return false
	}

	s.ghClient = github.NewClient(&http.Client{Transport: itr})
	return true
}

func (s *Service) initClaude(cfg Config) bool {
	if _, err := exec.LookPath("claude"); err != nil {
		s.logger.Warn("Bug report: Claude disabled (CLI not found in PATH)")
		return false
	}

	if cfg.RepoPath == "" {
		s.logger.Warn("Bug report: Claude disabled (TM_REPO_PATH not set)")
		return false
	}
	if _, err := os.Stat(cfg.RepoPath); os.IsNotExist(err) {
		s.logger.Warn("Bug report: Claude disabled (repo path not found: " + cfg.RepoPath + ")")
		return false
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "claude", "-p", "--dangerously-skip-permissions", "respond with ok")
	out, err := cmd.Output()
	if err != nil {
		s.logger.Warn("Bug report: Claude disabled (auth check failed)", zap.Error(err))
		return false
	}

	if !strings.Contains(strings.ToLower(strings.TrimSpace(string(out))), "ok") {
		s.logger.Warn("Bug report: Claude disabled (unexpected response from auth check)")
		return false
	}

	return true
}

// IsAvailable returns whether the bug report service can create issues.
func (s *Service) IsAvailable() bool {
	return s.capabilities.GitHubApp
}

// Capabilities returns which external services are available.
func (s *Service) Capabilities() Capabilities {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.capabilities
}

// GetReport returns a report by ID, or nil if not found.
func (s *Service) GetReport(id string) *Report {
	s.mu.Lock()
	defer s.mu.Unlock()
	r := s.reports[id]
	if r == nil {
		return nil
	}
	copy := *r
	return &copy
}

func (s *Service) updateReport(id, status, message string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if r := s.reports[id]; r != nil {
		r.Status = status
		r.StatusMessage = message
	}
}

func (s *Service) completeReport(id, issueURL string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if r := s.reports[id]; r != nil {
		r.Status = "completed"
		r.StatusMessage = "Issue created"
		r.IssueURL = issueURL
	}
}

func (s *Service) failReport(id, message string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if r := s.reports[id]; r != nil {
		r.Status = "failed"
		r.StatusMessage = message
	}
}

// SubmitBugReport starts async processing of a bug report and returns the report ID.
func (s *Service) SubmitBugReport(description string, author string, screenshot string, gameState json.RawMessage) string {
	id := uuid.New().String()

	s.mu.Lock()
	s.reports[id] = &Report{
		ID:            id,
		Status:        "processing",
		StatusMessage: "Preparing bug report...",
	}
	s.mu.Unlock()

	go s.processBugReport(id, description, author, screenshot, gameState)

	return id
}

func (s *Service) processBugReport(id string, description string, author string, screenshot string, gameState json.RawMessage) {
	tmpDir, err := os.MkdirTemp("", "bugreport-*")
	if err != nil {
		s.failReport(id, "Failed to create temp directory")
		s.logger.Error("Failed to create temp dir", zap.Error(err))
		return
	}
	defer func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			s.logger.Warn("Failed to remove temp directory", zap.String("path", tmpDir), zap.Error(err))
		}
	}()

	var gameStatePath string
	if len(gameState) > 0 {
		gameStatePath = filepath.Join(tmpDir, "gamestate.json")
		prettyState, err := json.MarshalIndent(json.RawMessage(gameState), "", "  ")
		if err != nil {
			prettyState = gameState
		}
		if err := os.WriteFile(gameStatePath, prettyState, 0600); err != nil {
			s.failReport(id, "Failed to write game state")
			s.logger.Error("Failed to write game state", zap.Error(err))
			return
		}
	}

	var screenshotPath string
	if screenshot != "" {
		screenshotPath = filepath.Join(tmpDir, "screenshot.png")
		raw := screenshot
		if idx := strings.Index(raw, ","); idx != -1 {
			raw = raw[idx+1:]
		}
		imgData, err := base64.StdEncoding.DecodeString(raw)
		if err != nil {
			s.logger.Warn("Failed to decode screenshot, skipping", zap.Error(err))
		} else {
			if err := os.WriteFile(screenshotPath, imgData, 0600); err != nil {
				s.logger.Warn("Failed to write screenshot, skipping", zap.Error(err))
				screenshotPath = ""
			}
		}
	}

	var analysis, title string
	analysisFailed := true

	if s.capabilities.Claude {
		s.updateReport(id, "processing", "Analyzing bug report...")

		claudeCtx, cancel := context.WithTimeout(context.Background(), 660*time.Second)
		analysis, title, analysisFailed = s.generateAnalysis(claudeCtx, description, gameStatePath, screenshotPath)
		cancel()
	}

	if analysisFailed {
		title = description
		if len(title) > 60 {
			title = title[:57] + "..."
		}
	}

	s.updateReport(id, "processing", "Creating GitHub issue...")

	body := buildIssueBody(description, author, analysis, analysisFailed, gameState)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	issueTitle := "Bug Report: " + title
	issue, _, err := s.ghClient.Issues.Create(ctx, s.config.GitHubRepoOwner, s.config.GitHubRepoName, &github.IssueRequest{
		Title:  github.Ptr(issueTitle),
		Body:   github.Ptr(body),
		Labels: &[]string{"bug", "in-game"},
	})
	if err != nil {
		s.failReport(id, "Failed to create GitHub issue")
		s.logger.Error("Failed to create GitHub issue", zap.Error(err))
		return
	}

	msg := "Issue created"
	if analysisFailed {
		msg = "Issue created (analysis failed, used description only)"
	}
	s.mu.Lock()
	if r := s.reports[id]; r != nil {
		r.Status = "completed"
		r.StatusMessage = msg
		r.IssueURL = issue.GetHTMLURL()
	}
	s.mu.Unlock()
	s.logger.Debug("Bug report submitted", zap.String("issue_url", issue.GetHTMLURL()))
}

func loadConfig() Config {
	cfg := Config{
		GitHubPrivateKeyPath: "./private-key.pem",
		GitHubRepoOwner:      "rackaracka123",
		GitHubRepoName:       "terraforming-mars",
	}

	if v := os.Getenv("GITHUB_APP_ID"); v != "" {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil {
			cfg.GitHubAppID = n
		}
	}
	if v := os.Getenv("GITHUB_INSTALLATION_ID"); v != "" {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil {
			cfg.GitHubInstallationID = n
		}
	}
	if v := os.Getenv("GITHUB_PRIVATE_KEY_PATH"); v != "" {
		cfg.GitHubPrivateKeyPath = v
	}
	if v := os.Getenv("GITHUB_REPO_OWNER"); v != "" {
		cfg.GitHubRepoOwner = v
	}
	if v := os.Getenv("GITHUB_REPO_NAME"); v != "" {
		cfg.GitHubRepoName = v
	}
	cfg.RepoPath = os.Getenv("TM_REPO_PATH")

	return cfg
}

func (s *Service) generateAnalysis(ctx context.Context, description string, gameStatePath string, screenshotPath string) (analysis string, title string, failed bool) {
	var promptParts []string
	promptParts = append(promptParts,
		"You are analyzing a bug report for the Terraforming Mars digital board game.",
		"The source code is at: "+s.config.RepoPath,
		"  - Backend (Go): "+filepath.Join(s.config.RepoPath, "backend/internal/")+" and "+filepath.Join(s.config.RepoPath, "backend/cmd/"),
		"  - Frontend (React/TypeScript): "+filepath.Join(s.config.RepoPath, "frontend/src/"),
		"  - Card database (JSON): "+filepath.Join(s.config.RepoPath, "backend/assets/"),
		"",
		"Player's description of the bug:",
		description,
		"",
	)

	if gameStatePath != "" {
		promptParts = append(promptParts, "The game state JSON is at: "+gameStatePath)
	}

	if screenshotPath != "" {
		promptParts = append(promptParts, "A screenshot of the game at the time of the report is at: "+screenshotPath)
	}

	promptParts = append(promptParts,
		"",
		"Your job is to provide context for whoever picks up this issue. You are NOT solving the bug. You are triaging it.",
		"",
		"Look through the source code and game state to identify likely causes. Read files, check the game state JSON, look at whatever is available to you. You have read-only access to the source and any files on this system.",
		"",
		"Be honest about your confidence level. If you can pinpoint the cause, say so. If you can only narrow it down to a few possibilities, list them. If you genuinely cannot figure out what went wrong, say that -- the player experienced something real, so the issue is worth investigating regardless.",
		"",
		"Do not force a conclusion. A good triage note that says 'I could not reproduce this from the code, but the player reported X and the game state shows Y, so it is worth looking at Z' is more useful than a wrong diagnosis.",
		"",
		"Respond with exactly this format:",
		"Line 1: A short bug title (max 60 characters)",
		"Line 2: empty",
		"Lines 3+: Your triage notes. Reference specific source files and functions when you can. Include relevant excerpts from the game state if they help. List likely causes ranked by probability. Flag anything unusual you noticed in the game state even if you are not sure it is related.",
		"",
		"Rules: No emojis. No filler phrases. No sycophancy. Write like a developer leaving notes for another developer.",
	)

	prompt := strings.Join(promptParts, "\n")

	cmdCtx, cancel := context.WithTimeout(ctx, 600*time.Second)
	defer cancel()

	cmd := exec.CommandContext(cmdCtx, "claude", "-p", "--dangerously-skip-permissions", prompt)
	var stderr strings.Builder
	cmd.Stderr = &stderr
	output, err := cmd.Output()
	if err != nil {
		s.logger.Warn("Claude CLI analysis failed",
			zap.Error(err),
			zap.String("stderr", stderr.String()))
		return "", "", true
	}

	lines := strings.SplitN(strings.TrimSpace(string(output)), "\n", 3)
	title = strings.TrimSpace(lines[0])
	if len(title) > 60 {
		title = title[:57] + "..."
	}

	if len(lines) > 2 {
		analysis = strings.TrimSpace(lines[2])
	} else {
		analysis = description
	}

	return analysis, title, false
}

func buildIssueBody(description, author, analysis string, analysisFailed bool, gameState json.RawMessage) string {
	var b strings.Builder

	b.WriteString("## Bug Report\n\n")

	if author != "" {
		b.WriteString("Generated by **" + author + "**\n\n")
	}

	if !analysisFailed {
		b.WriteString("### Analysis\n\n")
		b.WriteString(analysis)
		b.WriteString("\n\n")
	}

	b.WriteString("### Player Description\n\n")
	b.WriteString(description)
	b.WriteString("\n\n")

	b.WriteString("### Game Meta\n\n")
	b.WriteString(buildGameMeta(gameState))
	b.WriteString("\n")

	return b.String()
}

func buildGameMeta(gameState json.RawMessage) string {
	if len(gameState) == 0 {
		return "_No game state available_"
	}

	var game struct {
		ID               string  `json:"id"`
		Status           string  `json:"status"`
		CurrentPhase     string  `json:"currentPhase"`
		Generation       int     `json:"generation"`
		CurrentTurn      *string `json:"currentTurn"`
		GlobalParameters struct {
			Temperature int `json:"temperature"`
			Oxygen      int `json:"oxygen"`
			Oceans      int `json:"oceans"`
			MaxOceans   int `json:"maxOceans"`
			Venus       int `json:"venus"`
		} `json:"globalParameters"`
		Settings struct {
			MaxPlayers int      `json:"maxPlayers"`
			CardPacks  []string `json:"cardPacks"`
		} `json:"settings"`
		CurrentPlayer *struct {
			ID              string `json:"id"`
			Name            string `json:"name"`
			TerraformRating int    `json:"terraformRating"`
			Corporation     *struct {
				Name string `json:"name"`
			} `json:"corporation"`
			Resources struct {
				Credits  int `json:"credits"`
				Steel    int `json:"steel"`
				Titanium int `json:"titanium"`
				Plants   int `json:"plants"`
				Energy   int `json:"energy"`
				Heat     int `json:"heat"`
			} `json:"resources"`
			Production struct {
				Credits  int `json:"credits"`
				Steel    int `json:"steel"`
				Titanium int `json:"titanium"`
				Plants   int `json:"plants"`
				Energy   int `json:"energy"`
				Heat     int `json:"heat"`
			} `json:"production"`
			PlayedCards []struct {
				Name string `json:"name"`
			} `json:"playedCards"`
			Cards []struct{} `json:"cards"`
		} `json:"currentPlayer"`
		OtherPlayers []struct {
			ID              string `json:"id"`
			Name            string `json:"name"`
			TerraformRating int    `json:"terraformRating"`
			Corporation     *struct {
				Name string `json:"name"`
			} `json:"corporation"`
			Resources struct {
				Credits  int `json:"credits"`
				Steel    int `json:"steel"`
				Titanium int `json:"titanium"`
				Plants   int `json:"plants"`
				Energy   int `json:"energy"`
				Heat     int `json:"heat"`
			} `json:"resources"`
			PlayedCards []struct {
				Name string `json:"name"`
			} `json:"playedCards"`
			HandCardCount int `json:"handCardCount"`
		} `json:"otherPlayers"`
		Milestones []struct {
			Name      string  `json:"name"`
			ClaimedBy *string `json:"claimedBy"`
		} `json:"milestones"`
		Awards []struct {
			Name     string  `json:"name"`
			FundedBy *string `json:"fundedBy"`
		} `json:"awards"`
	}

	if err := json.Unmarshal(gameState, &game); err != nil {
		return "_Failed to parse game state_"
	}

	var b strings.Builder

	b.WriteString("| Property | Value |\n")
	b.WriteString("|----------|-------|\n")
	b.WriteString("| Game ID | `" + game.ID + "` |\n")
	b.WriteString("| Status | " + game.Status + " |\n")
	b.WriteString("| Phase | " + game.CurrentPhase + " |\n")
	b.WriteString("| Generation | " + strconv.Itoa(game.Generation) + " |\n")
	b.WriteString("| Temperature | " + strconv.Itoa(game.GlobalParameters.Temperature) + "°C |\n")
	b.WriteString("| Oxygen | " + strconv.Itoa(game.GlobalParameters.Oxygen) + "% |\n")
	b.WriteString("| Oceans | " + strconv.Itoa(game.GlobalParameters.Oceans) + "/" + strconv.Itoa(game.GlobalParameters.MaxOceans) + " |\n")
	b.WriteString("| Venus | " + strconv.Itoa(game.GlobalParameters.Venus) + "% |\n")

	if game.Settings.CardPacks != nil {
		b.WriteString("| Card Packs | " + strings.Join(game.Settings.CardPacks, ", ") + " |\n")
	}

	// Milestones claimed
	var claimedMilestones []string
	for _, m := range game.Milestones {
		if m.ClaimedBy != nil {
			claimedMilestones = append(claimedMilestones, m.Name+" ("+*m.ClaimedBy+")")
		}
	}
	if len(claimedMilestones) > 0 {
		b.WriteString("| Milestones | " + strings.Join(claimedMilestones, ", ") + " |\n")
	}

	// Awards funded
	var fundedAwards []string
	for _, a := range game.Awards {
		if a.FundedBy != nil {
			fundedAwards = append(fundedAwards, a.Name+" ("+*a.FundedBy+")")
		}
	}
	if len(fundedAwards) > 0 {
		b.WriteString("| Awards | " + strings.Join(fundedAwards, ", ") + " |\n")
	}

	// Players section
	b.WriteString("\n#### Players\n\n")

	if game.CurrentPlayer != nil {
		p := game.CurrentPlayer
		corp := "None"
		if p.Corporation != nil {
			corp = p.Corporation.Name
		}
		b.WriteString("**" + p.Name + "** (reporting player)\n")
		b.WriteString("- Corporation: " + corp + "\n")
		b.WriteString("- TR: " + strconv.Itoa(p.TerraformRating) + "\n")
		b.WriteString("- Credits: " + strconv.Itoa(p.Resources.Credits) +
			", Steel: " + strconv.Itoa(p.Resources.Steel) +
			", Titanium: " + strconv.Itoa(p.Resources.Titanium) +
			", Plants: " + strconv.Itoa(p.Resources.Plants) +
			", Energy: " + strconv.Itoa(p.Resources.Energy) +
			", Heat: " + strconv.Itoa(p.Resources.Heat) + "\n")
		b.WriteString("- Production: " +
			strconv.Itoa(p.Production.Credits) + "M€, " +
			strconv.Itoa(p.Production.Steel) + " steel, " +
			strconv.Itoa(p.Production.Titanium) + " titanium, " +
			strconv.Itoa(p.Production.Plants) + " plants, " +
			strconv.Itoa(p.Production.Energy) + " energy, " +
			strconv.Itoa(p.Production.Heat) + " heat\n")
		b.WriteString("- Cards in hand: " + strconv.Itoa(len(p.Cards)) + "\n")
		b.WriteString("- Cards played: " + strconv.Itoa(len(p.PlayedCards)) + "\n")
		if len(p.PlayedCards) > 0 {
			var cardNames []string
			for _, c := range p.PlayedCards {
				cardNames = append(cardNames, c.Name)
			}
			b.WriteString("- Played: " + strings.Join(cardNames, ", ") + "\n")
		}
		b.WriteString("\n")
	}

	for _, p := range game.OtherPlayers {
		corp := "None"
		if p.Corporation != nil {
			corp = p.Corporation.Name
		}
		b.WriteString("**" + p.Name + "**\n")
		b.WriteString("- Corporation: " + corp + "\n")
		b.WriteString("- TR: " + strconv.Itoa(p.TerraformRating) + "\n")
		b.WriteString("- Credits: " + strconv.Itoa(p.Resources.Credits) +
			", Steel: " + strconv.Itoa(p.Resources.Steel) +
			", Titanium: " + strconv.Itoa(p.Resources.Titanium) +
			", Plants: " + strconv.Itoa(p.Resources.Plants) +
			", Energy: " + strconv.Itoa(p.Resources.Energy) +
			", Heat: " + strconv.Itoa(p.Resources.Heat) + "\n")
		b.WriteString("- Cards in hand: " + strconv.Itoa(p.HandCardCount) + "\n")
		b.WriteString("- Cards played: " + strconv.Itoa(len(p.PlayedCards)) + "\n")
		if len(p.PlayedCards) > 0 {
			var cardNames []string
			for _, c := range p.PlayedCards {
				cardNames = append(cardNames, c.Name)
			}
			b.WriteString("- Played: " + strings.Join(cardNames, ", ") + "\n")
		}
		b.WriteString("\n")
	}

	return b.String()
}
