package redissc

import (
	"github.com/go-errors/errors"
	"github.com/hootsuite/healthchecks"
	"testing"
)

func TestOK(t *testing.T) {

	redisStatusChecker := RedisStatusChecker{OkRedis{}}

	s := redisStatusChecker.CheckStatus("the redis")

	if len(s.StatusList) != 1 {
		t.Errorf("Length of StatusList should be 1, was %d", len(s.StatusList))
	}

	actual := s.StatusList[0]
	if actual.Result != healthchecks.OK {
		t.Errorf("Result shoud be `OK`, was `%s`", actual.Result)
	}
}

func TestError(t *testing.T) {

	redisStatusChecker := RedisStatusChecker{ErrorRedis{}}
	name := "the redis"
	s := redisStatusChecker.CheckStatus(name)

	if len(s.StatusList) != 1 {
		t.Errorf("Length of StatusList should be 1, was %d", len(s.StatusList))
	}

	actual := s.StatusList[0]
	if actual.Result != healthchecks.CRITICAL {
		t.Errorf("Result shoud be `CRITICAL`, was `%s`", actual.Result)
	}

	if actual.Description != name {
		t.Errorf("Description shoud be `%s`, was `%s`", name, actual.Description)
	}

	eDetails := "An error message"
	if actual.Details != eDetails {
		t.Errorf("Details shoud be `%s`, was `%s`", eDetails, actual.Details)
	}
}

func TestPingError(t *testing.T) {

	redisStatusChecker := RedisStatusChecker{PingErrorRedis{}}
	name := "the redis"
	s := redisStatusChecker.CheckStatus(name)

	if len(s.StatusList) != 1 {
		t.Errorf("Length of StatusList should be 1, was %d", len(s.StatusList))
	}

	actual := s.StatusList[0]
	if actual.Result != healthchecks.CRITICAL {
		t.Errorf("Result shoud be `CRITICAL`, was `%s`", actual.Result)
	}

	if actual.Description != name {
		t.Errorf("Description shoud be `%s`, was `%s`", name, actual.Description)
	}

	eDetails := "Expecting `PONG` response, got `BLAH`"
	if actual.Details != eDetails {
		t.Errorf("Details shoud be `%s`, was `%s`", eDetails, actual.Details)
	}
}

// Mocks
type OkRedis struct {
}

func (r OkRedis) Ping() (string, error) {
	return "PONG", nil
}

type ErrorRedis struct {
}

func (r ErrorRedis) Ping() (string, error) {
	return "", errors.New("An error message")
}

type PingErrorRedis struct {
}

func (r PingErrorRedis) Ping() (string, error) {
	return "BLAH", nil
}
