package hystrixsc

import (
	"errors"
	"github.com/afex/hystrix-go/hystrix"
	"github.com/hootsuite/healthchecks"
	"testing"
	"time"
)

const COMMAND_NAME = "test-command"

var statusChecker = HystrixStatusChecker{CommandName: COMMAND_NAME}

func init() {
	hystrix.ConfigureCommand(COMMAND_NAME, hystrix.CommandConfig{
		Timeout:                5000,
		MaxConcurrentRequests:  10,
		ErrorPercentThreshold:  25,
		RequestVolumeThreshold: 1,
		SleepWindow:            0,
	})
}

func TestOK(t *testing.T) {
	runHystrixCommand(nil)

	s := statusChecker.CheckStatus("The thing")

	if len(s.StatusList) != 1 {
		t.Errorf("Length of StatusList should be 1, was %d", len(s.StatusList))
	}

	actual := s.StatusList[0]
	if actual.Result != healthchecks.OK {
		t.Errorf("Result shoud be `OK`, was `%s`", actual.Result)
	}
}

func TestOpenCircuit(t *testing.T) {
	// Open Circuit Breaker by simulating errors
	runHystrixCommand(errors.New("An error"))
	runHystrixCommand(errors.New("An error"))

	name := "The thing"
	s := statusChecker.CheckStatus(name)

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

	eDetails := "Circuit breaker is OPEN"
	if actual.Details != eDetails {
		t.Errorf("Details shoud be `%s`, was `%s`", eDetails, actual.Details)
	}
}

func runHystrixCommand(err error) (string, error) {
	resultChan := make(chan string, 1)
	errChan := hystrix.Go(COMMAND_NAME, func() error {
		if err != nil {
			return err
		}

		resultChan <- "done"
		return nil
	}, nil)

	// Wait for a small amount of time so CB can do it's processing
	// Without this, tests sometimes fail with improper results
	wait := time.Millisecond * 50

	// Block until we have a result or an error.
	select {
	case result := <-resultChan:
		time.Sleep(wait)
		return result, nil
	case err := <-errChan:
		time.Sleep(wait)
		return "", err
	}
}
