package healthchecks

import (
	"fmt"
	"github.com/go-errors/errors"
	"testing"
)

var defaultServiceId = "service-id"
var emptyCustomData map[string]interface{}

var testStatusEndpointA = StatusEndpoint{
	Name:          "AAA",
	Slug:          "aaa",
	Type:          "internal",
	IsTraversable: true,
	StatusCheck:   MockStatusChecker{"AAA", OK, "all good"},
	TraverseCheck: nil,
}
var testStatusEndpointB = StatusEndpoint{
	Name:          "BBB",
	Slug:          "bbb",
	Type:          "internal",
	IsTraversable: true,
	StatusCheck:   MockStatusChecker{"BBB", OK, "all good"},
	TraverseCheck: nil,
}
var testStatusEndpointC = StatusEndpoint{
	Name:          "CCC",
	Slug:          "ccc",
	Type:          "internal",
	IsTraversable: true,
	StatusCheck:   MockStatusChecker{"CCC", OK, "all good"},
	TraverseCheck: nil,
}
var testStatusEndpoints = []StatusEndpoint{
	testStatusEndpointA,
	testStatusEndpointB,
	testStatusEndpointC,
}

var testStatusEndpointNotTraversable = StatusEndpoint{
	Name:          "SSS",
	Slug:          "sss",
	Type:          "internal",
	IsTraversable: false,
	StatusCheck:   MockStatusChecker{"SSS", OK, "all good"},
	TraverseCheck: nil,
}

var testStatusEndpointMissingTraverseChecker = StatusEndpoint{
	Name:          "TTT",
	Slug:          "ttt",
	Type:          "internal",
	IsTraversable: true,
	StatusCheck:   MockStatusChecker{"TTT", OK, "all good"},
	TraverseCheck: nil,
}

var testStatusEndpointTraversable = StatusEndpoint{
	Name:          "UUU",
	Slug:          "uuu",
	Type:          "internal",
	IsTraversable: true,
	StatusCheck:   MockStatusChecker{"UUU", OK, "all good"},
	TraverseCheck: MockStatusChecker{"UUU", OK, "all good"},
}

var testStatusEndpointTraversableError = StatusEndpoint{
	Name:          "VVV",
	Slug:          "vvv",
	Type:          "internal",
	IsTraversable: true,
	StatusCheck:   MockStatusChecker{"VVV", OK, "all good"},
	TraverseCheck: MockErrorTraverseChecker{"VVV"},
}

type MockStatusChecker struct {
	Name    string
	Result  AlertLevel
	Details string
}

func (m MockStatusChecker) CheckStatus(name string) StatusList {
	return StatusList{
		StatusList: []Status{
			{
				Description: name,
				Result:      m.Result,
				Details:     m.Details,
			},
		},
	}
}

type MockErrorTraverseChecker struct {
	Name string
}

func (m MockErrorTraverseChecker) Traverse(traversalPath []string, action string) (string, error) {
	return fmt.Sprintf(`{"Name":"%s","Body":"Hello","Time":1294706395881547000}`, m.Name), errors.New("Test Error")
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

func Test_findStatusEndpoint_found(t *testing.T) {
	statusEndpoint := FindStatusEndpoint(testStatusEndpoints, "ccc")

	if statusEndpoint == nil {
		t.Error("StatusEndpoint could not be found but is in the list.")
	}
}

func Test_findStatusEndpoint_notFound(t *testing.T) {
	statusEndpoint := FindStatusEndpoint(testStatusEndpoints, "zzz")

	if statusEndpoint != nil {
		t.Error("StatusEndpoint was found but shouldn't have been in the list.")
	}
}

func Test_findStatusEndpoint_emptySlug(t *testing.T) {
	statusEndpoint := FindStatusEndpoint(testStatusEndpoints, "")

	if statusEndpoint != nil {
		t.Error("StatusEndpoint was returned for empty argument")
	}
}
