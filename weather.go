// Package weather eases getting and parsing weather data from OpenWeatherMap.org
package weather

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
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

// NewClient returns a pointer to a new weather client.
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

// formAPIUrl accepts a city and returns an OpenWeatherMap.org URL.
func (c Client) formAPIUrl(city string) string {
	// This will eventually vary the URL using client configuration,
	// E.G. temperature units.
	u := fmt.Sprintf("https://%s%s/?q=%s&appid=%s%s", c.ApiHost, c.ApiUri, url.QueryEscape(city), c.ApiKey, c.ApiQueryOptions)
	return u
}

// queryAPI accepts an OpenWeatherMap.org URL and queries its API.
func (c Client) queryAPI(url string) (ApiResponse, error) {
	var apiRes ApiResponse

	// This non-default client and its timeout is used
	// RE: https://medium.com/@nate510/don-t-use-go-s-default-http-client-4804cb19f779
	httpClient := http.Client{Timeout: time.Second * 3}
	httpRes, err := httpClient.Get(url)
	if err != nil {
		return apiRes, err
	}

	defer httpRes.Body.Close()

	// ioutil.ReadAll() returns a slice of bytes
	body, err := ioutil.ReadAll(httpRes.Body)
	if err != nil {
		return apiRes, err
	}

	if httpRes.StatusCode >= 400 {
		return apiRes, fmt.Errorf("HTTP %d returned from weather API: %v", httpRes.StatusCode, string(body))
	}

	jsonBytes := []byte(string(body))
	err = json.Unmarshal(jsonBytes, &apiRes)
	if err != nil {
		return apiRes, err
	}
	return apiRes, nil
}

// ForecastByCity queries the weather API for a `city,state,country-code`,
// and stores the result in the Client object.
func (c *Client) ForecastByCity(city string) (string, error) {
	res, err := c.queryAPI(c.formAPIUrl(city))
	if err != nil {
		return "", fmt.Errorf("Error querying weather API for city %q: %v", city, err)
	}

	c.response = res

	// The GetForecast method returns its own error.
	return c.GetForecast()
}

// GetForecasts returns formatted forecast output,
// using the last weather API result.
func (c *Client) GetForecast() (string, error) {
	if len(c.response.List) == 0 {
		return "", fmt.Errorf("GetForecast() has an empty response.List")
	}

	if len(c.response.List[0].Weather) == 0 {
		return "", fmt.Errorf("GetForecast() has an empty response.List[0].Weather")
	}

	return c.response.List[0].Weather[0].Description, nil
}

// GetApiResponse returns the response from the last query to the weather API.
func (c Client) GetApiResponse() ApiResponse {
	return c.response
}
