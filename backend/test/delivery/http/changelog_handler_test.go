package http_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
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
	router := httpHandler.SetupRouter(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, dto.MetaResponse{}, response)

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
	router := httpHandler.SetupRouter(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, dto.MetaResponse{}, dto.ToChangelogResponse(nil))

	request := httptest.NewRequest(http.MethodGet, "/api/v1/changelog", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)

	testutil.AssertEqual(t, `{"entries":[]}`+"\n", recorder.Body.String(), "body")
}
