package sqlsc

import (
	"database/sql"
	"fmt"
	"github.com/hootsuite/healthchecks"
)

type SQLDBStatusChecker struct {
	DB *sql.DB
}

func (d SQLDBStatusChecker) CheckStatus(name string) healthchecks.StatusList {

	_, err := d.DB.Exec("SELECT 1")

	var result healthchecks.Status
	if err != nil {
		result = healthchecks.Status{
			Description: name,
			Result:      healthchecks.CRITICAL,
			Details:     fmt.Sprintf("%v check failed: %v", name, err),
		}
	} else {
		result = healthchecks.Status{
			Description: name,
			Result:      healthchecks.OK,
			Details:     "",
		}
	}
	return healthchecks.StatusList{StatusList: []healthchecks.Status{result}}
}
