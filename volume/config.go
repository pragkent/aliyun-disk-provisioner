package volume

import (
	"fmt"
	"strings"

	"github.com/kubernetes-incubator/external-storage/lib/controller"
	"github.com/pragkent/aliyun-disk-provisioner/cloud"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/kubernetes/pkg/volume"
)

type diskConfig struct {
	Category string
	Zones    sets.String
}

func (c *diskConfig) ChooseZoneForVolume(name string) string {
	return volume.ChooseZoneForVolume(c.Zones, name)
}

func parseDiskConfig(options controller.VolumeOptions, provider cloud.Provider) (*diskConfig, error) {
	var category, zone, zones string

	for k, v := range options.Parameters {
		switch strings.ToLower(k) {
		case "category":
			category = v
		case "zone":
			zone = v
		case "zones":
			zones = v
		default:
			return nil, fmt.Errorf("invalid StorageClass option %q", k)
		}
	}

	zoneSet, err := parseZones(zone, zones, provider)
	if err != nil {
		return nil, err
	}

	cfg := &diskConfig{
		Category: category,
		Zones:    zoneSet,
	}
	return cfg, nil
}

func parseZones(zone string, zones string, provider cloud.Provider) (sets.String, error) {
	if len(zone) != 0 && len(zones) != 0 {
		return nil, fmt.Errorf("both zone and zones StorageClass parameters muster not be used at the same time")
	}

	if len(zones) != 0 {
		zoneSet, err := ZonesToSet(zones)
		if err != nil {
			return nil, err
		}

		return zoneSet, nil
	}

	if len(zone) != 0 {
		if err := ValidateZone(zone); err != nil {
			return nil, err
		}

		zoneSet := sets.NewString(zone)
		return zoneSet, nil
	}

	allZones, err := provider.Zones()
	if err != nil {
		return nil, fmt.Errorf("provider.Zones error: %v", err)
	}

	return sets.NewString(allZones...), nil
}
