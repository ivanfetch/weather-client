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
	defaultTemperatureUnits = "f"
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
	"CityName"
	"CityName,StateName"
	"CityName,StateName,CountryCode"
	For example: "Great Neck Plaza,NY,US"
`)

	temperatureUnits := flag.String("units", "", "Units to display temperature (k for kelvin, c for celsius, or f for fahrenheit). Also specified via the WEATHERCASTER_UNITS environment variable.")

	flag.Parse()

	// Use an environment variable if the command-line flag was not specified.
	if *temperatureUnits == "" {
		*temperatureUnits = os.Getenv("WEATHERCASTER_UNITS")
	}

	switch strings.ToLower(*temperatureUnits) {
	case "":
		// The default is handled here, to allow the environment variable to override a non-specified flag.
		*temperatureUnits = defaultTemperatureUnits
	case "k":
	case "c":
	case "f":
	default:
		fmt.Fprintf(os.Stderr, "Invalid temperature units %q - please specify one of k, c, or f\n", *temperatureUnits)
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

	wc := weather.NewClient(apiKey)

	forecast, err := wc.Forecast(*city, *temperatureUnits)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	fmt.Println(forecast)
}
