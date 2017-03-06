package hystrixsc

import (
	"github.com/afex/hystrix-go/hystrix"
	"github.com/hootsuite/healthchecks"
)

type HystrixStatusChecker struct {
	CommandName string
}

func (h HystrixStatusChecker) CheckStatus(name string) healthchecks.StatusList {
	s := healthchecks.Status{
		Description: name,
		Result:      healthchecks.OK,
		Details:     "",
	}

	c, _, err := hystrix.GetCircuit(h.CommandName)
	if err != nil {
		s.Result = healthchecks.CRITICAL
		s.Details = err.Error()
	} else if c.IsOpen() {
		s.Result = healthchecks.CRITICAL
		s.Details = "Circuit breaker is OPEN"
	}

	return healthchecks.StatusList{StatusList: []healthchecks.Status{s}}
}
