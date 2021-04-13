// Package weatherCaster eases getting and parsing weather data from OpenWeatherMap.org
package weatherCaster

import (
	"fmt"
	"net/url"
)

// An OpenWeatherMap.org client
type Client struct {
	ApiKey, ApiHost, ApiUri, ApiQueryOptions string
}

// Return a pointer to a new weather client,
// while setting defaults.
func NewClient(apiKey string) *Client {
	var c Client

	c.ApiKey = apiKey
	// Set other OpenWeatherMap.org defaults
	c.ApiHost = "api.openweathermap.org"
	c.ApiUri = "/data/2.5/forecast"
	// Additional API query options to always include, for now.
	// Some of these will eventually become configurable.
	// The `cnt` limits how many results are returned in a "circle" around a city.
	c.ApiQueryOptions = "&units=imperial&cnt=1"

	return &c
}

// Form a URL to the weather API,
// given a city to query.
func (c Client) FormUrl(city string) string {
	u := fmt.Sprintf("https://%s%s/?q=%s&appid=%s%s", c.ApiHost, c.ApiUri, url.QueryEscape(city), c.ApiKey, c.ApiQueryOptions)
	return u
}
