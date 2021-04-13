// Package weather eases getting and parsing weather data from OpenWeatherMap.org
package weather

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

const (
	defaultUnits = "imperial"
)

// weatherConditions stores API-agnostic weather information.
type weatherConditions struct {
	description            string
	temperature, feelsLike float64
	humidity               float64
	windSpeed              float64
}

// OWMResponse matches fields from the OpenWeatherMap.org API `/2.5/forecast`.
// This does not fully mirror the API!
type OWMResponse struct {
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

// A weather client
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

// GetUnits returns the configured units for a weather client.
func (c *Client) GetUnits() string {
	return c.units
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
		// All other possible units can be specified directly in the query-string.
		APIQueryOptions += fmt.Sprintf("&units=%s", c.units)
	}

	// Limit the weather API response to a single time-stamp.
	APIQueryOptions += "&cnt=1"

	u := fmt.Sprintf("%s%s/?q=%s&appid=%s%s", c.APIHost, c.APIURI, url.QueryEscape(city), c.APIKey, APIQueryOptions)
	return u, nil
}

// queryAPI accepts an OpenWeatherMap.org URL and queries its API.
func (c Client) queryAPI(url string) (weatherConditions, error) {
	resp, err := c.HTTPClient.Get(url)
	if err != nil {
		return weatherConditions{}, err
	}

	defer resp.Body.Close()

	// ioutil.ReadAll() returns a slice of bytes
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return weatherConditions{}, err
	}

	if resp.StatusCode != http.StatusOK {
		return weatherConditions{}, fmt.Errorf("HTTP %d returned from weather API: %v", resp.StatusCode, string(data))
	}

	var ar OWMResponse
	err = json.Unmarshal(data, &ar)
	if err != nil {
		return weatherConditions{}, err
	}

	if len(ar.List) == 0 {
		return weatherConditions{}, fmt.Errorf("Empty response.List while querying weather API")
	}

	if len(ar.List[0].Weather) == 0 {
		return weatherConditions{}, fmt.Errorf("Empty response.List[0].Weather while querying weather API")
	}

	var w weatherConditions
	w.description = ar.List[0].Weather[0].Description
	w.temperature = ar.List[0].Main.Temp
	w.feelsLike = ar.List[0].Main.Feels_like
	w.humidity = ar.List[0].Main.Humidity
	w.windSpeed = ar.List[0].Wind.Speed

	return w, nil
}

// Forecast accepts a city, and queries the weather API.
func (c *Client) Forecast(city string) (string, error) {
	url, err := c.FormAPIURL(city)
	if err != nil {
		return "", fmt.Errorf("Error forming weather API URL for city %q: %v", city, err)
	}

	resp, err := c.queryAPI(url)
	if err != nil {
		return "", fmt.Errorf("Error querying weather API for city %q: %v", city, err)
	}

	// The formatForecast method returns its own error.
	return c.formatForecast(resp)
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
	return "º"
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
	return "º"
}

// formatForecast accepts an API response,
// and returns formatted output.
func (c *Client) formatForecast(w weatherConditions) (string, error) {
	tempUnits := c.tempUnits()
	windUnits := c.speedUnits()

	forecast := fmt.Sprintf("%s, temp %.1f %v (feels like %.1f %v), humidity %.1f%%", w.description, w.temperature, tempUnits, w.feelsLike, tempUnits, w.humidity)

	if w.windSpeed > 0 {
		forecast += fmt.Sprintf(", wind %v %v", w.windSpeed, windUnits)
	}

	return forecast, nil
}

// RunCLI processes CLI arguments and outputs the forecast for a given city.
func RunCLI(args []string) error {
	apiKey := os.Getenv("OPENWEATHERMAP_API_KEY")
	if apiKey == "" {
		return fmt.Errorf(`Please set the OPENWEATHERMAP_API_KEY environment variable to an OpenWeatherMap API key.
		To obtain an API key, see https://home.openweathermap.org/api_keys`)
	}

	fs := flag.NewFlagSet("weather-caster", flag.ExitOnError)
	fs.SetOutput(os.Stderr)
	city := fs.String("city", "", `The name of the city for which you want a weather forecast. Also specified via the WEATHERCASTER_CITY environment variable.
	A city can be specified as:
	"CityName" (for well-known locations)
	"CityName,StateName,CountryCode"
	For example: "Great Neck Plaza,NY,US"
`)

	units := fs.String("units", "", "Units to use when obtaining and displaying temperature and wind-speed (si for kelvin and meters, metric for celsius and meters, or imperial for fahrenheit and miles-per-hour). Also specified via the WEATHERCASTER_UNITS environment variable.")

	err := fs.Parse(args[1:])
	if err != nil {
		return err
	}

	// Use an environment variable if the units command-line flag was not specified.
	if *units == "" {
		*units = os.Getenv("WEATHERCASTER_UNITS")
	}

	// Use an environment variable if the city command-line flag was not specified.
	if *city == "" {
		*city = os.Getenv("WEATHERCASTER_CITY")
	}

	if *city == "" {
		return fmt.Errorf("Please specify a city using either the -city command-line flag, or by setting the WEATHERCASTER_CITY environment variable.")
	}

	wc, err := NewClient(apiKey, WithUnits(*units))
	if err != nil {
		return fmt.Errorf("Error creating weather client: %v\n", err)
	}

	forecast, err := wc.Forecast(*city)
	if err != nil {
		return err
	}

	fmt.Println(forecast)
	return nil
}
