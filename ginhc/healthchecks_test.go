package ginhc

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/hootsuite/healthchecks"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

/* VARIABLES */
var testStatusEndpointA = healthchecks.StatusEndpoint{
	Name:          "AAA",
	Slug:          "aaa",
	Type:          "internal",
	IsTraversable: true,
	StatusCheck:   MockStatusChecker{"AAA", healthchecks.OK, "all good"},
	TraverseCheck: nil,
}
var testStatusEndpointB = healthchecks.StatusEndpoint{
	Name:          "BBB",
	Slug:          "bbb",
	Type:          "internal",
	IsTraversable: true,
	StatusCheck:   MockStatusChecker{"BBB", healthchecks.OK, "all good"},
	TraverseCheck: nil,
}
var testStatusEndpointC = healthchecks.StatusEndpoint{
	Name:          "CCC",
	Slug:          "ccc",
	Type:          "internal",
	IsTraversable: true,
	StatusCheck:   MockStatusChecker{"CCC", healthchecks.OK, "all good"},
	TraverseCheck: nil,
}
var testStatusEndpoints = []healthchecks.StatusEndpoint{
	testStatusEndpointA,
	testStatusEndpointB,
	testStatusEndpointC,
}

var engine = getGinAppEngine(testStatusEndpoints)

type MockStatusChecker struct {
	Name    string
	Result  healthchecks.AlertLevel
	Details string
}

func (m MockStatusChecker) CheckStatus(name string) healthchecks.StatusList {
	return healthchecks.StatusList{
		StatusList: []healthchecks.Status{
			{
				Description: name,
				Result:      m.Result,
				Details:     m.Details,
			},
		},
	}
}

func (m MockStatusChecker) Traverse(traversalPath []string, action string) (string, error) {
	return fmt.Sprintf(`{"Name":"%s","Body":"Hello","Time":1294706395881547000}`, m.Name), nil
}

type MockTraverseChecker struct {
	Error    error
	Response string
}

func (m MockTraverseChecker) Traverse(traversalPath []string, action string) (string, error) {
	if m.Error != nil {
		return "", m.Error
	}

	return m.Response, nil
}

/* TESTS */
func TestAmIUp(t *testing.T) {
	req, _ := http.NewRequest("GET", "/status/am-i-up", nil)
	w := httptest.NewRecorder()

	engine.ServeHTTP(w, req)

	assertStatusCode(http.StatusOK, t, w)
	assertContentTypeHeader("text/plain; charset=utf-8", t, w)
	assertBody("OK", t, w)
}

func TestAbout(t *testing.T) {
	req, _ := http.NewRequest("GET", "/status/about?action=", nil)
	w := httptest.NewRecorder()

	engine.ServeHTTP(w, req)

	assertSuccessfulJsonResponse(t, w)

	bodyAsString := strings.TrimSpace(w.Body.String())
	testAboutResponse := healthchecks.AboutResponse{}
	err := json.Unmarshal(w.Body.Bytes(), &testAboutResponse)
	if err != nil {
		t.Errorf("Response body is an invalid About format, was: `%s`", bodyAsString)
	}
}

func TestAggregate(t *testing.T) {
	var ra = getGinAppEngine(
		[]healthchecks.StatusEndpoint{
			{
				Name:          "AAA",
				Slug:          "aaa",
				Type:          "internal",
				IsTraversable: false,
				StatusCheck:   MockStatusChecker{"AAA", healthchecks.OK, "all good"},
				TraverseCheck: nil,
			},
			{
				Name:          "BBB",
				Slug:          "bbb",
				Type:          "internal",
				IsTraversable: false,
				StatusCheck:   MockStatusChecker{"BBB", healthchecks.OK, "all good"},
				TraverseCheck: nil,
			},
		},
	)
	req, _ := http.NewRequest("GET", "/status/aggregate", nil)
	w := httptest.NewRecorder()

	ra.ServeHTTP(w, req)

	assertSuccessfulJsonResponse(t, w)
	assertBody(`["OK"]`, t, w)
}

func TestInvalidEndpoint(t *testing.T) {
	req, _ := http.NewRequest("GET", "/status/something", nil)
	w := httptest.NewRecorder()

	engine.ServeHTTP(w, req)

	assertStatusCode(http.StatusNotFound, t, w)
	assertContentTypeHeader("application/json; charset=utf-8", t, w)
}

func TestTraverse(t *testing.T) {
	req, _ := http.NewRequest("GET", "/status/traverse", nil)
	w := httptest.NewRecorder()

	engine.ServeHTTP(w, req)

	assertSuccessfulJsonResponse(t, w)
}

/* HELPER FUNCTIONS */
func assertSuccessfulJsonResponse(t *testing.T, w *httptest.ResponseRecorder) {
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
		t.Errorf("Content-Type should be `%s`, was: %s", expectedStatusCode, w.Code)
	}
}

func assertBody(expected string, t *testing.T, w *httptest.ResponseRecorder) {
	bodyAsString := strings.TrimSpace(w.Body.String())
	if bodyAsString != expected {
		t.Errorf("Response body should be `%s`, was: `%s`", expected, bodyAsString)
	}
}

func getGinAppEngine(statusEndpoints []healthchecks.StatusEndpoint) *gin.Engine {
	app := gin.Default()
	var customData map[string]interface{}

	app.GET("/status/:slug", HealthChecksEndpoints(statusEndpoints, "test/about.json", "test/version.txt", customData))

	return app
}
