package volume

import (
	"testing"

	"k8s.io/apimachinery/pkg/util/sets"
)

func TestChooseZoneForVolume(t *testing.T) {
	cfg := diskConfig{
		Category: "ssd",
		Zones:    sets.NewString("zone-a", "zone-b", "zone-c"),
	}

	tests := []struct {
		name string
		zone string
	}{
		{"monday", "zone-c"},
		{"tuesday", "zone-a"},
		{"wednesday", "zone-b"},
		{"thursday", "zone-c"},
		{"friday", "zone-a"},
		{"saturday", "zone-a"},
		{"sunday", "zone-a"},
	}

	for _, tt := range tests {
		zone := cfg.ChooseZoneForVolume(tt.name)
		if zone != tt.zone {
			t.Errorf("ChooseZoneForVolume(%q) = %s; want %s", tt.name, zone, tt.zone)
		}
	}
}
