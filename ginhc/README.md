# gin healthchecks

- [Introduction](#introduction)
- [How to Use It](#how-to-use-it)
- [How To Contribute](#how-to-contribute)
- [License](#license)

# Introduction
A [gin](https://github.com/gin-gonic/gin) implementation of the [Health Checks API](https://github.com/hootsuite/health-checks-api) used for microservice
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
app.GET("/status/:slug", ginhc.HealthChecksEndpoints(statusEndpoints, aboutFilePath, versionFilePath, customData))
```

# How To Contribute
Contribute by submitting a PR and a bug report in GitHub.

# License
healthchecks is released under the Apache License, Version 2.0. See [LICENSE](LICENSE) for details.