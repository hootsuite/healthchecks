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

		apiVersion := APIV1
		if slug[2] == "v2" {
			apiVersion = APIV2
		}

		switch apiVersion {
		case APIV1:
			handleV1Api(w, r, slug[2], statusEndpoints, aboutFilePath, versionFilePath, customData)
		case APIV2:
			handleV2Api(w, r, slug[3], statusEndpoints, aboutFilePath, versionFilePath, customData)
		}
	})
}

func handleV1Api(
	w http.ResponseWriter,
	r *http.Request,
	endpoint string,
	statusEndpoints []StatusEndpoint,
	aboutFilePath string,
	versionFilePath string,
	customData map[string]interface{},
) {
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
	case "v2/am-i-up":
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		io.WriteString(
			w,
			SerializeStatusList(
				StatusList{
					StatusList: []Status{
						Status{
							Description: "Am I Up",
							Result:      OK,
							Details:     "The service is running",
						},
					},
				},
				APIV2,
			),
		)
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
			}, APIV1))
			return
		}

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		io.WriteString(w, ExecuteStatusCheck(endpoint, APIV1))
	}
}

func handleV2Api(
	w http.ResponseWriter,
	r *http.Request,
	endpoint string,
	statusEndpoints []StatusEndpoint,
	aboutFilePath string,
	versionFilePath string,
	customData map[string]interface{},
) {
	switch endpoint {
	case "about":
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		io.WriteString(w, About(statusEndpoints, ABOUT_PROTOCOL_HTTP, aboutFilePath, versionFilePath, customData))
	case "aggregate":
		typeFilter := r.URL.Query().Get("type")
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		io.WriteString(w, Aggregate(statusEndpoints, typeFilter))
	case "am-i-up":
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		io.WriteString(
			w,
			SerializeStatusList(
				StatusList{
					StatusList: []Status{
						Status{
							Description: "Am I Up",
							Result:      OK,
							Details:     "The service is running",
						},
					},
				},
				APIV2,
			),
		)
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
			}, APIV2))
			return
		}

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		io.WriteString(w, ExecuteStatusCheck(endpoint, APIV2))
	}
}
