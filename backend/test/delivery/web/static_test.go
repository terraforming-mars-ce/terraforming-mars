package web_test

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"terraforming-mars-backend/internal/delivery/web"
	"terraforming-mars-backend/test/testutil"
)

const shellBody = "<!doctype html><title>shell</title>"

var appJS = strings.Repeat("console.log('terraforming');\n", 50)

func newHandler(t *testing.T) http.Handler {
	t.Helper()
	root := t.TempDir()
	write := func(name string, body []byte) {
		full := filepath.Join(root, filepath.FromSlash(name))
		testutil.AssertNoError(t, os.MkdirAll(filepath.Dir(full), 0o755), "mkdir")
		testutil.AssertNoError(t, os.WriteFile(full, body, 0o644), "write "+name)
	}
	var gz bytes.Buffer
	zw := gzip.NewWriter(&gz)
	_, err := zw.Write([]byte(appJS))
	testutil.AssertNoError(t, err, "gzip")
	testutil.AssertNoError(t, zw.Close(), "gzip close")

	write("index.html", []byte(shellBody))
	write("manifest.json", []byte(`{"name":"tm"}`))
	write("assets/index-abc.js", []byte(appJS))
	write("assets/index-abc.js.gz", gz.Bytes())
	write("assets/textures/mars.webp", []byte("0123456789"))
	testutil.AssertNoError(t, os.WriteFile(filepath.Join(filepath.Dir(root), "secret.txt"), []byte("secret"), 0o644), "write outside")

	handler, err := web.NewStaticHandler(root)
	testutil.AssertNoError(t, err, "new static handler")
	return handler
}

func serve(handler http.Handler, method, target string, headers map[string]string) *httptest.ResponseRecorder {
	request := httptest.NewRequest(method, target, nil)
	for key, value := range headers {
		request.Header.Set(key, value)
	}
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)
	return recorder
}

func TestStatic_CacheControlPerPathClass(t *testing.T) {
	handler := newHandler(t)
	cases := map[string]string{
		"/":                    "no-cache",
		"/index.html":          "no-cache",
		"/game/abc":            "no-cache",
		"/assets/index-abc.js": "public, max-age=31536000, immutable",
		"/manifest.json":       "public, max-age=86400",
	}
	for target, want := range cases {
		response := serve(handler, http.MethodGet, target, nil)
		testutil.AssertEqual(t, http.StatusOK, response.Code, target+" status")
		testutil.AssertEqual(t, want, response.Header().Get("Cache-Control"), target+" cache-control")
	}
}

func TestStatic_FallsBackToShellForClientRoutes(t *testing.T) {
	response := serve(newHandler(t), http.MethodGet, "/game/abc?type=join", nil)
	testutil.AssertEqual(t, http.StatusOK, response.Code, "status")
	testutil.AssertEqual(t, shellBody, response.Body.String(), "body")
	testutil.AssertTrue(t, strings.HasPrefix(response.Header().Get("Content-Type"), "text/html"), "content type")
}

func TestStatic_MissingFilesAre404(t *testing.T) {
	handler := newHandler(t)
	for _, target := range []string{"/assets/missing.js", "/assets/textures", "/missing.png"} {
		response := serve(handler, http.MethodGet, target, nil)
		testutil.AssertEqual(t, http.StatusNotFound, response.Code, target)
	}
}

func TestStatic_ServesGzipOnlyWhenAccepted(t *testing.T) {
	handler := newHandler(t)

	plain := serve(handler, http.MethodGet, "/assets/index-abc.js", nil)
	testutil.AssertEqual(t, "", plain.Header().Get("Content-Encoding"), "plain encoding")
	testutil.AssertEqual(t, appJS, plain.Body.String(), "plain body")
	testutil.AssertTrue(t, strings.Contains(plain.Header().Get("Vary"), "Accept-Encoding"), "plain vary")

	compressed := serve(handler, http.MethodGet, "/assets/index-abc.js", map[string]string{"Accept-Encoding": "br, gzip"})
	testutil.AssertEqual(t, "gzip", compressed.Header().Get("Content-Encoding"), "gzip encoding")
	testutil.AssertTrue(t, strings.HasPrefix(compressed.Header().Get("Content-Type"), "text/javascript"), "gzip content type")
	reader, err := gzip.NewReader(compressed.Body)
	testutil.AssertNoError(t, err, "gzip reader")
	decoded, err := io.ReadAll(reader)
	testutil.AssertNoError(t, err, "gzip read")
	testutil.AssertEqual(t, appJS, string(decoded), "decoded body")
	testutil.AssertNotEqual(t, plain.Header().Get("ETag"), compressed.Header().Get("ETag"), "etag per encoding")

	refused := serve(handler, http.MethodGet, "/assets/index-abc.js", map[string]string{"Accept-Encoding": "gzip;q=0"})
	testutil.AssertEqual(t, "", refused.Header().Get("Content-Encoding"), "q=0 refuses gzip")
}

func TestStatic_AllowsCrossOriginReads(t *testing.T) {
	response := serve(newHandler(t), http.MethodGet, "/index.html", map[string]string{"Origin": "https://openmars.app"})
	testutil.AssertEqual(t, "*", response.Header().Get("Access-Control-Allow-Origin"), "cors")
	testutil.AssertEqual(t, "nosniff", response.Header().Get("X-Content-Type-Options"), "nosniff")
}

func TestStatic_RevalidatesWithEtag(t *testing.T) {
	handler := newHandler(t)
	first := serve(handler, http.MethodGet, "/assets/index-abc.js", nil)
	etag := first.Header().Get("ETag")
	testutil.AssertNotEqual(t, "", etag, "etag")
	second := serve(handler, http.MethodGet, "/assets/index-abc.js", map[string]string{"If-None-Match": etag})
	testutil.AssertEqual(t, http.StatusNotModified, second.Code, "not modified")
}

func TestStatic_ServesRanges(t *testing.T) {
	response := serve(newHandler(t), http.MethodGet, "/assets/textures/mars.webp", map[string]string{"Range": "bytes=2-5"})
	testutil.AssertEqual(t, http.StatusPartialContent, response.Code, "status")
	testutil.AssertEqual(t, "2345", response.Body.String(), "range body")
}

func TestStatic_HeadOmitsBody(t *testing.T) {
	response := serve(newHandler(t), http.MethodHead, "/index.html", nil)
	testutil.AssertEqual(t, http.StatusOK, response.Code, "status")
	testutil.AssertEqual(t, 0, response.Body.Len(), "body length")
}

func TestStatic_RejectsWrites(t *testing.T) {
	response := serve(newHandler(t), http.MethodPost, "/index.html", nil)
	testutil.AssertEqual(t, http.StatusMethodNotAllowed, response.Code, "status")
	testutil.AssertEqual(t, "GET, HEAD", response.Header().Get("Allow"), "allow")
}

func TestStatic_StaysInsideTheDirectory(t *testing.T) {
	handler := newHandler(t)
	for _, target := range []string{"/../secret.txt", "/assets/../../secret.txt", "/%2e%2e/secret.txt"} {
		response := serve(handler, http.MethodGet, target, nil)
		testutil.AssertTrue(t, !strings.Contains(response.Body.String(), "secret"), target+" must not leak")
	}
}

func TestStatic_RequiresIndexHTML(t *testing.T) {
	_, err := web.NewStaticHandler(t.TempDir())
	testutil.AssertError(t, err, "missing index.html")
}
