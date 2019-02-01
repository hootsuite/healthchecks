package healthchecks

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"sync"
	"time"
)

const (
	ABOUT_FIELD_NA      string = "N/A"
	ABOUT_PROTOCOL_HTTP string = "http"
	VERSION_NA          string = "N/A"
)

type ConfigAbout struct {
	Id          string                 `json:"id"`
	Summary     string                 `json:"summary"`
	Description string                 `json:"description"`
	Maintainers []string               `json:"maintainers"`
	ProjectRepo string                 `json:"projectRepo"`
	ProjectHome string                 `json:"projectHome"`
	LogsLinks   []string               `json:"logsLinks"`
	StatsLinks  []string               `json:"statsLinks"`
	CustomData  map[string]interface{} `json:"customData"`
}

type AboutResponse struct {
	Id           string                 `json:"id"`
	Name         string                 `json:"name"`
	Description  string                 `json:"description"`
	Protocol     string                 `json:"protocol"`
	Owners       []string               `json:"owners"`
	Version      string                 `json:"version"`
	Host         string                 `json:"host"`
	ProjectRepo  string                 `json:"projectRepo"`
	ProjectHome  string                 `json:"projectHome"`
	LogsLinks    []string               `json:"logsLinks"`
	StatsLinks   []string               `json:"statsLinks"`
	Dependencies []Dependency           `json:"dependencies"`
	CustomData   map[string]interface{} `json:"customData"`
}

type AboutResponseV2 struct {
	Id           string                 `json:"id"`
	Name         string                 `json:"name"`
	Description  string                 `json:"description"`
	Protocol     string                 `json:"protocol"`
	Owners       []string               `json:"owners"`
	Version      string                 `json:"version"`
	Host         string                 `json:"host"`
	ProjectRepo  string                 `json:"projectRepo"`
	ProjectHome  string                 `json:"projectHome"`
	LogsLinks    []string               `json:"logsLinks"`
	StatsLinks   []string               `json:"statsLinks"`
	Dependencies []DependencyInfo       `json:"dependencies"`
	CustomData   map[string]interface{} `json:"customData"`
}

type Dependency struct {
	Name           string         `json:"name"`
	Status         []JsonResponse `json:"status"`
	StatusDuration float64        `json:"statusDuration"`
	StatusPath     string         `json:"statusPath"`
	Type           string         `json:"type"`
	IsTraversable  bool           `json:"isTraversable"`
}

type DependencyInfo struct {
	Name           string       `json:"name"`
	Status         JsonResponse `json:"status"`
	StatusDuration float64      `json:"statusDuration"`
	StatusPath     string       `json:"statusPath"`
	Type           string       `json:"type"`
	IsTraversable  bool         `json:"isTraversable"`
}

type dependencyPosition struct {
	item     Dependency
	position int
}

type dependencyInfoPosition struct {
	item     DependencyInfo
	position int
}

func getAboutFieldValue(aboutConfigMap map[string]interface{}, key string, aboutFilePath string) string {
	value, ok := aboutConfigMap[key]
	if !ok {
		fmt.Printf("Field `%s` missing from %s.\n", key, aboutFilePath)
		return ABOUT_FIELD_NA

	}

	stringValue, ok := value.(string)
	if !ok {
		fmt.Printf("Field `%s` is not a String in %s.\n", key, aboutFilePath)
		return ABOUT_FIELD_NA

	}

	return stringValue
}

func getAboutFieldValues(aboutConfigMap map[string]interface{}, key string, aboutFilePath string) []string {
	value, ok := aboutConfigMap[key]
	if !ok {
		fmt.Printf("Field `%s` missing from %s.\n", key, aboutFilePath)
		return []string{}
	}

	interfaces, ok := value.([]interface{})
	if !ok {
		fmt.Printf("Field `%s` is not an array in %s.\n", key, aboutFilePath)
		return []string{}
	}

	strings := make([]string, len(interfaces))
	for i := range interfaces {
		stringValue, ok := interfaces[i].(string)
		if !ok {
			strings[i] = ABOUT_FIELD_NA
			fmt.Printf("Field[%d] `%s` is not a String in %s.\n", i, key, aboutFilePath)
		} else {
			strings[i] = stringValue
		}
	}

	return strings
}

func getAboutCustomDataFieldValues(aboutConfigMap map[string]interface{}, aboutFilePath string) map[string]interface{} {
	value, ok := aboutConfigMap["customData"]
	if !ok {
		return nil
	}

	mapValue, ok := value.(map[string]interface{})
	if !ok {
		fmt.Printf("Field `customData` is not a valid JSON object in %s.\n", aboutFilePath)
		return nil
	}

	return mapValue
}

func About(
	statusEndpoints []StatusEndpoint,
	protocol string, aboutFilePath string,
	versionFilePath string,
	customData map[string]interface{},
	apiVersion int,
	checkStatus bool,
) string {
	switch apiVersion {
	case APIV1:
		return aboutV1(statusEndpoints, protocol, aboutFilePath, versionFilePath, customData)
	case APIV2:
		return aboutV2(statusEndpoints, protocol, aboutFilePath, versionFilePath, customData, checkStatus)
	}
	// should never reach here
	return ""
}

func aboutV1(
	statusEndpoints []StatusEndpoint,
	protocol string, aboutFilePath string,
	versionFilePath string,
	customData map[string]interface{},
) string {
	aboutData, _ := ioutil.ReadFile(aboutFilePath)

	// Initialize ConfigAbout with default values in case we have problems reading from the file
	aboutConfig := ConfigAbout{
		Id:          ABOUT_FIELD_NA,
		Summary:     ABOUT_FIELD_NA,
		Description: ABOUT_FIELD_NA,
		Maintainers: []string{},
		ProjectRepo: ABOUT_FIELD_NA,
		ProjectHome: ABOUT_FIELD_NA,
		LogsLinks:   []string{},
		StatsLinks:  []string{},
	}

	// Unmarshal JSON into a generic object so we don't completely fail if one of the fields is invalid or missing
	var aboutConfigMap map[string]interface{}
	err := json.Unmarshal(aboutData, &aboutConfigMap)

	if err == nil {
		// Parse out each value individually
		aboutConfig.Id = getAboutFieldValue(aboutConfigMap, "id", aboutFilePath)
		aboutConfig.Summary = getAboutFieldValue(aboutConfigMap, "summary", aboutFilePath)
		aboutConfig.Description = getAboutFieldValue(aboutConfigMap, "description", aboutFilePath)
		aboutConfig.Maintainers = getAboutFieldValues(aboutConfigMap, "maintainers", aboutFilePath)
		aboutConfig.ProjectRepo = getAboutFieldValue(aboutConfigMap, "projectRepo", aboutFilePath)
		aboutConfig.ProjectHome = getAboutFieldValue(aboutConfigMap, "projectHome", aboutFilePath)
		aboutConfig.LogsLinks = getAboutFieldValues(aboutConfigMap, "logsLinks", aboutFilePath)
		aboutConfig.StatsLinks = getAboutFieldValues(aboutConfigMap, "statsLinks", aboutFilePath)
		aboutConfig.CustomData = getAboutCustomDataFieldValues(aboutConfigMap, aboutFilePath)
	} else {
		fmt.Printf("Error deserializing about data from %s. Error: %s JSON: %s\n", aboutFilePath, err.Error(), aboutData)
	}

	// Merge custom data from about.json with custom data passed in by client
	// and prefer values passed by client over values in about.json
	if customData != nil {
		if aboutConfig.CustomData == nil {
			aboutConfig.CustomData = make(map[string]interface{})
		}

		for key, value := range customData {
			aboutConfig.CustomData[key] = value
		}
	}

	// Extract version
	var version string
	versionData, err := ioutil.ReadFile(versionFilePath)
	if err != nil {
		fmt.Printf("Error reading version from %s. Error: %s\n", versionFilePath, err.Error())
		version = VERSION_NA
	} else {
		version = strings.TrimSpace(string(versionData))
	}

	// Get hostname
	host, err := os.Hostname()
	if err != nil {
		fmt.Printf("Error getting hostname. Error: %s\n", err.Error())
		host = "unknown"
	}

	aboutResponse := AboutResponse{
		Id:          aboutConfig.Id,
		Name:        aboutConfig.Summary,
		Description: aboutConfig.Description,
		Protocol:    protocol,
		Owners:      aboutConfig.Maintainers,
		Version:     version,
		Host:        host,
		ProjectRepo: aboutConfig.ProjectRepo,
		ProjectHome: aboutConfig.ProjectHome,
		LogsLinks:   aboutConfig.LogsLinks,
		StatsLinks:  aboutConfig.StatsLinks,
		CustomData:  aboutConfig.CustomData,
	}

	// Execute status checks async
	var wg sync.WaitGroup
	dc := make(chan dependencyPosition)
	wg.Add(len(statusEndpoints))

	for ie, se := range statusEndpoints {
		go func(s StatusEndpoint, i int) {
			start := time.Now()
			dependencyStatus := translateStatusList(s.StatusCheck.CheckStatus(s.Name))
			elapsed := float64(time.Since(start)) * 0.000000001
			dependency := Dependency{
				Name:           s.Name,
				Status:         dependencyStatus,
				StatusDuration: elapsed,
				StatusPath:     s.Slug,
				Type:           s.Type,
				IsTraversable:  s.IsTraversable,
			}

			dc <- dependencyPosition{
				item:     dependency,
				position: i,
			}
		}(se, ie)
	}

	// Collect our responses and put them in the right spot
	dependencies := make([]Dependency, len(statusEndpoints))
	go func() {
		for dp := range dc {
			dependencies[dp.position] = dp.item
			wg.Done()
		}
	}()

	// Wait until all async status checks are done and collected
	wg.Wait()
	close(dc)

	aboutResponse.Dependencies = dependencies

	aboutResponseJSON, err := json.Marshal(aboutResponse)
	if err != nil {
		msg := fmt.Sprintf("Error serializing AboutResponse: %s", err)
		sl := StatusList{
			StatusList: []Status{
				{Description: "Invalid AboutResponse", Result: CRITICAL, Details: msg},
			},
		}
		return SerializeStatusList(sl, APIV1)
	}

	return string(aboutResponseJSON)
}

func aboutV2(
	statusEndpoints []StatusEndpoint,
	protocol string, aboutFilePath string,
	versionFilePath string,
	customData map[string]interface{},
	checkStatus bool,
) string {
	aboutData, _ := ioutil.ReadFile(aboutFilePath)

	// Initialize ConfigAbout with default values in case we have problems reading from the file
	aboutConfig := ConfigAbout{
		Id:          ABOUT_FIELD_NA,
		Summary:     ABOUT_FIELD_NA,
		Description: ABOUT_FIELD_NA,
		Maintainers: []string{},
		ProjectRepo: ABOUT_FIELD_NA,
		ProjectHome: ABOUT_FIELD_NA,
		LogsLinks:   []string{},
		StatsLinks:  []string{},
	}

	// Unmarshal JSON into a generic object so we don't completely fail if one of the fields is invalid or missing
	var aboutConfigMap map[string]interface{}
	err := json.Unmarshal(aboutData, &aboutConfigMap)

	if err == nil {
		// Parse out each value individually
		aboutConfig.Id = getAboutFieldValue(aboutConfigMap, "id", aboutFilePath)
		aboutConfig.Summary = getAboutFieldValue(aboutConfigMap, "summary", aboutFilePath)
		aboutConfig.Description = getAboutFieldValue(aboutConfigMap, "description", aboutFilePath)
		aboutConfig.Maintainers = getAboutFieldValues(aboutConfigMap, "maintainers", aboutFilePath)
		aboutConfig.ProjectRepo = getAboutFieldValue(aboutConfigMap, "projectRepo", aboutFilePath)
		aboutConfig.ProjectHome = getAboutFieldValue(aboutConfigMap, "projectHome", aboutFilePath)
		aboutConfig.LogsLinks = getAboutFieldValues(aboutConfigMap, "logsLinks", aboutFilePath)
		aboutConfig.StatsLinks = getAboutFieldValues(aboutConfigMap, "statsLinks", aboutFilePath)
		aboutConfig.CustomData = getAboutCustomDataFieldValues(aboutConfigMap, aboutFilePath)
	} else {
		fmt.Printf("Error deserializing about data from %s. Error: %s JSON: %s\n", aboutFilePath, err.Error(), aboutData)
	}

	// Merge custom data from about.json with custom data passed in by client
	// and prefer values passed by client over values in about.json
	if customData != nil {
		if aboutConfig.CustomData == nil {
			aboutConfig.CustomData = make(map[string]interface{})
		}

		for key, value := range customData {
			aboutConfig.CustomData[key] = value
		}
	}

	// Extract version
	var version string
	versionData, err := ioutil.ReadFile(versionFilePath)
	if err != nil {
		fmt.Printf("Error reading version from %s. Error: %s\n", versionFilePath, err.Error())
		version = VERSION_NA
	} else {
		version = strings.TrimSpace(string(versionData))
	}

	// Get hostname
	host, err := os.Hostname()
	if err != nil {
		fmt.Printf("Error getting hostname. Error: %s\n", err.Error())
		host = "unknown"
	}

	aboutResponse := AboutResponseV2{
		Id:          aboutConfig.Id,
		Name:        aboutConfig.Summary,
		Description: aboutConfig.Description,
		Protocol:    protocol,
		Owners:      aboutConfig.Maintainers,
		Version:     version,
		Host:        host,
		ProjectRepo: aboutConfig.ProjectRepo,
		ProjectHome: aboutConfig.ProjectHome,
		LogsLinks:   aboutConfig.LogsLinks,
		StatsLinks:  aboutConfig.StatsLinks,
		CustomData:  aboutConfig.CustomData,
	}

	dependencies := make([]DependencyInfo, len(statusEndpoints))
	if checkStatus {
		// Execute status checks async
		var wg sync.WaitGroup
		dc := make(chan dependencyInfoPosition)
		wg.Add(len(statusEndpoints))

		for ie, se := range statusEndpoints {
			go func(s StatusEndpoint, i int) {
				start := time.Now()
				dependencyStatus := translateStatusListV2(s.StatusCheck.CheckStatus(s.Name))
				elapsed := float64(time.Since(start)) * 0.000000001
				dependency := DependencyInfo{
					Name:           s.Name,
					Status:         dependencyStatus,
					StatusDuration: elapsed,
					StatusPath:     s.Slug,
					Type:           s.Type,
					IsTraversable:  s.IsTraversable,
				}

				dc <- dependencyInfoPosition{
					item:     dependency,
					position: i,
				}
			}(se, ie)
		}

		// Collect our responses and put them in the right spot
		go func() {
			for dp := range dc {
				dependencies[dp.position] = dp.item
				wg.Done()
			}
		}()

		// Wait until all async status checks are done and collected
		wg.Wait()
		close(dc)
	} else {
		for _, statusEndpoint := range statusEndpoints {
			dependencies = append(dependencies, DependencyInfo{
				Name:          statusEndpoint.Name,
				StatusPath:    statusEndpoint.Slug,
				Type:          statusEndpoint.Type,
				IsTraversable: statusEndpoint.IsTraversable,
			})
		}
	}

	aboutResponse.Dependencies = dependencies

	aboutResponseJSON, err := json.Marshal(aboutResponse)
	if err != nil {
		msg := fmt.Sprintf("Error serializing AboutResponse: %s", err)
		sl := StatusList{
			StatusList: []Status{
				{Description: "Invalid AboutResponse", Result: CRITICAL, Details: msg},
			},
		}
		return SerializeStatusList(sl, APIV2)
	}

	return string(aboutResponseJSON)
}
