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

const (
	defaultUnits = "imperial"
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
	APIKey, APIHost, APIURI, units string
	HTTPClient                     *http.Client
}

// An option is implemented as a function, to set the state of that option.
type ClientOption func(*Client) error

func WithAPIHost(host string) ClientOption {
	return func(c *Client) error {
		c.APIHost = host
		return nil
	}
}

func WithAPIURI(uri string) ClientOption {
	return func(c *Client) error {
		c.APIURI = uri
		return nil
	}
}

func WithHTTPClient(hc *http.Client) ClientOption {
	return func(c *Client) error {
		c.HTTPClient = hc
		return nil
	}
}

func WithUnits(u string) ClientOption {
	return func(c *Client) error {
		err := c.SetUnits(u)
		if err != nil {
			return err
		}
		return nil
	}
}

// NewClient returns a pointer to a new weather client.
func NewClient(APIKey string, options ...ClientOption) (*Client, error) {
	c := &Client{
		APIKey:  APIKey,
		APIHost: "https://api.openweathermap.org",
		APIURI:  "/data/2.5/forecast",
		units:   defaultUnits,
		// This non-default client and its timeout is used
		// RE: https://medium.com/@nate510/don-t-use-go-s-default-http-client-4804cb19f779
		HTTPClient: &http.Client{Timeout: time.Second * 3},
	}

	for _, o := range options {
		err := o(c)
		if err != nil {
			return &Client{}, err
		}
	}
	return c, nil
}

// SetUnits validates then sets the units for the weather client.
func (c *Client) SetUnits(u string) error {
	units := strings.ToLower(u)

	switch units {
	// An empty string sets the default value.
	case "":
		c.units = defaultUnits
		return nil
	case "si":
		c.units = units
		return nil
	case "metric":
		c.units = units
		return nil
	case "imperial":
		c.units = units
		return nil
	default:
		return fmt.Errorf("invalid value %q while setting units - please specify one of si, metric, or imperial\n", u)
	}
	return nil
}

// FormAPIURL accepts a city and returns an OpenWeatherMap.org URL.
func (c Client) FormAPIURL(city string) (string, error) {
	var APIQueryOptions string

	// Convert the units to a weather API query-string.
	switch c.units {
	case "si":
		// This is the OpenWeatherMap.org API default,
		// no URL query-string is required.
	default:
		// All other valid units can be specified directly in the query-string.
		APIQueryOptions += fmt.Sprintf("&units=%s", c.units)
	}

	// Limit the weather API response to a single time-stamp.
	APIQueryOptions += "&cnt=1"

	u := fmt.Sprintf("%s%s/?q=%s&appid=%s%s", c.APIHost, c.APIURI, url.QueryEscape(city), c.APIKey, APIQueryOptions)
	return u, nil
}

// queryAPI accepts an OpenWeatherMap.org URL and queries its API.
func (c Client) queryAPI(url string) (APIResponse, error) {
	var apiRespp APIResponse

	resp, err := c.HTTPClient.Get(url)
	if err != nil {
		return APIResponse{}, err
	}

	defer resp.Body.Close()

	// ioutil.ReadAll() returns a slice of bytes
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return APIResponse{}, err
	}

	if resp.StatusCode != http.StatusOK {
		return apiRespp, fmt.Errorf("HTTP %d returned from weather API: %v", resp.StatusCode, string(data))
	}

	err = json.Unmarshal(data, &apiRespp)
	if err != nil {
		return APIResponse{}, err
	}

	if len(apiRespp.List) == 0 {
		return APIResponse{}, fmt.Errorf("Empty response.List while querying weather API")
	}

	if len(apiRespp.List[0].Weather) == 0 {
		return APIResponse{}, fmt.Errorf("Empty response.List[0].Weather while querying weather API")
	}
	return apiRespp, nil
}

// Forecast accepts a city, and queries the weather API.
func (c *Client) Forecast(city string) (string, error) {
	url, err := c.FormAPIURL(city)
	if err != nil {
		return "", fmt.Errorf("Error forming weather API URL for city %q: %v", city, err)
	}

	res, err := c.queryAPI(url)
	if err != nil {
		return "", fmt.Errorf("Error querying weather API for city %q: %v", city, err)
	}

	// The formatForecast method returns its own error.
	return c.formatForecast(res)
}

// speedUnits returns the unit of speed per the units set in the Client.
func (c *Client) speedUnits() string {
	switch c.units {
	case "si":
		return "m/s"
	case "metric":
		return "m/s"
	case "imperial":
		return "MPH"
	}
	// We should never get here, also perhaps we should return an error?
	return "unknown"
}

// tempUnits returns the unit of temperature per the units set in the Client.
func (c *Client) tempUnits() string {
	switch c.units {
	case "si":
		return "ºK"
	case "metric":
		return "ºC"
	case "imperial":
		return "ºF"
	}
	// We should never get here, also perhaps we should return an error?
	return "unknown"
}

// formatForecast accepts an API response,
// and returns formatted output.
func (c *Client) formatForecast(ar APIResponse) (string, error) {
	tempUnits := c.tempUnits()
	windUnits := c.speedUnits()

	forecast := fmt.Sprintf("%s, temp %.1f %v (feels like %.1f %v), humidity %.1f%%", ar.List[0].Weather[0].Description, ar.List[0].Main.Temp, tempUnits, ar.List[0].Main.Feels_like, tempUnits, ar.List[0].Main.Humidity)

	if ar.List[0].Wind.Speed > 0 {
		forecast += fmt.Sprintf(", wind %v %v", ar.List[0].Wind.Speed, windUnits)
	}

	return forecast, nil
}
