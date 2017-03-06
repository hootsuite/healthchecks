package ginhc

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/hootsuite/healthchecks"
	"net/http"
	"strings"
)

func HealthChecksEndpoints(statusEndpoints []healthchecks.StatusEndpoint, aboutFilePath string, versionFilePath string, customData map[string]interface{}) gin.HandlerFunc {
	// Use a closure here so we can pass the registered status endpoints
	// into the handler function.
	return func(c *gin.Context) {

		// Extract the path after /status/...
		slug := strings.Split(c.Request.URL.Path, "/")
		endpoint := slug[2]

		switch endpoint {
		case "about":
			c.Header("Content-Type", "application/json; charset=utf-8")
			c.String(http.StatusOK, healthchecks.About(statusEndpoints, healthchecks.ABOUT_PROTOCOL_HTTP, aboutFilePath, versionFilePath, customData))
		case "aggregate":
			var typeFilter string = ""
			queryType := c.Request.URL.Query()["type"]
			if queryType != nil && len(queryType) > 0 && len(queryType[0]) > 0 {
				typeFilter = queryType[0]
			}

			c.Header("Content-Type", "application/json; charset=utf-8")
			c.String(http.StatusOK, healthchecks.Aggregate(statusEndpoints, typeFilter))
		case "am-i-up":
			response := "OK"
			c.String(http.StatusOK, response)
		case "traverse":
			queryParams := c.Request.URL.Query()

			action := "about"
			queryAction := queryParams["action"]
			if queryAction != nil && len(queryAction) > 0 && len(queryAction[0]) > 0 {
				action = queryAction[0]
			}

			dependencies := []string{}
			queryDependencies := queryParams["dependencies"]
			if queryDependencies != nil && len(queryDependencies) > 0 && len(queryDependencies[0]) > 0 {
				commaDependencies := queryDependencies[0]
				dependencies = strings.Split(commaDependencies, ",")
			}

			c.Header("Content-Type", "application/json; charset=utf-8")
			c.String(http.StatusOK, healthchecks.Traverse(statusEndpoints, dependencies, action, healthchecks.ABOUT_PROTOCOL_HTTP, aboutFilePath, versionFilePath, customData))
		default:
			statusEndpoint := healthchecks.FindStatusEndpoint(statusEndpoints, endpoint)
			if statusEndpoint != nil {
				c.Header("Content-Type", "application/json; charset=utf-8")
				c.String(http.StatusOK, healthchecks.ExecuteStatusCheck(statusEndpoint))
				return
			}

			notFoundResponse := healthchecks.StatusList{
				StatusList: []healthchecks.Status{
					{
						Description: "Unknow Status endpoint",
						Result:      healthchecks.CRITICAL,
						Details:     fmt.Sprintf("Status endpoint does not exist: %s", c.Request.URL.Path),
					},
				},
			}
			c.Header("Content-Type", "application/json; charset=utf-8")
			c.String(http.StatusNotFound, healthchecks.SerializeStatusList(notFoundResponse))
		}
	}
}
