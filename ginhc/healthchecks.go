package ginhc

import (
	"github.com/gin-gonic/gin"
	"github.com/hootsuite/healthchecks"
)

func HealthChecksEndpoints(statusEndpoints []healthchecks.StatusEndpoint, aboutFilePath string, versionFilePath string, customData map[string]interface{}) gin.HandlerFunc {
	handler := healthchecks.Handler(statusEndpoints, aboutFilePath, versionFilePath, customData)
	return func(c *gin.Context) {
		handler.ServeHTTP(c.Writer, c.Request)
	}
}
