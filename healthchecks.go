package healthchecks

import (
	"encoding/json"
	"fmt"
)

/* Health checks */

type StatusResponse struct {
	Status string `json:"status"`
}

type StatusEndpoint struct {
	Name          string
	Slug          string
	Type          string
	IsTraversable bool
	StatusCheck   StatusCheck
	TraverseCheck TraverseCheck
}

type Status struct {
	Description string     `json:"description"`
	Result      AlertLevel `json:"result"`
	Details     string     `json:"details"`
}

type AlertLevel string

const (
	OK       AlertLevel = "OK"
	WARNING  AlertLevel = "WARN"
	CRITICAL AlertLevel = "CRIT"
)

type StatusList struct {
	StatusList []Status
}

// A status check for a dependency.
type StatusCheck interface {
	// Checks the status of some dependency.
	CheckStatus(name string) StatusList
}

type JsonResponse interface {
}

// TraverseCheck enables a traversal in the service graph.
type TraverseCheck interface {
	/* Traverse to the next level in the service graph. This function should be implemented by
	   SDKs or clients that want to enable service traversal.

	   This function should:
	   1. Call the dependent service and pass along both 'traversalPath' and 'action' params
	   2. Return the response from the dependent service without modifying it */
	Traverse(traversalPath []string, action string) (string, error)
}

func SerializeStatusList(s StatusList, apiVersion int) string {
	if apiVersion == APIV2 {
		statusListJSONResponse := translateStatusListV2(s)

		statusListJSON, err := json.Marshal(statusListJSONResponse)
		if err != nil {
			details := fmt.Sprintf("Error serializing StatusList: %v error: %s apiVersion: %v", s, err, apiVersion)
			fmt.Print(details)
			return fmt.Sprintf(`{"description":"Invalid StatusList","result":"CRIT","details":"%s"}`, details)
		}

		return string(statusListJSON)

	}

	statusListJSONResponse := translateStatusList(s)

	statusListJSON, err := json.Marshal(statusListJSONResponse)
	if err != nil {
		details := fmt.Sprintf("Error serializing StatusList: %v error: %s apiVersion: %v", s, err, apiVersion)
		fmt.Print(details)
		return fmt.Sprintf(`["CRIT",{"description":"Invalid StatusList","result":"CRIT","details":"%s"}]`, details)
	}

	return string(statusListJSON)
}

func translateStatusList(s StatusList) []JsonResponse {
	if len(s.StatusList) <= 0 {
		return []JsonResponse{
			CRITICAL,
			Status{
				Description: "Invalid status response",
				Result:      CRITICAL,
				Details:     "StatusList empty",
			},
		}
	}

	r := s.StatusList[0]

	if r.Result == OK {
		return []JsonResponse{
			OK,
		}
	}

	return []JsonResponse{
		r.Result,
		r,
	}
}

func translateStatusListV2(s StatusList) JsonResponse {
	if len(s.StatusList) <= 0 {
		return Status{
			Description: "Invalid status response",
			Result:      CRITICAL,
			Details:     "StatusList empty",
		}
	}

	return s.StatusList[0]
}

func ExecuteStatusCheck(s *StatusEndpoint, apiVersion int) string {
	result := s.StatusCheck.CheckStatus(s.Name)
	return SerializeStatusList(result, apiVersion)
}

// Find the StatusEndpoint given the slug (aka Status Path) to search for.
// The function will return the StatusEndpoint if found. If not found, returns nil
// If the slug is empty, it will also return nil
func FindStatusEndpoint(statusEndpoints []StatusEndpoint, slug string) *StatusEndpoint {
	if slug == "" {
		return nil
	}

	for _, s := range statusEndpoints {
		if slug == s.Slug {
			return &s
		}
	}
	return nil
}
