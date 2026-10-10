// Package web serves the built frontend from disk next to the API.
package web

import (
	"fmt"
	"io/fs"
	"log/slog"
	"mime"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"openmars/internal/logger"
)

const (
	shellFile    = "index.html"
	assetsPrefix = "assets/"

	cacheImmutable  = "public, max-age=31536000, immutable"
	cacheRevalidate = "no-cache"
	cacheDefault    = "public, max-age=86400"

	// Covers the largest files (a 130 MB skybox) on slow links; the server-wide
	// WriteTimeout is sized for API responses.
	fileWriteDeadline = 10 * time.Minute
)

// Alpine has no mime.types, so Go's built-in table is all there is; these fill its gaps.
var contentTypes = map[string]string{
	".webmanifest": "application/manifest+json",
	".ico":         "image/x-icon",
	".map":         "application/json",
	".exr":         "image/x-exr",
	".glb":         "model/gltf-binary",
	".txt":         "text/plain; charset=utf-8",
}

type staticFile struct {
	path         string
	size         int64
	modTime      time.Time
	contentType  string
	cacheControl string
	gzipPath     string
	gzipSize     int64
}

type staticHandler struct {
	files map[string]*staticFile
	shell *staticFile
}

// NewStaticHandler indexes every file under dir once and serves them from disk.
// Nothing is held in memory: the frontend build is hundreds of megabytes.
// Precompressed "<file>.gz" siblings are served to clients that accept gzip.
func NewStaticHandler(dir string) (http.Handler, error) {
	root, err := filepath.Abs(dir)
	if err != nil {
		return nil, err
	}
	files := map[string]*staticFile{}
	err = filepath.WalkDir(root, func(full string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() || strings.HasSuffix(full, ".gz") {
			return nil
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(root, full)
		if err != nil {
			return err
		}
		name := filepath.ToSlash(rel)
		file := &staticFile{
			path:         full,
			size:         info.Size(),
			modTime:      info.ModTime(),
			contentType:  contentTypeFor(name),
			cacheControl: cacheControlFor(name),
		}
		if gz, err := os.Stat(full + ".gz"); err == nil && !gz.IsDir() {
			file.gzipPath = full + ".gz"
			file.gzipSize = gz.Size()
		}
		files[name] = file
		return nil
	})
	if err != nil {
		return nil, err
	}
	shell, ok := files[shellFile]
	if !ok {
		return nil, fmt.Errorf("%s has no %s", root, shellFile)
	}
	logger.Get().Info("Serving frontend", slog.String("dir", root), slog.Int("files", len(files)))
	return &staticHandler{files: files, shell: shell}, nil
}

func (h *staticHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		w.Header().Set("Allow", "GET, HEAD")
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	name := strings.TrimPrefix(path.Clean("/"+r.URL.Path), "/")
	file, ok := h.files[name]
	if !ok {
		// Unknown paths are client-side routes, but a missing file must not get the shell
		if strings.HasPrefix(name, assetsPrefix) || path.Ext(name) != "" {
			http.NotFound(w, r)
			return
		}
		file = h.shell
	}

	header := w.Header()
	// The openmars.app gateway boots this frontend from its own origin
	header.Set("Access-Control-Allow-Origin", "*")
	header.Set("X-Content-Type-Options", "nosniff")
	header.Set("Cache-Control", file.cacheControl)
	header.Set("Content-Type", file.contentType)
	header.Add("Vary", "Accept-Encoding")

	served, etag := file.path, fmt.Sprintf(`"%x-%x"`, file.size, file.modTime.UnixNano())
	gzipped := file.gzipPath != "" && acceptsGzip(r)
	if gzipped {
		served, etag = file.gzipPath, fmt.Sprintf(`"%x-%x-gz"`, file.gzipSize, file.modTime.UnixNano())
		header.Set("Content-Encoding", "gzip")
		// Ranges would address the compressed bytes, not the file the client asked for
		r.Header.Del("Range")
	}
	header.Set("ETag", etag)

	f, err := os.Open(served)
	if err != nil {
		http.Error(w, "file unavailable", http.StatusInternalServerError)
		return
	}
	defer func() { _ = f.Close() }()

	if err := http.NewResponseController(w).SetWriteDeadline(time.Now().Add(fileWriteDeadline)); err != nil {
		logger.Get().Debug("Could not extend write deadline", slog.Any("error", err))
	}
	http.ServeContent(w, r, "", file.modTime, f)
}

func contentTypeFor(name string) string {
	ext := strings.ToLower(path.Ext(name))
	if contentType, ok := contentTypes[ext]; ok {
		return contentType
	}
	if contentType := mime.TypeByExtension(ext); contentType != "" {
		return contentType
	}
	return "application/octet-stream"
}

func cacheControlFor(name string) string {
	switch {
	case name == shellFile:
		return cacheRevalidate
	case strings.HasPrefix(name, assetsPrefix):
		return cacheImmutable
	default:
		return cacheDefault
	}
}

func acceptsGzip(r *http.Request) bool {
	for _, part := range strings.Split(r.Header.Get("Accept-Encoding"), ",") {
		encoding, params, _ := strings.Cut(strings.TrimSpace(part), ";")
		if strings.EqualFold(strings.TrimSpace(encoding), "gzip") && !strings.Contains(strings.ReplaceAll(params, " ", ""), "q=0") {
			return true
		}
	}
	return false
}
