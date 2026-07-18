package growiota

import "os"

// IsActivated returns true if the Growiota Premium module is activated.
func IsActivated() bool {
	return os.Getenv("GROWIOTA_PREMIUM") == "1"
}
