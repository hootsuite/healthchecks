# go healthchecks

- [Introduction](#introduction)
- [How to Use It](#how-to-use-it)
- [Writing a StatusCheck](#writing-a-statuscheck)
- [Writing a TraverseCheck](#writing-a-traversecheck)
- [How To Contribute](#how-to-contribute)
- [License](#license)
- [Maintainers](#maintainers)

# Introduction
A go implementation of the [Health Checks API](https://github.com/hootsuite/health-checks-api) used for microservice
exploration, documentation and monitoring.

# How to Use It
Using the `healthchecks` framework in your service is easy.
- Define a `StatusEndpoint` for each dependency in your service.
- Register the `healthchecks` framework to respond to all `/status/...` requests passing a slice of all your `StatusEndpoint`s.
- That's it! As long as you have defined your `StatusEndpoint`s correctly, the framework will take care of the rest.

Example:
```
// Define a StatusEndpoint at '/status/db' for a database dependency
db := healthchecks.StatusEndpoint{
  Name: "The DB",
  Slug: "db",
  Type: "internal",
  IsTraversable: false,
  StatusCheck: sqlsc.SQLDBStatusChecker{
    DB: myDB
  },
  TraverseCheck: nil,
}

// Define a StatusEndpoint at '/status/service-organization' for the Organization service
org := healthchecks.StatusEndpoint{
  Name: "Organization Service",
  Slug: "service-organization",
  Type: "http",
  IsTraversable: true,
  StatusCheck: httpsc.HttpStatusChecker{
    BaseUrl: "[Read value from config]",
  },
  TraverseCheck: httpsc.HttpStatusChecker{
    BaseUrl: "[Read value from config]",
  },
}

// Define the list of StatusEndpoints for your service
statusEndpoints := []healthchecks.StatusEndpoint{ db, org }

// Set the path for the about and version files
aboutFilePath := "conf/about.json"
versionFilePath := "conf/version.txt"

// Set up any service injected customData for /status/about response.
// Values can be any valid JSON conversion and will override values set in about.json.
customData := make(map[string]interface{})
// Examples:
//
// String value
// customData["a-string"] = "some-value"
//
// Number value
// customData["a-number"] = 123
//
// Boolean value
// customData["a-bool"] = true
//
// Array
// customData["an-array"] = []string{"val1", "val2"}
//
// Custom object
// customObject := make(map[string]interface{})
// customObject["key1"] = 1
// customObject["key2"] = "some-value"
// customData["an-object"] = customObject

// Register all the "/status/..." requests to use our health checking framework
http.Handle("/status/", healthchecks.Handler(statusEndpoints, aboutFilePath, versionFilePath, customData))
```

# Writing a StatusCheck
A `StatusCheck` is a struct which implements the function `func CheckStatus(name string) StatusList`. A `StatusCheck` is defined or used in
a service but executed by the `healthchecks` framework. The key to a successful `StatusCheck` is to handle all errors on the
dependency you are checking. Below is an example of a `StatusCheck` that checks the connection of `Redis` using the
`gopkg.in/redis.v4` driver.

```
type RedisStatusChecker struct {
	client RedisClient
}

func (r RedisStatusChecker) CheckStatus(name string) healthchecks.StatusList {
	pong, err := r.client.Ping()

	// Set a default response
	s := healthchecks.Status{
		Description:  name,
		Result: healthchecks.OK,
		Details: "",
	}

	// Handle any errors that Ping() function returned
	if err != nil {
		s = healthchecks.Status{
			Description:  name,
			Result: healthchecks.CRITICAL,
			Details: err.Error(),
		}
	}

	// Make sure the pong response is what we expected
	if pong != "PONG" {
		s = healthchecks.Status{
			Description:  name,
			Result: healthchecks.CRITICAL,
			Details: fmt.Sprintf("Expecting `PONG` response, got `%s`", pong),
		}
	}

	// Return our response
	return healthchecks.StatusList{ StatusList: []healthchecks.Status{ s }}
}
```

# Writing a TraverseCheck
A `TraverseCheck` is a struct which implements the function `func Traverse(traversalPath []string, action string) (string, error)`.
A `TraverseCheck` is defined or used in a service but executed by the `healthchecks` framework. The key to a successful
`TraverseCheck` is to build and execute the `/status/traverse?action=[action]&dependencies=[dependencies]` request to
the service you are trying to traverse to and returning the response or error you got. Below is an example of a
`TraverseCheck` for an HTTP service.

```
type HttpStatusChecker struct {
	BaseUrl string
	Name    string
}

func (h HttpStatusChecker) Traverse(traversalPath []string, action string) (string, error) {
	dependencies := ""
	if len(traversalPath) > 0 {
		dependencies = fmt.Sprintf("&dependencies=%s", strings.Join(traversalPath, ","))
	}

	// Build our HTTP request
	url := fmt.Sprintf("%s/status/traverse?action=%s%s", h.BaseUrl, action, dependencies)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Printf("Error creating request: %s \n", err.Error())
		return "", err
	}

	// Execute HTTP request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error executing request: %s \n", err.Error())
		return "", err
	}

	// Defer the closing of the body
	defer resp.Body.Close()

	// Read our response
	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading response body: %s", err.Error())
		return "", err
	}

	// Return our response
	return string(responseBody), nil
}
```

# How To Contribute
Contribute by submitting a PR and a bug report in GitHub.

# License
healthchecks is released under the Apache License, Version 2.0. See [LICENSE](LICENSE) for details.

# Maintainers
- :octocat: [Adam Arsenault](https://github.com/HootAdam) - [@Adam_Arsenault](https://twitter.com/Adam_Arsenault)
- :octocat: [Mike Sample](https://github.com/michael-sample-hs) - [@mikesample](https://twitter.com/mikesample)
- :octocat: [Brandon McRae](https://github.com/brandon-mcrae-hs) - [@HootBrandon](https://twitter.com/HootBrandon)
- :octocat: [Denis Golovan](https://github.com/denis-golovan-hs) - [@dgolovan](https://twitter.com/dgolovan)