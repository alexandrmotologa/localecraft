package api

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"net/http"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/alexandrmotologa/localecraft/internal/agent"
	"github.com/alexandrmotologa/localecraft/internal/config"
	"github.com/alexandrmotologa/localecraft/internal/icu"
	"github.com/alexandrmotologa/localecraft/internal/parser"
	"github.com/alexandrmotologa/localecraft/internal/store"
	"github.com/alexandrmotologa/localecraft/internal/typography"
)

// Server coordinates HTTP endpoints, database operations, and live UI assets.
type Server struct {
	cfg        *config.Config
	db         *store.DB
	typo       *typography.Engine
	optimizer  *agent.LayoutOptimizer
	sourceCat  *parser.Catalog
	targetCats map[string]*parser.Catalog
	webFS      fs.FS
	eventMu    sync.RWMutex
	clients    map[chan string]bool
}

// NewServer initializes the HTTP server with its supporting components.
func NewServer(cfg *config.Config, db *store.DB, typo *typography.Engine, optimizer *agent.LayoutOptimizer, sourceCat *parser.Catalog, targetCats map[string]*parser.Catalog, webFS fs.FS) *Server {
	return &Server{
		cfg:        cfg,
		db:         db,
		typo:       typo,
		optimizer:  optimizer,
		sourceCat:  sourceCat,
		targetCats: targetCats,
		webFS:      webFS,
		clients:    make(map[chan string]bool),
	}
}

// Router returns an http.Handler with all registered API and static routes.
func (s *Server) Router() http.Handler {
	mux := http.NewServeMux()

	// API endpoints
	mux.HandleFunc("GET /api/config", s.handleGetConfig)
	mux.HandleFunc("GET /api/summary", s.handleGetSummary)
	mux.HandleFunc("GET /api/keys", s.handleListKeys)
	mux.HandleFunc("POST /api/measure", s.handleMeasure)
	mux.HandleFunc("POST /api/translate", s.handleTranslate)
	mux.HandleFunc("POST /api/save-translation", s.handleSaveTranslation)
	mux.HandleFunc("POST /api/export", s.handleExport)
	mux.HandleFunc("POST /api/pr", s.handleCreatePR)
	mux.HandleFunc("GET /api/events", s.handleSSE)

	// SPA & Static assets handler
	if s.webFS != nil {
		fileServer := http.FileServer(http.FS(s.webFS))
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			if strings.HasPrefix(r.URL.Path, "/api/") {
				http.NotFound(w, r)
				return
			}

			// Check if file exists in webFS
			cleanPath := strings.TrimPrefix(r.URL.Path, "/")
			if cleanPath == "" {
				cleanPath = "index.html"
			}

			if _, err := fs.Stat(s.webFS, cleanPath); err != nil {
				// Fallback to index.html for SPA routing
				indexData, err := fs.ReadFile(s.webFS, "index.html")
				if err != nil {
					http.Error(w, "index.html not found", http.StatusNotFound)
					return
				}
				w.Header().Set("Content-Type", "text/html; charset=utf-8")
				w.Write(indexData)
				return
			}

			fileServer.ServeHTTP(w, r)
		})
	} else {
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			if strings.HasPrefix(r.URL.Path, "/api/") {
				http.NotFound(w, r)
				return
			}
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			w.Write([]byte("localecraft API server running."))
		})
	}

	return corsMiddleware(mux)
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func (s *Server) broadcastEvent(eventType string, data any) {
	s.eventMu.RLock()
	defer s.eventMu.RUnlock()

	payloadBytes, _ := json.Marshal(data)
	msg := fmt.Sprintf("event: %s\ndata: %s\n\n", eventType, string(payloadBytes))

	for ch := range s.clients {
		select {
		case ch <- msg:
		default:
		}
	}
}

func (s *Server) handleSSE(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "SSE not supported", http.StatusInternalServerError)
		return
	}

	ch := make(chan string, 16)
	s.eventMu.Lock()
	s.clients[ch] = true
	s.eventMu.Unlock()

	defer func() {
		s.eventMu.Lock()
		delete(s.clients, ch)
		s.eventMu.Unlock()
	}()

	fmt.Fprintf(w, "event: connected\ndata: {\"status\":\"ok\"}\n\n")
	flusher.Flush()

	for {
		select {
		case <-r.Context().Done():
			return
		case msg := <-ch:
			fmt.Fprint(w, msg)
			flusher.Flush()
		}
	}
}

func (s *Server) handleGetConfig(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, s.cfg)
}

func (s *Server) handleGetSummary(w http.ResponseWriter, r *http.Request) {
	sum, err := s.db.GetSummary()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, sum)
}

func (s *Server) handleListKeys(w http.ResponseWriter, r *http.Request) {
	search := r.URL.Query().Get("search")
	status := r.URL.Query().Get("status")

	keys, err := s.db.ListKeys(search, status)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, keys)
}

type measureReqBody struct {
	Text            string  `json:"text"`
	Family          string  `json:"family"`
	SizePX          float64 `json:"size_px"`
	Weight          int     `json:"weight"`
	LetterSpacingPX float64 `json:"letter_spacing_px"`
	PaddingXPX      float64 `json:"padding_x_px"`
	MaxWidthPX      float64 `json:"max_width_px"`
	MaxLines        int     `json:"max_lines"`
}

func (s *Server) handleMeasure(w http.ResponseWriter, r *http.Request) {
	var body measureReqBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	res := s.typo.Measure(typography.MeasureRequest{
		Text:            body.Text,
		Family:          body.Family,
		SizePX:          body.SizePX,
		Weight:          body.Weight,
		LetterSpacingPX: body.LetterSpacingPX,
		PaddingXPX:      body.PaddingXPX,
		MaxWidthPX:      body.MaxWidthPX,
		MaxLines:        body.MaxLines,
	})

	writeJSON(w, http.StatusOK, res)
}

type translateReqBody struct {
	Key          string `json:"key"`
	TargetLocale string `json:"target_locale"`
}

func (s *Server) handleTranslate(w http.ResponseWriter, r *http.Request) {
	var body translateReqBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	srcEntry, ok := s.sourceCat.Get(body.Key)
	if !ok {
		http.Error(w, fmt.Sprintf("key '%s' not found in source catalog", body.Key), http.StatusNotFound)
		return
	}

	constraint := s.cfg.GetKeyConstraint(body.Key)

	// Glossary lookup
	glossary := make(map[string]string)
	for _, gt := range s.cfg.Glossary {
		if val, exists := gt.Translations[body.TargetLocale]; exists {
			glossary[gt.Term] = val
		}
	}

	optRes, err := s.optimizer.Optimize(r.Context(), agent.OptimizeRequest{
		Key:           body.Key,
		SourceText:    srcEntry.Value,
		SourceLocale:  s.sourceCat.Locale,
		TargetLocale:  body.TargetLocale,
		Context:       constraint.Context,
		Constraint:    constraint,
		Glossary:      glossary,
		MaxIterations: 4,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Update in-memory catalog
	if cat, exists := s.targetCats[body.TargetLocale]; exists {
		cat.Set(&parser.Entry{
			Key:   body.Key,
			Value: optRes.FinalTranslation,
		})
	}

	// Update SQLite store
	status := "ACCEPTED"
	if optRes.FinalMeasure.IsOverflow {
		status = "OVERFLOW"
	}
	if optRes.Status == "ICU_SYNTAX_ERROR" {
		status = "ICU_ERROR"
	}

	_ = s.db.SaveTranslation(body.Key, body.TargetLocale, optRes.FinalTranslation, status, optRes.FinalMeasure, optRes.IterationsUsed)

	s.broadcastEvent("translation_updated", map[string]any{
		"key":           body.Key,
		"locale":        body.TargetLocale,
		"translation":   optRes.FinalTranslation,
		"status":        status,
		"measure":       optRes.FinalMeasure,
		"iterations":    optRes.IterationsUsed,
	})

	writeJSON(w, http.StatusOK, optRes)
}

type saveTransReqBody struct {
	Key    string `json:"key"`
	Locale string `json:"locale"`
	Value  string `json:"value"`
}

func (s *Server) handleSaveTranslation(w http.ResponseWriter, r *http.Request) {
	var body saveTransReqBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	constraint := s.cfg.GetKeyConstraint(body.Key)
	measure := s.typo.Measure(typography.MeasureRequest{
		Text:       body.Value,
		Family:     constraint.FontFamily,
		SizePX:     constraint.FontSizePX,
		Weight:     constraint.FontWeight,
		PaddingXPX: 12,
		MaxWidthPX: constraint.MaxWidthPX,
		MaxLines:   constraint.MaxLines,
	})

	status := "ACCEPTED"
	if measure.IsOverflow {
		status = "OVERFLOW"
	}

	// Check ICU
	if srcEntry, ok := s.sourceCat.Get(body.Key); ok {
		if constraint.Format == "icu_messageformat" || strings.Contains(srcEntry.Value, "{") {
			valRes := icu.Validate(srcEntry.Value, body.Value, body.Locale)
			if !valRes.Valid {
				status = "ICU_ERROR"
			}
		}
	}

	// Update in-memory catalog
	if cat, exists := s.targetCats[body.Locale]; exists {
		cat.Set(&parser.Entry{
			Key:   body.Key,
			Value: body.Value,
		})
	}

	if err := s.db.SaveTranslation(body.Key, body.Locale, body.Value, status, measure, 1); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	s.broadcastEvent("translation_updated", map[string]any{
		"key":         body.Key,
		"locale":      body.Locale,
		"translation": body.Value,
		"status":      status,
		"measure":     measure,
	})

	writeJSON(w, http.StatusOK, map[string]any{
		"status":  status,
		"measure": measure,
	})
}

func (s *Server) handleExport(w http.ResponseWriter, r *http.Request) {
	if len(s.cfg.Files) == 0 {
		http.Error(w, "no files configured in localecraft.yaml", http.StatusBadRequest)
		return
	}

	fileCfg := s.cfg.Files[0]
	exportedFiles := make([]string, 0)

	for loc, cat := range s.targetCats {
		outPath := fileCfg.ResolveOutputPath(loc, s.cfg.WorkingDir)
		if err := parser.SaveFile(cat, outPath); err != nil {
			http.Error(w, fmt.Sprintf("save file %s: %v", outPath, err), http.StatusInternalServerError)
			return
		}
		exportedFiles = append(exportedFiles, outPath)
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success":        true,
		"exported_files": exportedFiles,
	})
}

type prReqBody struct {
	Branch string `json:"branch"`
	Title  string `json:"title"`
	Body   string `json:"body"`
}

func (s *Server) handleCreatePR(w http.ResponseWriter, r *http.Request) {
	var body prReqBody
	_ = json.NewDecoder(r.Body).Decode(&body)

	if body.Branch == "" {
		body.Branch = fmt.Sprintf("localecraft/update-%d", time.Now().Unix())
	}
	if body.Title == "" {
		body.Title = "i18n: visual software localization and overflow prevention"
	}
	if body.Body == "" {
		body.Body = "Autonomous localization updates with verified container boundaries and ICU parity."
	}

	// Trigger git branch and PR via gh CLI
	cmdBranch := exec.Command("git", "checkout", "-b", body.Branch)
	cmdBranch.Dir = s.cfg.WorkingDir
	_ = cmdBranch.Run()

	cmdAdd := exec.Command("git", "add", ".")
	cmdAdd.Dir = s.cfg.WorkingDir
	_ = cmdAdd.Run()

	cmdCommit := exec.Command("git", "commit", "-m", "chore(i18n): update localecraft translations")
	cmdCommit.Dir = s.cfg.WorkingDir
	_ = cmdCommit.Run()

	cmdPush := exec.Command("git", "push", "-u", "origin", body.Branch)
	cmdPush.Dir = s.cfg.WorkingDir
	_ = cmdPush.Run()

	cmdPR := exec.Command("gh", "pr", "create", "--title", body.Title, "--body", body.Body, "--head", body.Branch)
	cmdPR.Dir = s.cfg.WorkingDir
	prOut, err := cmdPR.CombinedOutput()
	prURL := strings.TrimSpace(string(prOut))

	writeJSON(w, http.StatusOK, map[string]any{
		"success": err == nil,
		"branch":  body.Branch,
		"pr_url":  prURL,
		"error":   fmt.Sprintf("%v", err),
	})
}
