// Package weather eases getting and parsing weather data from OpenWeatherMap.org
package weather

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// APIResponse matches fields from the OpenWeatherMap.org API `/2.5/forecast`.
// This does not fully mirror the API!
type APIResponse struct {
	List []struct {
		Weather []struct {
			Description string
		}
		Main struct {
			Temp       float64
			Feels_like float64
			Humidity   float64
		}
	}
}

// An OpenWeatherMap.org client
type Client struct {
	APIKey, APIHost, APIUri string
}

// NewClient returns a pointer to a new weather client.
func NewClient(APIKey string) *Client {
	return &Client{
		APIKey:  APIKey,
		APIHost: "api.openweathermap.org",
		APIUri:  "/data/2.5/forecast",
	}
}

// formAPIUrl accepts a city and temperature units, and returns an OpenWeatherMap.org URL.
func (c Client) formAPIUrl(city, temperatureUnits string) (string, error) {
	var APIQueryOptions string

	// Validate the temperature units and convert to a weather API query-string.
	switch strings.ToLower(temperatureUnits) {
	case "k":
		APIQueryOptions += "&units=kelvin"
	case "c":
		APIQueryOptions += "&units=metric"
	case "f":
		APIQueryOptions += "&units=imperial"
	default:
		return "", fmt.Errorf("Invalid temperature units %q while forming weather API url", temperatureUnits)
	}

	// Limit the weather API response to a single time-stamp.
	APIQueryOptions += "&cnt=1"

	u := fmt.Sprintf("https://%s%s/?q=%s&appid=%s%s", c.APIHost, c.APIUri, url.QueryEscape(city), c.APIKey, APIQueryOptions)
	return u, nil
}

// queryAPI accepts an OpenWeatherMap.org URL and queries its API.
func (c Client) queryAPI(url string) (APIResponse, error) {
	var apiRes APIResponse

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

// Forecast queries the weather API for the specified `city,state,country-code`,
// and temperature units.
func (c *Client) Forecast(city, temperatureUnits string) (string, error) {
	url, err := c.formAPIUrl(city, temperatureUnits)
	if err != nil {
		return "", fmt.Errorf("Error forming weather API URL for city %q, temperature units %q: %v", city, temperatureUnits, err)
	}

	res, err := c.queryAPI(url)
	if err != nil {
		return "", fmt.Errorf("Error querying weather API for city %q: %v", city, err)
	}

	// The formatForecast method returns its own error.
	return c.formatForecast(res)
}

// formatForecast accepts an API response and returns formatted output.
func (c *Client) formatForecast(ar APIResponse) (string, error) {
	if len(ar.List) == 0 {
		return "", fmt.Errorf("Empty response.List while formatting forecast")
	}

	if len(ar.List[0].Weather) == 0 {
		return "", fmt.Errorf("Empty response.List[0].Weather while formatting forecast")
	}

	forecast := fmt.Sprintf("%s, temp %.1f (feels like %.1f), humidity %.1f%%", ar.List[0].Weather[0].Description, ar.List[0].Main.Temp, ar.List[0].Main.Feels_like, ar.List[0].Main.Humidity)
	return forecast, nil
}
