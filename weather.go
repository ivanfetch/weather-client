// Package weather eases getting and parsing weather data from OpenWeatherMap.org
package weather

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

// SpeedUnit represents a unit of speed as an integer.
type SpeedUnit int

// TempUnit represents a unit of temperature as an integer.
type TempUnit int

// Units of speed, the first listed is the default.
const (
	SpeedUnitMiles SpeedUnit = iota
	SpeedUnitMeters
)

// Units of temperature, the first listed is the default.
const (
	TempUnitFahrenheit TempUnit = iota
	TempUnitCelsius
	TempUnitKelvin
)

// speedUnitName stores friendly names for the speedUnit... constants.
var speedUnitName = map[SpeedUnit]string{
	SpeedUnitMiles:  "MPH",
	SpeedUnitMeters: "m/s",
}

// tempUnitName stores friendly names for the tempUnit... constants.
var tempUnitName = map[TempUnit]string{
	TempUnitFahrenheit: "ºF",
	TempUnitCelsius:    "ºC",
	TempUnitKelvin:     "ºK",
}

// conditions stores API-agnostic weather information.
type conditions struct {
	description            *string
	temperature, feelsLike *float64
	humidity               *float64
	windSpeed              *float64
}

// owmResponse stores fields from the OpenWeatherMap.org API `/2.5/forecast`.
// This does not fully mirror the API!
type owmResponse struct {
	List []struct {
		Weather []struct {
			Description *string
		}
		Main struct {
			Temp       *float64
			Feels_like *float64
			Humidity   *float64
		}
		Wind struct {
			Speed *float64
		}
	}
}

// Client stores properties of a weather client.
type Client struct {
	APIKey, APIHost, APIURI string
	speedUnit               SpeedUnit
	tempUnit                TempUnit
	HTTPClient              *http.Client
}

// ClientOption specifies weather.client options as functions.
type ClientOption func(*Client) error

// WithAPIHost sets the corresponding weather.client option.
func WithAPIHost(host string) ClientOption {
	return func(c *Client) error {
		c.APIHost = host
		return nil
	}
}

// WithAPIURI sets the corresponding weather.client option.
func WithAPIURI(uri string) ClientOption {
	return func(c *Client) error {
		c.APIURI = uri
		return nil
	}
}

// WithHTTPClient sets the corresponding weather.client option.
func WithHTTPClient(hc *http.Client) ClientOption {
	return func(c *Client) error {
		c.HTTPClient = hc
		return nil
	}
}

// WithSpeedUnit sets the corresponding weather.client option.
func WithSpeedUnit(u SpeedUnit) ClientOption {
	return func(c *Client) error {
		err := c.SetSpeedUnit(u)
		if err != nil {
			return err
		}
		return nil
	}
}

// WithTempUnit sets the corresponding weather.client option.
func WithTempUnit(u TempUnit) ClientOption {
	return func(c *Client) error {
		err := c.SetTempUnit(u)
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
		// This non-default client and its timeout is used
		// RE: https://medium.com/@nate510/don-t-use-go-s-default-http-client-4804cb19f779
		HTTPClient: &http.Client{Timeout: time.Second * 3},
	}

	for _, o := range options {
		err := o(c)
		if err != nil {
			return nil, err
		}
	}
	return c, nil
}

// GetSpeedUnit returns the configured speed unit for a weather client.
func (c *Client) GetSpeedUnit() SpeedUnit {
	return c.speedUnit
}

// GetTempUnit returns the configured temperature unit for a weather client.
func (c *Client) GetTempUnit() TempUnit {
	return c.tempUnit
}

// SetSpeedUnit validates then sets the speed unit for the weather client.
// Valid units are in the range of `SpeedUnit...` package constants.
func (c *Client) SetSpeedUnit(u SpeedUnit) error {
	if u == SpeedUnitMiles || u == SpeedUnitMeters {
		c.speedUnit = u
	} else {
		return fmt.Errorf("speed unit %v out of range, please use one of the SpeedUnitMeters or SpeedUnitMiles constants.\n", u)
	}
	return nil
}

// SetTempUnit validates then sets the temperature unit for the weather client.
// Valid units are in the range of `TempUnit...` package constants.
func (c *Client) SetTempUnit(u TempUnit) error {
	if u == TempUnitCelsius || u == TempUnitFahrenheit || u == TempUnitKelvin {
		c.tempUnit = u
	} else {
		return fmt.Errorf("temperature unit %v out of range, please use one of the TempUnitCelsius, TempUnitFahrenheit, or TempUnitKelvin constants.\n", u)
	}
	return nil
}

// ConvertTemp converts a temperature from Kelvin to the unit set in the weather client.
func (c Client) ConvertTemp(kelvin float64) float64 {
	var t float64
	switch c.tempUnit {
	case TempUnitCelsius:
		t = kelvin - 273.15
	case TempUnitFahrenheit:
		t = 1.8*(kelvin-273) + 32
	case TempUnitKelvin:
		// Input is already Kelvin
		return kelvin
	}
	return t
}

// ConvertSpeed converts a speed from meters/sec to the unit set in the weather client.
func (c Client) ConvertSpeed(meters float64) float64 {
	var s float64
	switch c.speedUnit {
	case SpeedUnitMeters:
		// Input is already meters/sec
		return meters
	case SpeedUnitMiles:
		s = meters * 2.236936
	}
	return s
}

// queryAPI accepts an OpenWeatherMap.org URL and queries its API.
func (c Client) queryAPI(url string) (conditions, error) {
	resp, err := c.HTTPClient.Get(url)
	if err != nil {
		return conditions{}, err
	}

	defer resp.Body.Close()

	// ioutil.ReadAll() returns a slice of bytes
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return conditions{}, err
	}

	if resp.StatusCode != http.StatusOK {
		return conditions{}, fmt.Errorf("HTTP %s returned from weather API: %v", resp.Status, string(data))
	}

	var ar owmResponse
	err = json.Unmarshal(data, &ar)
	if err != nil {
		return conditions{}, err
	}

	if len(ar.List) == 0 {
		return conditions{}, fmt.Errorf("unexpected empty `List` from weather API: %+v", ar)
	}

	if len(ar.List[0].Weather) == 0 {
		return conditions{}, fmt.Errorf("unexpected empty List[0].Weather from weather API: %+v", ar)
	}

	return conditions{

		description: ar.List[0].Weather[0].Description,
		temperature: ar.List[0].Main.Temp,
		feelsLike:   ar.List[0].Main.Feels_like,
		humidity:    ar.List[0].Main.Humidity,
		windSpeed:   ar.List[0].Wind.Speed,
	}, nil
}

// Forecast accepts a location and queries the weather API.
func (c *Client) Forecast(location string) (string, error) {
	url := fmt.Sprintf("%s%s/?q=%s&appid=%s&cnt=1", c.APIHost, c.APIURI, url.QueryEscape(location), c.APIKey)

	resp, err := c.queryAPI(url)
	if err != nil {
		return "", fmt.Errorf("Error querying weather API for location %q: %v", location, err)
	}

	// The formatForecast method returns its own error.
	return c.formatForecast(resp)
}

// formatForecast accepts weather conditions and returns formatted output.
func (c *Client) formatForecast(w conditions) (string, error) {
	tempUnit := tempUnitName[c.tempUnit]
	speedUnit := speedUnitName[c.speedUnit]
	var temperature, feelsLike, humidity, wind string

	if w.temperature != nil {
		temperature = fmt.Sprintf(", temp %.1f %v", c.ConvertTemp(*w.temperature), tempUnit)
	}

	if w.feelsLike != nil {
		feelsLike = fmt.Sprintf(", feels like %.1f %v", c.ConvertTemp(*w.feelsLike), tempUnit)
	}

	if w.humidity != nil {
		humidity = fmt.Sprintf(", humidity %.1f%%", *w.humidity)
	}

	if w.windSpeed != nil {
		wind = fmt.Sprintf(", wind %.1f %v", c.ConvertSpeed(*w.windSpeed), speedUnit)
	}

	forecast := fmt.Sprintf("%v%v%v%v%v", *w.description, temperature, feelsLike, humidity, wind)

	return forecast, nil
}

// RunCLI accepts CLI arguments, and output and error io.Writers,
// and outputs the forecast for a given location.
func RunCLI(args []string, output, errOutput io.Writer) error {
	apiKey := os.Getenv("OPENWEATHERMAP_API_KEY")
	if apiKey == "" {
		return fmt.Errorf(`Please set the OPENWEATHERMAP_API_KEY environment variable to an OpenWeatherMap API key.
		To obtain an API key, see https://home.openweathermap.org/api_keys`)
	}

	fs := flag.NewFlagSet("weather-caster", flag.ExitOnError)
	fs.SetOutput(errOutput)
	cliLocation := fs.String("l", "", `The location for which you want a weather forecast. Also specified via the WEATHERCASTER_LOCATION environment variable.
	A location can be specified as:
	"LocationName" (for well-known locations, such as London)
	"CitynName,StateName,CountryCode"
	For example: "Great Neck Plaza,NY,US"
`)

	cliSpeedUnit := fs.String("s", "", "Unit of measure to use when displaying wind speed (miles or meters). Also specified via the WEATHERCASTER_SPEED_UNIT environment variable. The default is miles.")
	cliTempUnit := fs.String("t", "", "Unit of measure to use when displaying temperature (c for Celsius, f for Fahrenheit, or k for kelvin). Also specified via the WEATHERCASTER_TEMP_UNIT environment variable. The default is Fahrenheit.")

	err := fs.Parse(args[1:])
	if err != nil {
		return err
	}

	// Use an environment variable if the unit command-line flags were not specified.
	if *cliSpeedUnit == "" {
		*cliSpeedUnit = os.Getenv("WEATHERCASTER_SPEED_UNIT")
	}
	if *cliTempUnit == "" {
		*cliTempUnit = os.Getenv("WEATHERCASTER_TEMP_UNIT")
	}

	// Use an environment variable if the location command-line flag was not specified.
	if *cliLocation == "" {
		*cliLocation = os.Getenv("WEATHERCASTER_LOCATION")
	}

	if *cliLocation == "" {
		return fmt.Errorf("Please specify a location using either the -l command-line flag, or by setting the WEATHERCASTER_LOCATION environment variable.")
	}

	var speedUnit SpeedUnit
	switch strings.ToLower(*cliSpeedUnit) {
	case "":
		// Use the `SpeedUnit` type default.
	case "mile", "miles":
		speedUnit = SpeedUnitMiles
	case "meter", "meters":
		speedUnit = SpeedUnitMeters
	default:
		return fmt.Errorf("Speed unit %q is invalid, please specify one of miles or meters.", *cliSpeedUnit)
	}

	var tempUnit TempUnit
	switch strings.ToLower(*cliTempUnit) {
	case "":
		// Use the `SpeedUnit` type default.
	case "c", "celsius":
		tempUnit = TempUnitCelsius
	case "f", "fahrenheit":
		tempUnit = TempUnitFahrenheit
	case "k", "kelvin":
		tempUnit = TempUnitKelvin
	default:
		return fmt.Errorf("Temperature unit %q is invalid, please specify one of c, f, or k for Celsius, Fahrenheit, or Kelvin respectively.", *cliTempUnit)
	}

	wc, err := NewClient(apiKey, WithSpeedUnit(speedUnit), WithTempUnit(tempUnit))
	if err != nil {
		return fmt.Errorf("Error creating weather client: %v\n", err)
	}

	forecast, err := wc.Forecast(*cliLocation)
	if err != nil {
		return err
	}

	fmt.Fprintln(output, forecast)
	return nil
}
