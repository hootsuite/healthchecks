package healthchecks

// APIVersion is the version of the API
type APIVersion int

const (
	// APIV1 is enum for V1 of the healthchecks API
	APIV1 = iota
	// APIV2 is enum for V2 of the healthchecks API
	APIV2
)
