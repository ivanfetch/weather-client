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
	APIKey, APIHost, APIUri, MeasurementSystem string
	HTTPClient                                 *http.Client
}

// An option is implemented as a function, to set the state of that option.
type ClientOption func(*Client)

func WithAPIHost(host string) ClientOption {
	return func(c *Client) {
		c.APIHost = host
	}
}

func WithAPIUri(uri string) ClientOption {
	return func(c *Client) {
		c.APIUri = uri
	}
}

func WithHTTPClient(hc *http.Client) ClientOption {
	return func(c *Client) {
		c.HTTPClient = hc
	}
}

func WithMeasurementSystem(ms string) ClientOption {
	return func(c *Client) {
		c.MeasurementSystem = ms
	}
}

// NewClient returns a pointer to a new weather client.
func NewClient(APIKey string, options ...ClientOption) *Client {
	c := &Client{
		APIKey:  APIKey,
		APIHost: "https://api.openweathermap.org",
		APIUri:  "/data/2.5/forecast",
		// This non-default client and its timeout is used
		// RE: https://medium.com/@nate510/don-t-use-go-s-default-http-client-4804cb19f779
		HTTPClient:        &http.Client{Timeout: time.Second * 3},
		MeasurementSystem: "imperial",
	}

	for _, o := range options {
		o(c)
	}
	return c
}

// formAPIUrl accepts a city and returns an OpenWeatherMap.org URL.
func (c Client) formAPIUrl(city string) (string, error) {
	var APIQueryOptions string

	// Convert the measurement system to a weather API query-string.
	switch strings.ToLower(c.MeasurementSystem) {
	case "standard":
		// The OpenWeatherMap.org API default is standard,
		// so no URL query-string is required.
	default:
		// All other valid metric system values can be specified directly in the query-string.
		APIQueryOptions += fmt.Sprintf("&units=%s", strings.ToLower(c.MeasurementSystem))
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

	if len(apiRes.List) == 0 {
		return APIResponse{}, fmt.Errorf("Empty response.List while querying weather API")
	}

	if len(apiRes.List[0].Weather) == 0 {
		return APIResponse{}, fmt.Errorf("Empty response.List[0].Weather while querying weather API")
	}
	return apiRes, nil
}

// Forecast accepts a city and measurement system, and queries the weather API.
func (c *Client) Forecast(city, measurementSystem string) (string, error) {
	url, err := c.formAPIUrl(city)
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

	var tempUnits, windUnits string
	switch strings.ToLower(measurementSystem) {
	case "standard":
		tempUnits = "ºK"
		windUnits = "m/s"
	case "metric":
		tempUnits = "ºC"
		windUnits = "m/s"
	case "imperial":
		tempUnits = "ºF"
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
