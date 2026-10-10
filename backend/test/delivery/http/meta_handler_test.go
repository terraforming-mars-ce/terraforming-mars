package http_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"openmars/internal/delivery/dto"
	httpHandler "openmars/internal/delivery/http"
	"openmars/test/testutil"
)

func TestMeta_ReturnsServerIdentity(t *testing.T) {
	meta := dto.MetaResponse{Alias: "saffronbun", Name: "Saffronbun", Version: "v1.2.3"}
	router := httpHandler.SetupRouter(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, meta)

	request := httptest.NewRequest(http.MethodGet, "/api/v1/meta", nil)
	request.Header.Set("Origin", "https://gateway.example")
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)

	testutil.AssertEqual(t, http.StatusOK, recorder.Code, "status")
	testutil.AssertEqual(t, "https://gateway.example", recorder.Header().Get("Access-Control-Allow-Origin"), "CORS origin")

	var got dto.MetaResponse
	testutil.AssertNoError(t, json.Unmarshal(recorder.Body.Bytes(), &got), "decode meta")
	testutil.AssertEqual(t, meta, got, "meta body")
}
