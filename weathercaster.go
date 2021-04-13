// Package weatherCaster eases getting and parsing weather data from OpenWeatherMap.org
package weatherCaster

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

// Multiple structs representing nested json from the OpenWeatherMap.org API,
// which match output from its `/2.5/forecast` URI.
// These are defined inside-out, starting with the most inner data.
type ApiResponseListWeather struct {
	Description string
}
type ApiResponseList struct {
	Weather []ApiResponseListWeather
}
type ApiResponse struct {
	List []ApiResponseList
}

// An OpenWeatherMap.org client
type Client struct {
	ApiKey, ApiHost, ApiUri, ApiQueryOptions string
	response                                 ApiResponse
}

// Return a pointer to a new weather client.
func NewClient(apiKey string) *Client {
	var c Client

	c.ApiKey = apiKey
	// Set other OpenWeatherMap.org defaults
	c.ApiHost = "api.openweathermap.org"
	c.ApiUri = "/data/2.5/forecast"
	// Additional API query options to always include, for now.
	// Some of these will eventually become configurable.
	// The `cnt` limits how many time-stamps are returned.
	// ref: https://openweathermap.org/forecast5#limit
	c.ApiQueryOptions = "&units=imperial&cnt=1"

	return &c
}

// Given a city to query,
// return a weather API URL.
func (c Client) FormUrl(city string) string {
	u := fmt.Sprintf("https://%s%s/?q=%s&appid=%s%s", c.ApiHost, c.ApiUri, url.QueryEscape(city), c.ApiKey, c.ApiQueryOptions)
	return u
}

// Send an HTTP GET request.
func (c Client) HttpGet(url string) (string, error) {
	httpClient := http.Client{}
	res, err := httpClient.Get(url)
	if err != nil {
		return "", err
	}

	if res.Body != nil {
		defer res.Body.Close()
	}

	// ioutil.ReadAll() returns a slice of bytes
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", err
	}

	if res.StatusCode >= 400 {
		return "", fmt.Errorf("HTTP %d returned from weather API: %v\n", res.StatusCode, string(body))
	}

	return string(body), nil
}

// Parse JSON returned from the weather API,
// storing the response in the weather client.
func (c *Client) ParseJson(j string) error {
	jsonBytes := []byte(j)
	err := json.Unmarshal(jsonBytes, &c.response)
	if err != nil {
		return err
	}

	return nil
}

// Return the weather description,
// from the last query to the weather API.
func (c *Client) GetDescription() string {
	return c.response.List[0].Weather[0].Description
}

// Return the response from the last query to the weather API.
func (c Client) GetApiResponse() ApiResponse {
	return c.response
}
