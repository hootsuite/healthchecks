package healthchecks

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

/* VARIABLES */

var handler = Handler(testStatusEndpoints, "test/about.json", "test/version.txt", make(map[string]interface{}))

/* TESTS */
func TestHttpAmIUp(t *testing.T) {
	req, _ := http.NewRequest("GET", "/status/am-i-up", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assertStatusCode(http.StatusOK, t, w)
	assertContentTypeHeader("text/plain; charset=utf-8", t, w)
	assertBody("OK", t, w)
}

func TestHttpAbout(t *testing.T) {
	req, _ := http.NewRequest("GET", "/status/about?action=", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assertSuccessfulJSONResponse(t, w)

	bodyAsString := strings.TrimSpace(w.Body.String())
	testAboutResponse := AboutResponse{}
	err := json.Unmarshal(w.Body.Bytes(), &testAboutResponse)
	if err != nil {
		t.Errorf("Response body is an invalid About format, was: `%s`", bodyAsString)
	}
}

func TestHttpAggregate(t *testing.T) {
	aggregateHandler := Handler([]StatusEndpoint{
		{
			Name:          "AAA",
			Slug:          "aaa",
			Type:          "internal",
			IsTraversable: false,
			StatusCheck:   MockStatusChecker{"AAA", OK, "all good"},
			TraverseCheck: nil,
		},
		{
			Name:          "BBB",
			Slug:          "bbb",
			Type:          "internal",
			IsTraversable: false,
			StatusCheck:   MockStatusChecker{"BBB", OK, "all good"},
			TraverseCheck: nil,
		},
	},
		"test/about.json",
		"test/version.txt",
		make(map[string]interface{}),
	)
	req, _ := http.NewRequest("GET", "/status/aggregate", nil)
	w := httptest.NewRecorder()

	aggregateHandler.ServeHTTP(w, req)

	assertSuccessfulJSONResponse(t, w)
	assertBody(`["OK"]`, t, w)
}

func TestHttpInvalidEndpoint(t *testing.T) {
	req, _ := http.NewRequest("GET", "/status/something", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assertStatusCode(http.StatusNotFound, t, w)
	assertContentTypeHeader("application/json; charset=utf-8", t, w)
}

func TestHttpTraverse(t *testing.T) {
	req, _ := http.NewRequest("GET", "/status/traverse", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assertSuccessfulJSONResponse(t, w)
}

/* HELPER FUNCTIONS */
func assertSuccessfulJSONResponse(t *testing.T, w *httptest.ResponseRecorder) {
	assertStatusCode(http.StatusOK, t, w)
	assertContentTypeHeader("application/json; charset=utf-8", t, w)
}

func assertContentTypeHeader(expectedHeader string, t *testing.T, w *httptest.ResponseRecorder) {
	if w.HeaderMap["Content-Type"][0] != expectedHeader {
		t.Errorf("Content-Type should be `%s`, was: %s", expectedHeader, w.HeaderMap["Content-Type"][0])
	}
}

func assertStatusCode(expectedStatusCode int, t *testing.T, w *httptest.ResponseRecorder) {
	if w.Code != expectedStatusCode {
		t.Errorf("Content-Type should be `%d`, was: %d", expectedStatusCode, w.Code)
	}
}

func assertBody(expected string, t *testing.T, w *httptest.ResponseRecorder) {
	bodyAsString := strings.TrimSpace(w.Body.String())
	if bodyAsString != expected {
		t.Errorf("Response body should be `%s`, was: `%s`", expected, bodyAsString)
	}
}
