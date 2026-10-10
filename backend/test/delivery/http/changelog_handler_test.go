package http_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"openmars/internal/changelog"
	"openmars/internal/delivery/dto"
	httpHandler "openmars/internal/delivery/http"
	"openmars/test/testutil"
)

func TestChangelog_ReturnsEntries(t *testing.T) {
	response := dto.ToChangelogResponse([]changelog.Entry{
		{Version: "v7.1.2", Sections: []changelog.Section{
			{Title: "Major update: Colonies", Major: true, Intro: "Colonies is playable.", Items: []string{"Trade."}},
			{Title: "Fixed", Items: []string{"A bug."}},
		}},
		{Version: "v7.1.0", Intro: "Behind-the-scenes improvements."},
	})
	router := httpHandler.SetupRouter(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, dto.MetaResponse{}, "", response)

	request := httptest.NewRequest(http.MethodGet, "/api/v1/changelog", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)

	testutil.AssertEqual(t, http.StatusOK, recorder.Code, "status")

	var got dto.ChangelogResponse
	testutil.AssertNoError(t, json.Unmarshal(recorder.Body.Bytes(), &got), "decode changelog")
	testutil.AssertEqual(t, 2, len(got.Entries), "entry count")
	testutil.AssertEqual(t, "v7.1.2", got.Entries[0].Version, "first version")
	major := got.Entries[0].Sections[0]
	testutil.AssertTrue(t, major.Major, "major flag")
	testutil.AssertEqual(t, "Colonies is playable.", major.Intro, "major intro")
	testutil.AssertEqual(t, "Fixed", got.Entries[0].Sections[1].Title, "section title")
	testutil.AssertEqual(t, "A bug.", got.Entries[0].Sections[1].Items[0], "section item")
	testutil.AssertEqual(t, 0, len(got.Entries[1].Sections), "intro-only entry has no sections")
}

func TestChangelog_EmptyIsAnEmptyList(t *testing.T) {
	router := httpHandler.SetupRouter(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, dto.MetaResponse{}, "", dto.ToChangelogResponse(nil))

	request := httptest.NewRequest(http.MethodGet, "/api/v1/changelog", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)

	testutil.AssertEqual(t, `{"entries":[]}`+"\n", recorder.Body.String(), "body")
}

func TestChangelog_ServesReferencedImagesOnly(t *testing.T) {
	dir := t.TempDir()
	testutil.AssertNoError(t, os.MkdirAll(filepath.Join(dir, "v3"), 0o755), "mkdir")
	testutil.AssertNoError(t, os.WriteFile(filepath.Join(dir, "v3", "colonies.png"), []byte("colonies"), 0o644), "write image")
	testutil.AssertNoError(t, os.WriteFile(filepath.Join(dir, "v3", "other.png"), []byte("other"), 0o644), "write other")

	response := dto.ToChangelogResponse([]changelog.Entry{
		{Version: "v3", Sections: []changelog.Section{
			{Title: "Major update: Colonies", Major: true, Intro: "Colonies.", Image: &changelog.Image{File: "colonies.png", Alt: "Colonies"}},
		}},
	})
	router := httpHandler.SetupRouter(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, dto.MetaResponse{}, dir, response)

	get := func(path string) *httptest.ResponseRecorder {
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, path, nil))
		return recorder
	}

	served := get("/api/v1/changelog/v3/colonies.png")
	testutil.AssertEqual(t, http.StatusOK, served.Code, "referenced image status")
	testutil.AssertEqual(t, "colonies", served.Body.String(), "referenced image body")
	testutil.AssertEqual(t, "image/png", served.Header().Get("Content-Type"), "content type")

	testutil.AssertEqual(t, http.StatusNotFound, get("/api/v1/changelog/v3/other.png").Code, "unreferenced image")
	testutil.AssertEqual(t, http.StatusNotFound, get("/api/v1/changelog/v4/colonies.png").Code, "wrong version")
	testutil.AssertNotEqual(t, http.StatusOK, get("/api/v1/changelog/v3/..%2Fv3%2Fother.png").Code, "encoded traversal")
}
