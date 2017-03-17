package healthchecks

import (
	"fmt"
	"io"
	"net/http"
	"strings"
)

// Handler returns a http.Handler that responds to status check requests. It should be registered at `/status/...`
func Handler(statusEndpoints []StatusEndpoint, aboutFilePath string, versionFilePath string, customData map[string]interface{}) http.Handler {
	return HandlerFunc(statusEndpoints, aboutFilePath, versionFilePath, customData)
}

// HandlerFunc returns a http.HandlerFunc that responds to status check requests. It should be registered at `/status/...`
func HandlerFunc(statusEndpoints []StatusEndpoint, aboutFilePath string, versionFilePath string, customData map[string]interface{}) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		slug := strings.Split(r.URL.Path, "/")
		endpoint := slug[2]

		switch endpoint {
		case "about":
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			io.WriteString(w, About(statusEndpoints, ABOUT_PROTOCOL_HTTP, aboutFilePath, versionFilePath, customData))
		case "aggregate":
			typeFilter := r.URL.Query().Get("type")
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			io.WriteString(w, Aggregate(statusEndpoints, typeFilter))
		case "am-i-up":
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			io.WriteString(w, "OK")
		case "traverse":
			action := r.URL.Query().Get("action")
			if action == "" {
				action = "about"
			}
			dependencies := []string{}
			queryDependencies := r.URL.Query().Get("dependencies")
			if queryDependencies != "" {
				dependencies = strings.Split(queryDependencies, ",")
			}

			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			io.WriteString(w, Traverse(statusEndpoints, dependencies, action, ABOUT_PROTOCOL_HTTP, aboutFilePath, versionFilePath, customData))
		default:
			endpoint := FindStatusEndpoint(statusEndpoints, endpoint)
			if endpoint == nil {
				w.Header().Set("Content-Type", "application/json; charset=utf-8")
				w.WriteHeader(http.StatusNotFound)
				io.WriteString(w, SerializeStatusList(StatusList{
					StatusList: []Status{
						{
							Description: "Unknow Status endpoint",
							Result:      CRITICAL,
							Details:     fmt.Sprintf("Status endpoint does not exist: %s", r.URL.Path),
						},
					},
				}))
				return
			}

			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			io.WriteString(w, ExecuteStatusCheck(endpoint))
		}
	})
}
