package sqlsc

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/hootsuite/healthchecks"
	"github.com/DATA-DOG/go-sqlmock"
	"testing"
)

func TestDatabaseOK(t *testing.T) {

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}

	mock.ExpectExec("SELECT 1").WillReturnResult(sqlmock.NewResult(0, 0))

	checker := SQLDBStatusChecker{DB: db}
	statusList := checker.CheckStatus(expectedOKResponse.Description).StatusList
	status := statusList[0]

	assert.True(t, len(statusList) == 1, "Expected length of statusList to be 1. Got %v", len(statusList))
	assert.Equal(t, expectedOKResponse, status)
}

func TestDatabaseCRIT(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}

	mockedSQLError := errors.New("Mocked connection error occurred.")
	mock.ExpectExec("SELECT 1").WillReturnError(mockedSQLError)

	expectedCRITResult := createExpectedCRITResponse(mockedSQLError)

	checker := SQLDBStatusChecker{DB: db}
	statusList := checker.CheckStatus(expectedCRITResult.Description).StatusList
	status := statusList[0]

	assert.True(t, len(statusList) == 1, "Expected length of statusList to be 1. Got %v", len(statusList))
	assert.Equal(t, expectedCRITResult, status)
}

var expectedOKResponse = healthchecks.Status{
	Description: "Mysql Test Database",
	Result:      healthchecks.OK,
	Details:     "",
}

func createExpectedCRITResponse(err error) healthchecks.Status {
	desc := "Mysql Test Database"
	return healthchecks.Status{
		Description: desc,
		Result:      healthchecks.CRITICAL,
		Details:     fmt.Sprintf("%v check failed: %v", desc, err),
	}
}
