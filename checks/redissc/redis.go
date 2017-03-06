package redissc

import (
	"fmt"
	"github.com/hootsuite/healthchecks"
)

const PONG = "PONG"

// A thin Redis wrapper used for mocks / tests
type RedisClient interface {
	Ping() (string, error)
}

type RedisStatusChecker struct {
	Client RedisClient
}

// Check the status of redis by trying to `Ping`
func (r RedisStatusChecker) CheckStatus(name string) healthchecks.StatusList {
	pong, err := r.Client.Ping()

	s := healthchecks.Status{
		Description: name,
		Result:      healthchecks.OK,
		Details:     "",
	}

	if err != nil {
		s = healthchecks.Status{
			Description: name,
			Result:      healthchecks.CRITICAL,
			Details:     err.Error(),
		}
	} else if pong != PONG {
		s = healthchecks.Status{
			Description: name,
			Result:      healthchecks.CRITICAL,
			Details:     fmt.Sprintf("Expecting `PONG` response, got `%s`", pong),
		}
	}

	return healthchecks.StatusList{StatusList: []healthchecks.Status{s}}
}
