package healthchecks

import (
	"testing"
)

func TestAggregateOK(t *testing.T) {
	statusEndpoints := []StatusEndpoint{
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
	}

	aggregateResponse := Aggregate(statusEndpoints, "", APIV1)
	expected := `["OK"]`
	if aggregateResponse != expected {
		t.Errorf("Response body should be `%s`, was: `%s`", expected, aggregateResponse)
	}
}

func TestAggregateCRIT(t *testing.T) {
	statusEndpoints := []StatusEndpoint{
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
			StatusCheck:   MockStatusChecker{"BBB", CRITICAL, "explosion"},
			TraverseCheck: nil,
		},
		{
			Name:          "CCC",
			Slug:          "ccc",
			Type:          "internal",
			IsTraversable: false,
			StatusCheck:   MockStatusChecker{"CCC", WARNING, "warning"},
			TraverseCheck: nil,
		},
	}

	aggregateResponse := Aggregate(statusEndpoints, "", APIV1)
	expected := `["CRIT",{"description":"BBB","result":"CRIT","details":"explosion"}]`
	if aggregateResponse != expected {
		t.Errorf("Response body should be `%s`, was: `%s`", expected, aggregateResponse)
	}
}

func TestAggregateWARN(t *testing.T) {
	statusEndpoints := []StatusEndpoint{
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
			StatusCheck:   MockStatusChecker{"BBB", WARNING, "this is a warning"},
			TraverseCheck: nil,
		},
		{
			Name:          "CCC",
			Slug:          "ccc",
			Type:          "internal",
			IsTraversable: false,
			StatusCheck:   MockStatusChecker{"CCC", OK, "all good"},
			TraverseCheck: nil,
		},
	}

	aggregateResponse := Aggregate(statusEndpoints, "", APIV1)
	expected := `["WARN",{"description":"BBB","result":"WARN","details":"this is a warning"}]`
	if aggregateResponse != expected {
		t.Errorf("Response body should be `%s`, was: `%s`", expected, aggregateResponse)
	}
}

func TestAggregateInvalidType(t *testing.T) {
	statusEndpoints := []StatusEndpoint{
		{
			Name:          "AAA",
			Slug:          "aaa",
			Type:          "internal",
			IsTraversable: false,
			StatusCheck:   MockStatusChecker{"AAA", OK, "all good"},
			TraverseCheck: nil,
		},
	}

	aggregateResponse := Aggregate(statusEndpoints, "something", APIV1)
	expected := `["CRIT",{"description":"Invalid type","result":"CRIT","details":"Unknown check type given for aggregate check"}]`
	if aggregateResponse != expected {
		t.Errorf("Response body should be `%s`, was: `%s`", expected, aggregateResponse)
	}
}

func TestAggregateInternalType(t *testing.T) {
	statusEndpoints := []StatusEndpoint{
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
			Type:          "http",
			IsTraversable: false,
			StatusCheck:   MockStatusChecker{"BBB", CRITICAL, "Not good"},
			TraverseCheck: nil,
		},
	}

	aggregateResponse := Aggregate(statusEndpoints, "internal", APIV1)
	expected := `["OK"]`
	if aggregateResponse != expected {
		t.Errorf("Response body should be `%s`, was: `%s`", expected, aggregateResponse)
	}
}

func TestAggregateExternalType(t *testing.T) {
	statusEndpoints := []StatusEndpoint{
		{
			Name:          "AAA",
			Slug:          "aaa",
			Type:          "internal",
			IsTraversable: false,
			StatusCheck:   MockStatusChecker{"AAA", CRITICAL, "Not good"},
			TraverseCheck: nil,
		},
		{
			Name:          "BBB",
			Slug:          "bbb",
			Type:          "http",
			IsTraversable: false,
			StatusCheck:   MockStatusChecker{"BBB", OK, "all good"},
			TraverseCheck: nil,
		},
	}

	aggregateResponse := Aggregate(statusEndpoints, "external", APIV1)
	expected := `["OK"]`
	if aggregateResponse != expected {
		t.Errorf("Response body should be `%s`, was: `%s`", expected, aggregateResponse)
	}
}
