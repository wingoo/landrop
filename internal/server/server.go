package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"mime"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/wingoo/landrop/internal/fs"
)

type Config struct {
	SaveDir          string
	Token            string
	Once             bool
	TextToStdout     bool
	ClipboardEnabled bool
	ClipboardReady   bool
	CopyToClipboard  func(string) error
	MaxBodyBytes     int64
	TextMaxBytes     int64
}

type Server struct {
	cfg            Config
	onFirstSuccess func()
	onceTrigger    sync.Once
}

func New(cfg Config) *Server {
	if cfg.MaxBodyBytes <= 0 {
		cfg.MaxBodyBytes = 200 * 1024 * 1024
	}
	if cfg.TextMaxBytes <= 0 {
		cfg.TextMaxBytes = 5 * 1024 * 1024
	}
	return &Server{cfg: cfg}
}

func (s *Server) SetOnFirstSuccess(fn func()) {
	s.onFirstSuccess = fn
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.handleIndex)
	mux.HandleFunc("/upload", s.handleUpload)
	mux.HandleFunc("/text", s.handleText)
	return mux
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if !s.authorized(r) {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := renderIndex(w, uiData{
		Token:            s.cfg.Token,
		ClipboardEnabled: s.cfg.ClipboardEnabled,
		ClipboardReady:   s.cfg.ClipboardReady,
	}); err != nil {
		http.Error(w, "render failed", http.StatusInternalServerError)
	}
}

func (s *Server) handleUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	if !s.authorized(r) {
		writeJSONError(w, http.StatusForbidden, "forbidden")
		return
	}

	start := time.Now()
	r.Body = http.MaxBytesReader(w, r.Body, s.cfg.MaxBodyBytes)
	if err := r.ParseMultipartForm(16 << 20); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid multipart body")
		return
	}
	files := r.MultipartForm.File["file"]
	if len(files) == 0 {
		writeJSONError(w, http.StatusBadRequest, "missing file")
		return
	}

	type savedItem struct {
		Name string `json:"name"`
		Path string `json:"path"`
	}
	resp := struct {
		OK    bool        `json:"ok"`
		Saved []savedItem `json:"saved"`
	}{OK: true}

	for _, fh := range files {
		f, err := fh.Open()
		if err != nil {
			writeJSONError(w, http.StatusBadRequest, "open upload failed")
			return
		}
		fullPath, finalName, size, err := fs.SaveReader(s.cfg.SaveDir, fh.Filename, f)
		_ = f.Close()
		if err != nil {
			writeJSONError(w, http.StatusInternalServerError, "save failed")
			return
		}
		resp.Saved = append(resp.Saved, savedItem{Name: finalName, Path: fullPath})
		log.Printf("received file ip=%s original=%q saved=%q bytes=%d", clientIP(r), fh.Filename, fullPath, size)
	}

	writeJSON(w, http.StatusOK, resp)
	log.Printf("upload done ip=%s count=%d duration=%s", clientIP(r), len(resp.Saved), time.Since(start).Round(time.Millisecond))
	s.maybeExitOnce()
}

func (s *Server) handleText(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	if !s.authorized(r) {
		writeJSONError(w, http.StatusForbidden, "forbidden")
		return
	}

	start := time.Now()
	r.Body = http.MaxBytesReader(w, r.Body, s.cfg.TextMaxBytes)
	contentType := r.Header.Get("Content-Type")
	mediaType, _, _ := mime.ParseMediaType(contentType)

	var text string
	switch mediaType {
	case "text/plain", "":
		b, err := io.ReadAll(r.Body)
		if err != nil {
			writeTextReadError(w, err)
			return
		}
		text = string(b)
	case "application/x-www-form-urlencoded", "multipart/form-data":
		if err := r.ParseForm(); err != nil {
			writeTextReadError(w, err)
			return
		}
		text = r.FormValue("text")
	default:
		writeJSONError(w, http.StatusUnsupportedMediaType, "unsupported content-type")
		return
	}

	if strings.TrimSpace(text) == "" {
		writeJSONError(w, http.StatusBadRequest, "empty text")
		return
	}

	clipboardCopied := false
	if s.cfg.ClipboardEnabled && s.cfg.ClipboardReady && s.cfg.CopyToClipboard != nil {
		if err := s.cfg.CopyToClipboard(text); err != nil {
			log.Printf("clipboard copy failed: %v", err)
		} else {
			clipboardCopied = true
		}
	}

	resp := struct {
		OK        bool   `json:"ok"`
		Mode      string `json:"mode"`
		Saved     string `json:"saved"`
		Path      string `json:"path"`
		Clipboard bool   `json:"clipboard"`
	}{
		OK:        true,
		Clipboard: clipboardCopied,
	}

	if s.cfg.TextToStdout {
		fmt.Printf("\n----- landrop text (ip=%s at %s) -----\n%s\n----- end -----\n\n", clientIP(r), time.Now().Format(time.RFC3339), text)
		resp.Mode = "stdout"
		writeJSON(w, http.StatusOK, resp)
		log.Printf("received text ip=%s output=stdout chars=%d clipboard=%t duration=%s", clientIP(r), len([]rune(text)), clipboardCopied, time.Since(start).Round(time.Millisecond))
		s.maybeExitOnce()
		return
	}

	name := fmt.Sprintf("text_%s.txt", time.Now().Format("20060102_150405"))
	fullPath, finalName, _, err := fs.SaveText(s.cfg.SaveDir, name, text)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "save text failed")
		return
	}
	resp.Mode = "save"
	resp.Saved = finalName
	resp.Path = fullPath
	writeJSON(w, http.StatusOK, resp)

	log.Printf("received text ip=%s saved=%q chars=%d clipboard=%t duration=%s", clientIP(r), fullPath, len([]rune(text)), clipboardCopied, time.Since(start).Round(time.Millisecond))
	s.maybeExitOnce()
}

func (s *Server) authorized(r *http.Request) bool {
	if s.cfg.Token == "" {
		return true
	}
	return r.URL.Query().Get("t") == s.cfg.Token
}

func (s *Server) maybeExitOnce() {
	if !s.cfg.Once || s.onFirstSuccess == nil {
		return
	}
	s.onceTrigger.Do(func() {
		go s.onFirstSuccess()
	})
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

func writeJSONError(w http.ResponseWriter, code int, msg string) {
	writeJSON(w, code, map[string]any{
		"ok":    false,
		"error": msg,
	})
}

func writeTextReadError(w http.ResponseWriter, err error) {
	var maxErr *http.MaxBytesError
	if errors.As(err, &maxErr) {
		writeJSONError(w, http.StatusRequestEntityTooLarge, "text payload too large")
		return
	}
	writeJSONError(w, http.StatusBadRequest, "invalid text body")
}

func clientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil {
		return host
	}
	return r.RemoteAddr
}
