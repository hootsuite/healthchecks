package burrowsc

import (
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"testing"

	"github.com/hootsuite/healthchecks"
	"github.com/jarcoal/httpmock"
)

func TestBurrowStatusChecker_CheckGroupStatusOK(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("GET", "http://something.com/kafka/cluster1/consumer/group1/lag",
		httpmock.NewStringResponder(200, getTestData("status_ok.json")))

	burrowStatusChecker := BurrowStatusChecker{
		BaseUrl:       "http://something.com/kafka",
		Cluster:       "cluster1",
		ConsumerGroup: "group1",
	}
	status := burrowStatusChecker.CheckStatus("Consumer")

	expected := healthchecks.StatusList{
		StatusList: []healthchecks.Status{
			{
				Description: "Consumer",
				Result:      healthchecks.OK,
				Details:     "Consumer group status is OK, total lag of 17 for group group1 on cluster cluster1",
			},
		},
	}

	if !reflect.DeepEqual(status.StatusList, expected.StatusList) {
		t.Errorf("Status response should be `%v`, was: `%v`", expected, status)
	}
}

func TestBurrowStatusChecker_CheckTopicStatusOK(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("GET", "http://something.com/kafka/cluster1/consumer/group1/lag",
		httpmock.NewStringResponder(200, getTestData("status_ok.json")))

	topic := "topic1"
	burrowStatusChecker := BurrowStatusChecker{
		BaseUrl:       "http://something.com/kafka",
		Cluster:       "cluster1",
		ConsumerGroup: "group1",
		Topic:         &topic,
	}
	status := burrowStatusChecker.CheckStatus("Consumer")

	expected := healthchecks.StatusList{
		StatusList: []healthchecks.Status{
			{
				Description: "Consumer",
				Result:      healthchecks.OK,
				Details:     "Topic topic1 has total lag of 6 for group group1 on cluster cluster1",
			},
		},
	}

	if !reflect.DeepEqual(status.StatusList, expected.StatusList) {
		t.Errorf("Status response should be `%v`, was: `%v`", expected, status)
	}
}

func TestBurrowStatusChecker_CheckGroupStatusThresholdExceeded(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("GET", "http://something.com/kafka/cluster1/consumer/group1/lag",
		httpmock.NewStringResponder(200, getTestData("status_ok.json")))

	criticalLagThreshold := int64(10)
	burrowStatusChecker := BurrowStatusChecker{
		BaseUrl:              "http://something.com/kafka",
		Cluster:              "cluster1",
		ConsumerGroup:        "group1",
		CriticalLagThreshold: &criticalLagThreshold,
	}
	status := burrowStatusChecker.CheckStatus("Consumer")

	expected := healthchecks.StatusList{
		StatusList: []healthchecks.Status{
			{
				Description: "Consumer",
				Result:      healthchecks.CRITICAL,
				Details:     "Consumer group status is OK, total lag of 17 for group group1 on cluster cluster1 exceeds threshold",
			},
		},
	}

	if !reflect.DeepEqual(status.StatusList, expected.StatusList) {
		t.Errorf("Status response should be `%v`, was: `%v`", expected, status)
	}
}

func TestBurrowStatusChecker_CheckTopicStatusThresholdExceeded(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("GET", "http://something.com/kafka/cluster1/consumer/group1/lag",
		httpmock.NewStringResponder(200, getTestData("status_ok.json")))

	topic := "topic2"
	criticalLagThreshold := int64(10)
	burrowStatusChecker := BurrowStatusChecker{
		BaseUrl:              "http://something.com/kafka",
		Cluster:              "cluster1",
		ConsumerGroup:        "group1",
		Topic:                &topic,
		CriticalLagThreshold: &criticalLagThreshold,
	}
	status := burrowStatusChecker.CheckStatus("Consumer")

	expected := healthchecks.StatusList{
		StatusList: []healthchecks.Status{
			{
				Description: "Consumer",
				Result:      healthchecks.CRITICAL,
				Details:     "Topic topic2 has total lag of 11 for group group1 on cluster cluster1 exceeds threshold",
			},
		},
	}

	if !reflect.DeepEqual(status.StatusList, expected.StatusList) {
		t.Errorf("Status response should be `%v`, was: `%v`", expected, status)
	}
}

func TestBurrowStatusChecker_CheckGroupStatusERR(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("GET", "http://something.com/kafka/cluster1/consumer/group1/lag",
		httpmock.NewStringResponder(200, getTestData("status_err.json")))

	burrowStatusChecker := BurrowStatusChecker{
		BaseUrl:       "http://something.com/kafka",
		Cluster:       "cluster1",
		ConsumerGroup: "group1",
	}
	status := burrowStatusChecker.CheckStatus("Consumer")

	expected := healthchecks.StatusList{
		StatusList: []healthchecks.Status{
			{
				Description: "Consumer",
				Result:      healthchecks.CRITICAL,
				Details:     "Consumer group status is ERR, total lag of 17717 for group group1 on cluster cluster1",
			},
		},
	}

	if !reflect.DeepEqual(status.StatusList, expected.StatusList) {
		t.Errorf("Status response should be `%v`, was: `%v`", expected, status)
	}
}

func TestBurrowStatusChecker_CheckTopicStatusSTOP(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("GET", "http://something.com/kafka/cluster1/consumer/group1/lag",
		httpmock.NewStringResponder(200, getTestData("status_err.json")))

	topic := "topic1"
	burrowStatusChecker := BurrowStatusChecker{
		BaseUrl:       "http://something.com/kafka",
		Cluster:       "cluster1",
		ConsumerGroup: "group1",
		Topic:         &topic,
	}
	status := burrowStatusChecker.CheckStatus("Consumer")

	expected := healthchecks.StatusList{
		StatusList: []healthchecks.Status{
			{
				Description: "Consumer",
				Result:      healthchecks.CRITICAL,
				Details:     "Topic topic1 has total lag of 8035 for group group1 on cluster cluster1, partition 0 status is STOP and has lag of 8035",
			},
		},
	}

	if !reflect.DeepEqual(status.StatusList, expected.StatusList) {
		t.Errorf("Status response should be `%v`, was: `%v`", expected, status)
	}
}

func TestBurrowStatusChecker_CheckTopicStatusSTALL(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("GET", "http://something.com/kafka/cluster1/consumer/group1/lag",
		httpmock.NewStringResponder(200, getTestData("status_err.json")))

	topic := "topic2"
	burrowStatusChecker := BurrowStatusChecker{
		BaseUrl:       "http://something.com/kafka",
		Cluster:       "cluster1",
		ConsumerGroup: "group1",
		Topic:         &topic,
	}
	status := burrowStatusChecker.CheckStatus("Consumer")

	expected := healthchecks.StatusList{
		StatusList: []healthchecks.Status{
			{
				Description: "Consumer",
				Result:      healthchecks.CRITICAL,
				Details:     "Topic topic2 has total lag of 9682 for group group1 on cluster cluster1, partition 1 status is STALL and has lag of 7294, partition 2 status is WARN and has lag of 2285",
			},
		},
	}

	if !reflect.DeepEqual(status.StatusList, expected.StatusList) {
		t.Errorf("Status response should be `%v`, was: `%v`", expected, status)
	}
}

func TestBurrowStatusChecker_CheckGroupStatusWARN(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("GET", "http://something.com/kafka/cluster1/consumer/group1/lag",
		httpmock.NewStringResponder(200, getTestData("status_warn.json")))

	burrowStatusChecker := BurrowStatusChecker{
		BaseUrl:       "http://something.com/kafka",
		Cluster:       "cluster1",
		ConsumerGroup: "group1",
	}
	status := burrowStatusChecker.CheckStatus("Consumer")

	expected := healthchecks.StatusList{
		StatusList: []healthchecks.Status{
			{
				Description: "Consumer",
				Result:      healthchecks.WARNING,
				Details:     "Consumer group status is WARN, total lag of 2764 for group group1 on cluster cluster1",
			},
		},
	}

	if !reflect.DeepEqual(status.StatusList, expected.StatusList) {
		t.Errorf("Status response should be `%v`, was: `%v`", expected, status)
	}
}
func TestBurrowStatusChecker_CheckTopicStatusWARN(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("GET", "http://something.com/kafka/cluster1/consumer/group1/lag",
		httpmock.NewStringResponder(200, getTestData("status_warn.json")))

	topic := "topic1"
	burrowStatusChecker := BurrowStatusChecker{
		BaseUrl:       "http://something.com/kafka",
		Cluster:       "cluster1",
		ConsumerGroup: "group1",
		Topic:         &topic,
	}
	status := burrowStatusChecker.CheckStatus("Consumer")

	expected := healthchecks.StatusList{
		StatusList: []healthchecks.Status{
			{
				Description: "Consumer",
				Result:      healthchecks.WARNING,
				Details:     "Topic topic1 has total lag of 2740 for group group1 on cluster cluster1, partition 1 status is WARN and has lag of 2736",
			},
		},
	}

	if !reflect.DeepEqual(status.StatusList, expected.StatusList) {
		t.Errorf("Status response should be `%v`, was: `%v`", expected, status)
	}
}

func TestBurrowStatusChecker_CheckTopicStatusOKGroupStatusWarn(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("GET", "http://something.com/kafka/cluster1/consumer/group1/lag",
		httpmock.NewStringResponder(200, getTestData("status_warn.json")))

	topic := "topic2"
	burrowStatusChecker := BurrowStatusChecker{
		BaseUrl:       "http://something.com/kafka",
		Cluster:       "cluster1",
		ConsumerGroup: "group1",
		Topic:         &topic,
	}
	status := burrowStatusChecker.CheckStatus("Consumer")

	expected := healthchecks.StatusList{
		StatusList: []healthchecks.Status{
			{
				Description: "Consumer",
				Result:      healthchecks.OK,
				Details:     "Topic topic2 has total lag of 24 for group group1 on cluster cluster1",
			},
		},
	}

	if !reflect.DeepEqual(status.StatusList, expected.StatusList) {
		t.Errorf("Status response should be `%v`, was: `%v`", expected, status)
	}
}

func TestBurrowStatusChecker_CheckGroupNotFound(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("GET", "http://something.com/kafka/cluster1/consumer/group1/lag",
		httpmock.NewStringResponder(200, getTestData("status_notfound.json")))

	burrowStatusChecker := BurrowStatusChecker{
		BaseUrl:       "http://something.com/kafka",
		Cluster:       "cluster1",
		ConsumerGroup: "group1",
	}
	status := burrowStatusChecker.CheckStatus("Consumer")

	expected := healthchecks.StatusList{
		StatusList: []healthchecks.Status{
			{
				Description: "Consumer",
				Result:      healthchecks.WARNING,
				Details:     "Consumer group status is NOTFOUND, total lag of 0 for group group1 on cluster cluster1",
			},
		},
	}

	if !reflect.DeepEqual(status.StatusList, expected.StatusList) {
		t.Errorf("Status response should be `%v`, was: `%v`", expected, status)
	}
}

func TestBurrowStatusChecker_CheckTopicNotFound(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("GET", "http://something.com/kafka/cluster1/consumer/group1/lag",
		httpmock.NewStringResponder(200, getTestData("status_ok.json")))

	topic := "topic3"
	burrowStatusChecker := BurrowStatusChecker{
		BaseUrl:       "http://something.com/kafka",
		Cluster:       "cluster1",
		ConsumerGroup: "group1",
		Topic:         &topic,
	}
	status := burrowStatusChecker.CheckStatus("Consumer")

	expected := healthchecks.StatusList{
		StatusList: []healthchecks.Status{
			{
				Description: "Consumer",
				Result:      healthchecks.WARNING,
				Details:     "Topic topic3 not found in group group1 on cluster cluster1",
			},
		},
	}

	if !reflect.DeepEqual(status.StatusList, expected.StatusList) {
		t.Errorf("Status response should be `%v`, was: `%v`", expected, status)
	}
}

func TestBurrowStatusChecker_CheckStatusInvalidJson(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("GET", "http://something.com/kafka/cluster1/consumer/group1/lag",
		httpmock.NewStringResponder(200, `hi`))

	burrowStatusChecker := BurrowStatusChecker{
		BaseUrl:       "http://something.com/kafka/",
		Cluster:       "cluster1",
		ConsumerGroup: "group1",
	}
	status := burrowStatusChecker.CheckStatus("Consumer")

	expected := healthchecks.StatusList{
		StatusList: []healthchecks.Status{
			{
				Description: "Consumer",
				Result:      healthchecks.CRITICAL,
				Details:     "Error decoding json response: invalid character 'h' looking for beginning of value",
			},
		},
	}

	if !reflect.DeepEqual(status.StatusList, expected.StatusList) {
		t.Errorf("Status response should be `%v`, was: `%v`", expected, status)
	}
}

func TestBurrowStatusChecker_CheckStatusNot200(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("GET", "http://something.com/kafka/cluster1/consumer/group1/lag",
		httpmock.NewStringResponder(400, "hi"))

	burrowStatusChecker := BurrowStatusChecker{
		BaseUrl:       "http://something.com/kafka/",
		Cluster:       "cluster1",
		ConsumerGroup: "group1",
	}
	status := burrowStatusChecker.CheckStatus("Consumer")

	expected := healthchecks.StatusList{
		StatusList: []healthchecks.Status{
			{
				Description: "Consumer",
				Result:      healthchecks.CRITICAL,
				Details:     "Invalid response. Code: 400, Body: hi",
			},
		},
	}

	if !reflect.DeepEqual(status.StatusList, expected.StatusList) {
		t.Errorf("Status response should be `%v`, was: `%v`", expected, status)
	}
}

func TestBurrowStatusChecker_CheckStatusTrimBaseUrl(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("GET", "http://something.com/kafka/cluster1/consumer/group1/lag",
		httpmock.NewStringResponder(200, getTestData("status_ok.json")))

	// Invalid BaseUrl with extra /
	burrowStatusChecker := BurrowStatusChecker{
		BaseUrl:       "http://something.com/kafka/",
		Cluster:       "cluster1",
		ConsumerGroup: "group1",
	}
	status := burrowStatusChecker.CheckStatus("Consumer")

	expected := healthchecks.StatusList{
		StatusList: []healthchecks.Status{
			{
				Description: "Consumer",
				Result:      healthchecks.OK,
				Details:     "Consumer group status is OK, total lag of 17 for group group1 on cluster cluster1",
			},
		},
	}

	if !reflect.DeepEqual(status.StatusList, expected.StatusList) {
		t.Errorf("Status response should be `%v`, was: `%v`", expected, status)
	}
}

func getTestData(filename string) string {
	file, e := ioutil.ReadFile(fmt.Sprintf("./test/%s", filename))
	if e != nil {
		fmt.Printf("File error: %v\n", e)
		os.Exit(1)
	}
	return string(file)
}
