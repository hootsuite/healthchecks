package burrowsc

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/hootsuite/healthchecks"
)

// Burrow consumer group and partition statuses (https://github.com/linkedin/Burrow/wiki/http-request-consumer-group-status)
const (
	statusOK       = "OK"
	statusWarn     = "WARN"
	statusErr      = "ERR"
	statusStop     = "STOP"
	statusStall    = "STALL"
	statusNotFound = "NOTFOUND"
)

type lagResponse struct {
	Error     bool      `json:"error"`
	Message   string    `json:"message"`
	LagStatus lagStatus `json:"status"`
	Request   struct {
		URL  string `json:"url"`
		Host string `json:"host"`
	} `json:"request"`
}

type lagStatus struct {
	Cluster        string      `json:"cluster"`
	Group          string      `json:"group"`
	Status         string      `json:"status"`
	Complete       float64     `json:"complete"`
	Partitions     []partition `json:"partitions"`
	PartitionCount int         `json:"partition_count"`
	MaxLag         partition   `json:"maxlag"`
	TotalLag       int64       `json:"totallag"`
}

type partition struct {
	Topic     string `json:"topic"`
	Partition int    `json:"partition"`
	Owner     string `json:"owner"`
	ClientID  string `json:"client_id"`
	Status    string `json:"status"`
	Start     struct {
		Offset    int64 `json:"offset"`
		Timestamp int64 `json:"timestamp"`
		Lag       int64 `json:"lag"`
	} `json:"start"`
	End struct {
		Offset    int64 `json:"offset"`
		Timestamp int64 `json:"timestamp"`
		Lag       int64 `json:"lag"`
	} `json:"end"`
	CurrentLag int64   `json:"current_lag"`
	Complete   float64 `json:"complete"`
}

type BurrowStatusChecker struct {
	BaseUrl              string  // Base url of the Burrow API
	Cluster              string  // The Kafka cluster to monitor
	ConsumerGroup        string  // The consumer group on the cluster to monitor
	Topic                *string // Optional if monitoring the status of a specific topic, leave nil to get the consumer group status
	CriticalLagThreshold *int64  // Optional lag threshold to trigger a critical alert when exceeded
}

func (b BurrowStatusChecker) CheckStatus(name string) healthchecks.StatusList {
	baseUrl := strings.TrimSuffix(b.BaseUrl, "/")
	url := fmt.Sprintf("%s/%s/consumer/%s/lag", baseUrl, b.Cluster, b.ConsumerGroup)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return healthchecks.StatusList{
			StatusList: []healthchecks.Status{
				{
					Description: name,
					Result:      healthchecks.CRITICAL,
					Details:     err.Error(),
				},
			},
		}
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return healthchecks.StatusList{
			StatusList: []healthchecks.Status{
				{
					Description: name,
					Result:      healthchecks.CRITICAL,
					Details:     err.Error(),
				},
			},
		}
	}

	// Callers should close resp.Body when done reading from it
	// Defer the closing of the body
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		responseBody, _ := ioutil.ReadAll(resp.Body)
		return healthchecks.StatusList{
			StatusList: []healthchecks.Status{
				{
					Description: name,
					Result:      healthchecks.CRITICAL,
					Details:     fmt.Sprintf("Invalid response. Code: %d, Body: %s", resp.StatusCode, responseBody),
				},
			},
		}
	}

	lagResponse := lagResponse{}
	err = json.NewDecoder(resp.Body).Decode(&lagResponse)
	if err != nil {
		return healthchecks.StatusList{
			StatusList: []healthchecks.Status{
				{
					Description: name,
					Result:      healthchecks.CRITICAL,
					Details:     fmt.Sprintf("Error decoding json response: %s", err.Error()),
				},
			},
		}
	}

	var s healthchecks.Status
	if b.Topic == nil {
		s = checkGroupStatus(name, lagResponse.LagStatus, b.CriticalLagThreshold)
	} else {
		s = checkTopicStatus(name, *b.Topic, lagResponse.LagStatus, b.CriticalLagThreshold)
	}

	return healthchecks.StatusList{
		StatusList: []healthchecks.Status{
			s,
		},
	}
}

func checkGroupStatus(name string, lagStatus lagStatus, criticalLagThreshold *int64) healthchecks.Status {
	// If critical lag threshold is specified and exceeded, return a critical alert
	if criticalLagThreshold != nil && lagStatus.TotalLag > *criticalLagThreshold {
		return healthchecks.Status{
			Description: name,
			Result:      healthchecks.CRITICAL,
			Details:     fmt.Sprintf("%s exceeds threshold", formatConsumerGroupDetails(lagStatus)),
		}
	}

	// If critical lag threshold was not exceeded or not specified, check Burrow consumer group status
	return healthchecks.Status{
		Description: name,
		Result:      getAlertLevel(lagStatus.Status),
		Details:     formatConsumerGroupDetails(lagStatus),
	}
}

func checkTopicStatus(name string, topic string, lagStatus lagStatus, criticalLagThreshold *int64) healthchecks.Status {
	alertLevel := healthchecks.OK
	critWarnPartitions := []partition{}
	topicExists := false
	totalLag := int64(0)

	for _, partition := range lagStatus.Partitions {
		if partition.Topic == topic {
			partitionAlertLevel := getAlertLevel(partition.Status)
			topicExists = true
			totalLag += partition.CurrentLag

			if partitionAlertLevel == healthchecks.CRITICAL || partitionAlertLevel == healthchecks.WARNING {
				critWarnPartitions = append(critWarnPartitions, partition)

				// If status is already CRITICAL, don't overwrite
				if alertLevel != healthchecks.CRITICAL {
					alertLevel = partitionAlertLevel
				}
			}
		}
	}

	if !topicExists {
		return healthchecks.Status{
			Description: name,
			Result:      healthchecks.WARNING,
			Details:     fmt.Sprintf("Topic %s not found in group %s on cluster %s", topic, lagStatus.Group, lagStatus.Cluster),
		}
	}

	topicDetails := fmt.Sprintf(
		"Topic %s has total lag of %d for group %s on cluster %s",
		topic,
		totalLag,
		lagStatus.Group,
		lagStatus.Cluster,
	)
	partitionDetails := formatPartitionsDetails(critWarnPartitions)

	// If critical lag threshold is specified and exceeded, return a critical alert
	if criticalLagThreshold != nil && totalLag > *criticalLagThreshold {
		return healthchecks.Status{
			Description: name,
			Result:      healthchecks.CRITICAL,
			Details:     fmt.Sprintf("%s exceeds threshold%s", topicDetails, partitionDetails),
		}
	}

	// If critical lag threshold was not exceeded or not specified, check Burrow partition statuses
	return healthchecks.Status{
		Description: name,
		Result:      alertLevel,
		Details:     fmt.Sprintf("%s%s", topicDetails, partitionDetails),
	}
}

func getAlertLevel(status string) healthchecks.AlertLevel {
	var alertLevel healthchecks.AlertLevel

	switch status {
	case statusOK:
		alertLevel = healthchecks.OK
	case statusWarn, statusNotFound:
		alertLevel = healthchecks.WARNING
	case statusErr, statusStop, statusStall:
		alertLevel = healthchecks.CRITICAL
	default:
		alertLevel = healthchecks.WARNING
	}

	return alertLevel
}

func formatConsumerGroupDetails(status lagStatus) string {
	return fmt.Sprintf(
		"Consumer group status is %s, total lag of %d for group %s on cluster %s",
		status.Status,
		status.TotalLag,
		status.Group,
		status.Cluster,
	)
}

func formatPartitionsDetails(partitions []partition) string {
	var str strings.Builder

	for _, partition := range partitions {
		str.WriteString(fmt.Sprintf(
			", partition %d status is %s and has lag of %d",
			partition.Partition,
			partition.Status,
			partition.CurrentLag,
		))
	}

	return str.String()
}
