package model

import (
	"strings"

	core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
)

const (
	// LocalityLabel indicates the region/zone/subzone of an instance. It is used to override the native
	// registry's value.
	//
	// Note: because k8s labels does not support `/`, so we use `.` instead in k8s.
	LocalityLabel = "dapr-locality"
	// k8s dapr runtime locality label separator
	k8sSeparator = "."
)

// GetLocalityLabelOrDefault returns the locality from the supplied label, or falls back to
// the supplied default locality if the supplied label is empty. Because Kubernetes
// labels don't support `/`, we replace "." with "/" in the supplied label as a workaround.
func GetLocalityLabelOrDefault(label, defaultLabel string) string {
	if len(label) > 0 {
		// if there are /'s present we don't need to replace
		if strings.Contains(label, "/") {
			return label
		}
		// replace "." with "/"
		return strings.Replace(label, k8sSeparator, "/", -1)
	}
	return defaultLabel
}

// SplitLocalityLabel splits a locality label into region, zone and subzone strings.
func SplitLocalityLabel(locality string) (region, zone, subzone string) {
	items := strings.Split(locality, "/")
	switch len(items) {
	case 1:
		return items[0], "", ""
	case 2:
		return items[0], items[1], ""
	default:
		return items[0], items[1], items[2]
	}
}

// ConvertLocality converts '/' separated locality string to Locality struct.
func ConvertLocality(locality string) *core.Locality {
	if locality == "" {
		return &core.Locality{}
	}

	region, zone, subzone := SplitLocalityLabel(locality)
	return &core.Locality{
		Region:  region,
		Zone:    zone,
		SubZone: subzone,
	}
}
