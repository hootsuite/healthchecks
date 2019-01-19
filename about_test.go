package healthchecks

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAboutResponse(t *testing.T) {
	aboutResponseString := About(testStatusEndpoints, ABOUT_PROTOCOL_HTTP, "test/about.json", "test/version.txt", emptyCustomData, APIV1, true)

	testAboutResponse := AboutResponse{}
	err := json.Unmarshal([]byte(aboutResponseString), &testAboutResponse)
	if err != nil {
		t.Errorf("Response body is an invalid About format, was: `%s`", aboutResponseString)
	}

	assertEqualAboutData(t, testAboutResponse, emptyCustomData, defaultServiceId)
}

func TestAboutEmptyAboutData(t *testing.T) {
	aboutResponseString := About(testStatusEndpoints, ABOUT_PROTOCOL_HTTP, "", "", emptyCustomData, APIV1, true)

	testAboutResponse := AboutResponse{}
	err := json.Unmarshal([]byte(aboutResponseString), &testAboutResponse)
	if err != nil {
		t.Errorf("Response body is an invalid About format, was: `%s`", aboutResponseString)
	}

	assertDefaultAboutResponse(t, testAboutResponse)
	assertEmptyVersionResponse(t, testAboutResponse)
	assert.Len(t, testAboutResponse.Dependencies, 3)
	assertEqualAboutDependency(t, testAboutResponse.Dependencies[0], testStatusEndpointA)
	assertEqualAboutDependency(t, testAboutResponse.Dependencies[1], testStatusEndpointB)
	assertEqualAboutDependency(t, testAboutResponse.Dependencies[2], testStatusEndpointC)
}

func TestAboutFieldMissingAboutData(t *testing.T) {
	aboutResponseString := About(testStatusEndpoints, ABOUT_PROTOCOL_HTTP, "test/service-id-field-missing.json", "test/version.txt", emptyCustomData, APIV1, true)

	testAboutResponse := AboutResponse{}
	err := json.Unmarshal([]byte(aboutResponseString), &testAboutResponse)
	if err != nil {
		t.Errorf("Response body is an invalid About format, was: `%s`", aboutResponseString)
	}

	assertEqualAboutData(t, testAboutResponse, emptyCustomData, "N/A")
}

func TestAboutCustomData(t *testing.T) {
	serviceCustomData := make(map[string]interface{})
	aboutResponseString := About(testStatusEndpoints, ABOUT_PROTOCOL_HTTP, "test/about-custom.json", "test/version.txt", serviceCustomData, APIV1, true)

	testAboutResponse := AboutResponse{}
	err := json.Unmarshal([]byte(aboutResponseString), &testAboutResponse)
	if err != nil {
		t.Errorf("Response body is an invalid About format, was: `%s`", aboutResponseString)
	}

	var customData = buildCustomData()
	assertEqualAboutData(t, testAboutResponse, customData, defaultServiceId)
}

func TestAboutServiceCustomData(t *testing.T) {
	var serviceCustomData map[string]interface{}
	serviceCustomData = make(map[string]interface{})
	serviceCustomData["some-key"] = "some-value"

	aboutResponseString := About(testStatusEndpoints, ABOUT_PROTOCOL_HTTP, "test/about.json", "test/version.txt", serviceCustomData, APIV1, true)

	testAboutResponse := AboutResponse{}
	err := json.Unmarshal([]byte(aboutResponseString), &testAboutResponse)
	if err != nil {
		t.Errorf("Response body is an invalid About format, was: `%s`", aboutResponseString)
	}

	assertEqualAboutData(t, testAboutResponse, serviceCustomData, defaultServiceId)
}

func TestAboutServiceOverwritesCustomData(t *testing.T) {
	var serviceCustomData map[string]interface{}
	serviceCustomData = make(map[string]interface{})
	serviceCustomData["custom1"] = [2]string{"cool", "bogus"}
	serviceCustomData["custom2"] = 111
	serviceCustomData["custom3"] = "howdy"
	serviceCustom4 := make(map[string]interface{})
	serviceCustom4["innerCustom1"] = 1
	serviceCustom4["innerCustom2"] = "innerCustom2"
	serviceCustomData["custom4"] = serviceCustom4
	serviceCustomData["custom5"] = true

	aboutResponseString := About(testStatusEndpoints, ABOUT_PROTOCOL_HTTP, "test/about-custom.json", "test/version.txt", serviceCustomData, APIV1, true)

	testAboutResponse := AboutResponse{}
	err := json.Unmarshal([]byte(aboutResponseString), &testAboutResponse)
	if err != nil {
		t.Errorf("Response body is an invalid About format, was: `%s`", aboutResponseString)
	}

	var customData = buildCustomData()
	custom1 := make([]interface{}, 2)
	custom1[0] = "cool"
	custom1[1] = "bogus"
	customData["custom1"] = custom1
	customData["custom2"] = float64(111)
	customData["custom3"] = "howdy"
	custom4 := make(map[string]interface{})
	custom4["innerCustom1"] = float64(1)
	custom4["innerCustom2"] = "innerCustom2"
	customData["custom4"] = custom4
	customData["custom5"] = true
	assertEqualAboutData(t, testAboutResponse, customData, defaultServiceId)
}

func TestAboutDoesNotCheckStatus(t *testing.T) {
	aboutResponseString := About(testStatusEndpoints, ABOUT_PROTOCOL_HTTP, "", "", emptyCustomData, APIV2, false)

	testAboutResponse := AboutResponse{}
	err := json.Unmarshal([]byte(aboutResponseString), &testAboutResponse)
	if err != nil {
		t.Errorf("Response body is an invalid About format, was: `%s`", aboutResponseString)
	}

	assertDefaultAboutResponse(t, testAboutResponse)
	assertEmptyVersionResponse(t, testAboutResponse)

	// TODO: COMPLETE TEST
}

func assertEqualAboutData(t *testing.T, aboutResponse AboutResponse, customData map[string]interface{}, serviceId string) {
	assert.Equal(t, aboutResponse.Id, serviceId)
	assert.Equal(t, aboutResponse.Name, "Test Service")
	assert.Equal(t, aboutResponse.Description, "A test service")
	assert.Equal(t, aboutResponse.Owners[0], "Test1 Testerson <test1.testerson@hootsuite.com>")
	assert.Equal(t, aboutResponse.Owners[1], "Test2 Testerson <test2.testerson@hootsuite.com>")
	assert.Equal(t, aboutResponse.ProjectHome, "https://home.com/hootsuite/test-service")
	assert.Equal(t, aboutResponse.ProjectRepo, "https://github.com/hootsuite/test-service")
	assert.Equal(t, aboutResponse.LogsLinks[0], "https://logging.com/hootsuite/test-service-1")
	assert.Equal(t, aboutResponse.LogsLinks[1], "https://logging.com/hootsuite/test-service-2")
	assert.Equal(t, aboutResponse.StatsLinks[0], "https://stats.com/hootsuite/test-service-1")
	assert.Equal(t, aboutResponse.StatsLinks[1], "https://stats.com/hootsuite/test-service-2")
	assert.Equal(t, aboutResponse.Protocol, "http")
	assert.Equal(t, aboutResponse.CustomData, customData)

	assert.Len(t, aboutResponse.Dependencies, 3)
	assertEqualAboutDependency(t, aboutResponse.Dependencies[0], testStatusEndpointA)
	assertEqualAboutDependency(t, aboutResponse.Dependencies[1], testStatusEndpointB)
	assertEqualAboutDependency(t, aboutResponse.Dependencies[2], testStatusEndpointC)
}

func assertEqualAboutDependency(t *testing.T, dependency Dependency, statusEndpoint StatusEndpoint) {
	assert.Equal(t, statusEndpoint.Name, dependency.Name)
	assert.Equal(t, statusEndpoint.IsTraversable, dependency.IsTraversable)
	assert.Equal(t, statusEndpoint.Slug, dependency.StatusPath)
	assert.Equal(t, statusEndpoint.Type, dependency.Type)
	assert.True(t, dependency.StatusDuration > 0)
	assert.NotEmpty(t, dependency.Status)
}

func assertDefaultAboutResponse(t *testing.T, response AboutResponse) {
	assert.Equal(t, response.Name, "N/A")
	assert.Equal(t, response.Description, "N/A")
	assert.Equal(t, len(response.Owners), 0)
	assert.Equal(t, response.ProjectHome, "N/A")
	assert.Equal(t, response.ProjectRepo, "N/A")
	assert.Equal(t, response.Protocol, "http")
	assert.Equal(t, len(response.LogsLinks), 0)
	assert.Equal(t, len(response.StatsLinks), 0)
}

func assertEmptyVersionResponse(t *testing.T, response AboutResponse) {
	assert.Equal(t, response.Version, "N/A")
}

func buildCustomData() map[string]interface{} {
	/* The default concrete Go types are:
	- bool for JSON booleans
	- float64 for JSON numbers
	- string for JSON strings
	- nil for JSON null
	Make sure we test each type in our custom JSON */
	var data map[string]interface{}
	data = make(map[string]interface{})

	custom1 := make([]interface{}, 2)
	custom1[0] = "custom1-val1"
	custom1[1] = "custom1-val2"
	data["custom1"] = custom1

	data["custom2"] = float64(789)
	data["custom3"] = "custom3"

	custom4 := make(map[string]interface{})
	custom4["innerCustom1"] = float64(1)
	custom4["innerCustom2"] = "innerCustom2"
	innerCustom3 := make([]interface{}, 2)
	innerCustom3[0] = "innerCustom3-val1"
	innerCustom3[1] = "innerCustom3-val2"
	custom4["innerCustom3"] = innerCustom3
	data["custom4"] = custom4

	data["custom5"] = true
	data["custom6"] = nil

	return data
}
