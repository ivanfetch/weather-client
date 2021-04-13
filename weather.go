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
		Wind struct {
			Speed float64
		}
	}
}

// An OpenWeatherMap.org client
type Client struct {
	APIKey, APIHost, APIUri string
	HTTPClient              *http.Client
}

// NewClient returns a pointer to a new weather client.
func NewClient(APIKey string) *Client {
	return &Client{
		APIKey:  APIKey,
		APIHost: "https://api.openweathermap.org",
		APIUri:  "/data/2.5/forecast",
		// This non-default client and its timeout is used
		// RE: https://medium.com/@nate510/don-t-use-go-s-default-http-client-4804cb19f779
		HTTPClient: &http.Client{Timeout: time.Second * 3},
	}
}

// formAPIUrl accepts a city and measurement system, and returns an OpenWeatherMap.org URL.
func (c Client) formAPIUrl(city, measurementSystem string) (string, error) {
	var APIQueryOptions string

	// Validate the measurement system and convert to a weather API query-string.
	switch strings.ToLower(measurementSystem) {
	case "standard":
		// The OpenWeatherMap.org API default is standard,
		// so no URL query-string is required.
	case "metric":
		APIQueryOptions += "&units=metric"
	case "imperial":
		APIQueryOptions += "&units=imperial"
	default:
		return "", fmt.Errorf("Invalid measurement system %q while forming weather API url", measurementSystem)
	}

	// Limit the weather API response to a single time-stamp.
	APIQueryOptions += "&cnt=1"

	u := fmt.Sprintf("%s%s/?q=%s&appid=%s%s", c.APIHost, c.APIUri, url.QueryEscape(city), c.APIKey, APIQueryOptions)
	return u, nil
}

// queryAPI accepts an OpenWeatherMap.org URL and queries its API.
func (c Client) queryAPI(url string) (APIResponse, error) {
	var apiRes APIResponse

	httpRes, err := c.HTTPClient.Get(url)
	if err != nil {
		return APIResponse{}, err
	}

	defer httpRes.Body.Close()

	// ioutil.ReadAll() returns a slice of bytes
	data, err := ioutil.ReadAll(httpRes.Body)
	if err != nil {
		return APIResponse{}, err
	}

	if httpRes.StatusCode != http.StatusOK {
		return apiRes, fmt.Errorf("HTTP %d returned from weather API: %v", httpRes.StatusCode, string(data))
	}

	err = json.Unmarshal(data, &apiRes)
	if err != nil {
		return APIResponse{}, err
	}
	return apiRes, nil
}

// Forecast accepts a city and measurement system, and queries the weather API.
func (c *Client) Forecast(city, measurementSystem string) (string, error) {
	url, err := c.formAPIUrl(city, measurementSystem)
	if err != nil {
		return "", fmt.Errorf("Error forming weather API URL for city %q, measurement system %q: %v", city, measurementSystem, err)
	}

	res, err := c.queryAPI(url)
	if err != nil {
		return "", fmt.Errorf("Error querying weather API for city %q: %v", city, err)
	}

	// The formatForecast method returns its own error.
	return c.formatForecast(res, measurementSystem)
}

// formatForecast accepts an API response and measurement system,
// and returns formatted output.
func (c *Client) formatForecast(ar APIResponse, measurementSystem string) (string, error) {
	if len(ar.List) == 0 {
		return "", fmt.Errorf("Empty response.List while formatting forecast")
	}

	if len(ar.List[0].Weather) == 0 {
		return "", fmt.Errorf("Empty response.List[0].Weather while formatting forecast")
	}

	var tempUnits, windUnits string
	switch strings.ToLower(measurementSystem) {
	case "standard":
		tempUnits = "K"
		windUnits = "m/s"
	case "metric":
		tempUnits = "C"
		windUnits = "m/s"
	case "imperial":
		tempUnits = "F"
		windUnits = "MPH"
	default:
		return "", fmt.Errorf("unknown measurement system while formatting forecast: %q", measurementSystem)
	}

	forecast := fmt.Sprintf("%s, temp %.1f %v (feels like %.1f %v), humidity %.1f%%", ar.List[0].Weather[0].Description, ar.List[0].Main.Temp, tempUnits, ar.List[0].Main.Feels_like, tempUnits, ar.List[0].Main.Humidity)

	if ar.List[0].Wind.Speed > 0 {
		forecast += fmt.Sprintf(", wind %v %v", ar.List[0].Wind.Speed, windUnits)
	}

	return forecast, nil
}
