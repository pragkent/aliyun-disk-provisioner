package volume

import (
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/util/sets"
)

// ZonesToSet converts a string containing a comma separated list of zones to set
func ZonesToSet(zonesString string) (sets.String, error) {
	zonesSlice := strings.Split(zonesString, ",")
	zonesSet := make(sets.String)
	for _, zone := range zonesSlice {
		trimmedZone := strings.TrimSpace(zone)
		if trimmedZone == "" {
			return make(sets.String), fmt.Errorf("comma separated list of zones (%q) must not contain an empty zone", zonesString)
		}
		zonesSet.Insert(trimmedZone)
	}
	return zonesSet, nil
}

// ValidateZone returns:
// - an error in case zone is an empty string or contains only any combination of spaces and tab characters
// - nil otherwise
func ValidateZone(zone string) error {
	if strings.TrimSpace(zone) == "" {
		return fmt.Errorf("the provided %q zone is not valid, it's an empty string or contains only spaces and tab characters", zone)
	}
	return nil
}
