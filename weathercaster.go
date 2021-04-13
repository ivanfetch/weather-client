// Package weatherCaster eases getting and parsing weather data from an API
package weatherCaster

import (
	"fmt"
	"net/url"
)

// Constants that will later become proper options, passed in from CLI.
const ApiHost = "api.openweathermap.org"
const ApiUri = "/data/2.5/forecast"

// Additional API query options to always include, for now.
// Some of these will eventually become configurable.
// The `cnt` limits how many results are returned in a "circle" around a city.
const ApiQueryOptions = "&units=imperial&cnt=1"

// Form a URL to the weather API,
// given an API key and city to query.
func FormUrl(apiKey, city string) string {
	u := fmt.Sprintf("https://%s%s/?q=%s&appid=%s%s", ApiHost, ApiUri, url.QueryEscape(city), apiKey, ApiQueryOptions)
	return u
}
