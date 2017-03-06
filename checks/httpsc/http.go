package httpsc

import (
	"encoding/json"
	"fmt"
	"github.com/hootsuite/healthchecks"
	"io/ioutil"
	"net/http"
	"strings"
)

type HttpStatusChecker struct {
	BaseUrl string
}

func (h HttpStatusChecker) CheckStatus(name string) healthchecks.StatusList {
	url := fmt.Sprintf("%s/status/aggregate", h.BaseUrl)
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

	if resp.StatusCode == http.StatusOK {

		var responseFormat []interface{}
		err := json.NewDecoder(resp.Body).Decode(&responseFormat)
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
		var errorDetails string

		if len(responseFormat) >= 2 {
			b, err := json.Marshal(responseFormat[1])
			if err != nil {
				errorDetails = fmt.Sprintf("Error reading repsonse details: %s", err.Error())
			} else {
				errorDetails = string(b)
			}
		} else {
			errorDetails = "Error details missing"
		}

		switch responseFormat[0] {
		case "OK":
			s = healthchecks.Status{
				Description: fmt.Sprintf("%s check OK", name),
				Result:      healthchecks.OK,
				Details:     "",
			}
		case "WARN":
			s = healthchecks.Status{
				Description: name,
				Result:      healthchecks.WARNING,
				Details:     fmt.Sprintf("%s check failed: WARN - %s", name, errorDetails),
			}
		default:
			s = healthchecks.Status{
				Description: name,
				Result:      healthchecks.CRITICAL,
				Details:     fmt.Sprintf("%s check failed: CRIT - %s", name, errorDetails),
			}
		}

		return healthchecks.StatusList{
			StatusList: []healthchecks.Status{
				s,
			},
		}

	} else {

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
}

func (h HttpStatusChecker) Traverse(traversalPath []string, action string) (string, error) {
	dependencies := ""
	if len(traversalPath) > 0 {
		dependencies = fmt.Sprintf("&dependencies=%s", strings.Join(traversalPath, ","))
	}

	url := fmt.Sprintf("%s/status/traverse?action=%s%s", h.BaseUrl, action, dependencies)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Printf("Error creating request: %s \n", err.Error())
		return "", err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error executing request: %s \n", err.Error())
		return "", err
	}

	// Callers should close resp.Body when done reading from it
	// Defer the closing of the body
	defer resp.Body.Close()
	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading response body: %s", err.Error())
		return "", err
	}

	return string(responseBody), nil
}
