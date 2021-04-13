// The weather CLI returns a brief weather forecast.
package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"weather"
)

const (
	defaultUnits = "imperial"
)

func main() {
	apiKey := os.Getenv("OPENWEATHERMAP_API_KEY")
	if apiKey == "" {
		fmt.Fprintf(os.Stderr, "Please set the OPENWEATHERMAP_API_KEY environment variable to an OpenWeatherMap API key.\n")
		fmt.Fprintf(os.Stderr, "To obtain an API key, see https://home.openweathermap.org/api_keys\n")
		os.Exit(1)
	}

	city := flag.String("city", "", `The name of the city for which you want a weather forecast. Also specified via the WEATHERCASTER_CITY environment variable.
	A city can be specified as:
	"CityName" (for well-known locations)
	"CityName,StateName,CountryCode"
	For example: "Great Neck Plaza,NY,US"
`)

	units := flag.String("units", "", "Units to use when obtaining and displaying temperature and wind-speed (standard for kelvin and meters, metric for celsius and meters, or imperial for fahrenheit and miles-per-hour). Also specified via the WEATHERCASTER_UNITS environment variable.")

	flag.Parse()

	// Use an environment variable if the command-line flag was not specified.
	if *units == "" {
		*units = os.Getenv("WEATHERCASTER_UNITS")
	}

	switch strings.ToLower(*units) {
	case "":
		// The default is handled here, to allow the environment variable to override a non-specified flag.
		*units = defaultUnits
	case "standard":
	case "metric":
	case "imperial":
	default:
		fmt.Fprintf(os.Stderr, "Invalid units %q - please specify one of standard, metric, or imperial\n", *units)
		os.Exit(1)
	}

	// Use an environment variable if the command-line flag was not specified.
	if *city == "" {
		*city = os.Getenv("WEATHERCASTER_CITY")
	}

	if *city == "" {
		fmt.Println("Please specify a city using either the -city command-line flag, or by setting the WEATHERCASTER_CITY environment variable.")
		os.Exit(1)
	}

	wc := weather.NewClient(apiKey, weather.WithUnits(*units))

	forecast, err := wc.Forecast(*city)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	fmt.Println(forecast)
}
