package httpsc

import (
	"github.com/jarcoal/httpmock"
	"github.com/hootsuite/healthchecks"
	"reflect"
	"testing"
)

func TestHttpStatusChecker_CheckStatusOK(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("GET", "http://something.com/status/aggregate",
		httpmock.NewStringResponder(200, `["OK"]`))

	httpStatusChecker := HttpStatusChecker{BaseUrl: "http://something.com"}
	status := httpStatusChecker.CheckStatus("Service name")

	expected := healthchecks.StatusList{
		StatusList: []healthchecks.Status{
			{
				Description: "Service name check OK",
				Result:      healthchecks.OK,
				Details:     "",
			},
		},
	}

	if !reflect.DeepEqual(status.StatusList, expected.StatusList) {
		t.Errorf("Status response should be `%v`, was: `%v`", expected, status)
	}
}

func TestHttpStatusChecker_CheckStatusWARN(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("GET", "http://something.com/status/aggregate",
		httpmock.NewStringResponder(200, `["WARN",{"description":"AAA","result":"WARN","details":"this is a warning"}]`))

	httpStatusChecker := HttpStatusChecker{BaseUrl: "http://something.com"}
	status := httpStatusChecker.CheckStatus("AAA")

	expected := healthchecks.StatusList{
		StatusList: []healthchecks.Status{
			{
				Description: "AAA",
				Result:      healthchecks.WARNING,
				Details:     "AAA check failed: WARN - {\"description\":\"AAA\",\"details\":\"this is a warning\",\"result\":\"WARN\"}",
			},
		},
	}

	if !reflect.DeepEqual(status.StatusList, expected.StatusList) {
		t.Errorf("Status response should be `%v`, was: `%v`", expected, status)
	}
}

func TestHttpStatusChecker_CheckStatusCRIT(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("GET", "http://something.com/status/aggregate",
		httpmock.NewStringResponder(200, `["CRIT",{"description":"AAA","result":"CRIT","details":"this is an error"}]`))

	httpStatusChecker := HttpStatusChecker{BaseUrl: "http://something.com"}
	status := httpStatusChecker.CheckStatus("AAA")

	expected := healthchecks.StatusList{
		StatusList: []healthchecks.Status{
			{
				Description: "AAA",
				Result:      healthchecks.CRITICAL,
				Details:     "AAA check failed: CRIT - {\"description\":\"AAA\",\"details\":\"this is an error\",\"result\":\"CRIT\"}",
			},
		},
	}

	if !reflect.DeepEqual(status.StatusList, expected.StatusList) {
		t.Errorf("Status response should be `%v`, was: `%v`", expected, status)
	}
}

func TestHttpStatusChecker_CheckStatusInvalidJson(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("GET", "http://something.com/status/aggregate",
		httpmock.NewStringResponder(200, `hi`))

	httpStatusChecker := HttpStatusChecker{BaseUrl: "http://something.com"}
	status := httpStatusChecker.CheckStatus("AAA")

	expected := healthchecks.StatusList{
		StatusList: []healthchecks.Status{
			{
				Description: "AAA",
				Result:      healthchecks.CRITICAL,
				Details:     "Error decoding json response: invalid character 'h' looking for beginning of value",
			},
		},
	}

	if !reflect.DeepEqual(status.StatusList, expected.StatusList) {
		t.Errorf("Status response should be `%v`, was: `%v`", expected, status)
	}
}

func TestHttpStatusChecker_CheckStatusNot200(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("GET", "http://something.com/status/aggregate",
		httpmock.NewStringResponder(400, "hi"))

	httpStatusChecker := HttpStatusChecker{BaseUrl: "http://something.com"}
	status := httpStatusChecker.CheckStatus("AAA")

	expected := healthchecks.StatusList{
		StatusList: []healthchecks.Status{
			{
				Description: "AAA",
				Result:      healthchecks.CRITICAL,
				Details:     "Invalid response. Code: 400, Body: hi",
			},
		},
	}

	if !reflect.DeepEqual(status.StatusList, expected.StatusList) {
		t.Errorf("Status response should be `%v`, was: `%v`", expected, status)
	}
}

func TestHttpStatusChecker_Traverse(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("GET", "http://something.com/status/traverse",
		httpmock.NewStringResponder(200, `something`))

	httpStatusChecker := HttpStatusChecker{BaseUrl: "http://something.com"}
	traverseResponse, err := httpStatusChecker.Traverse([]string{"aaa"}, "about")

	expected := `something`
	if traverseResponse != expected {
		t.Errorf("Traverse response should be `%s`, was: `%s`", expected, traverseResponse)
	}

	if err != nil {
		t.Errorf("Error should be nil")
	}
}
