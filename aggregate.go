package healthchecks

import (
	"sync"
)

// Execute all statusEndpoint StatusCheck() functions asynchronously and return the
// overall status by returning the highest severity item in the following order:
// CRIT, WARN, OK
func Aggregate(statusEndpoints []StatusEndpoint, typeFilter string, apiVersion int) string {

	if len(typeFilter) > 0 {
		if typeFilter != "internal" && typeFilter != "external" {
			sl := StatusList{
				StatusList: []Status{
					{
						Description: "Invalid type",
						Result:      CRITICAL,
						Details:     "Unknown check type given for aggregate check",
					},
				},
			}

			return SerializeStatusList(sl, apiVersion)
		}
	}

	s := statusEndpoints
	if typeFilter != "" {
		s = []StatusEndpoint{}
		for _, statusEndpoint := range statusEndpoints {
			if typeFilter == "internal" {
				if statusEndpoint.Type == "internal" {
					s = append(s, statusEndpoint)
				}
			} else if typeFilter == "external" {
				if statusEndpoint.Type != "internal" {
					s = append(s, statusEndpoint)
				}
			}
		}
	}

	responses := make(chan StatusList)

	var wg sync.WaitGroup
	wg.Add(len(s))

	for _, statusEndpoint := range s {
		go func(statusEndpoint StatusEndpoint) {
			responses <- statusEndpoint.StatusCheck.CheckStatus(statusEndpoint.Name)
		}(statusEndpoint)
	}

	var crits []StatusList
	var warns []StatusList
	var oks []StatusList

	go func() {
		for r := range responses {
			switch r.StatusList[0].Result {
			case CRITICAL:
				crits = append(crits, r)
			case WARNING:
				warns = append(warns, r)
			case OK:
				oks = append(oks, r)
			default:
				panic("Invalid AlertLevel")
			}
			wg.Done()
		}
	}()

	wg.Wait()
	close(responses)

	sl := StatusList{
		StatusList: []Status{
			{
				Description: "Aggregate Check",
				Result:      OK,
				Details:     "",
			},
		},
	}

	if len(crits) > 0 {
		sl = crits[0]
	} else if len(warns) > 0 {
		sl = warns[0]
	}

	return SerializeStatusList(sl, apiVersion)
}
