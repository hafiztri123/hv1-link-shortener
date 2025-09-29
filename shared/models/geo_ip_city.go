package models


type GeoIPCity struct {
	Country struct {
		ISOCode string `maxminddb:"iso_code"`
	} `maxinddb:"country"`

	City struct {
		Names map[string]string `maxminddb:"names"`
	} `maxminddb:"city"`
}