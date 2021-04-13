// The weather CLI returns a brief weather forecast.
package main

import (
	"flag"
	"fmt"
	"os"
	"weather"
)

func main() {
	apiKey := os.Getenv("OPENWEATHERMAP_API_KEY")
	if apiKey == "" {
		fmt.Fprintf(os.Stderr, "Please set the OPENWEATHERMAP_API_KEY environment variable to an OpenWeatherMap API key.\n")
		fmt.Fprintf(os.Stderr, "To obtain an API key, see https://home.openweathermap.org/api_keys\n")
		os.Exit(1)
	}

	var city *string
	city = flag.String("city", "", `The name of the city  to get a weather forecast. A city can be specified as:
	"CityName"
	"CityName,StateName"
	"CityName,StateName,CountryCode"
	For example: "Great Neck Plaza,NY,US"
`)

	flag.Parse()
	if *city == "" {
		*city = os.Getenv("WEATHERCASTER_CITY")
	}
	if *city == "" {
		fmt.Println("Please specify a city using either the -city command-line flag, or by setting the WEATHERCASTER_CITY environment variable.")
		os.Exit(1)
	}

	wc := weather.NewClient(apiKey)

	forecast, err := wc.Forecast(*city)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	fmt.Printf("The brief weather forecast is %s\n", forecast)

}
