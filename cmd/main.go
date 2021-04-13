// The weather CLI returns a brief weather forecast.
// TODO: use the cobra package
// to provide usage and nicer environment variable and command-line options.
package main

import (
	"fmt"
	"os"
	"weather"
)

func main() {
	apiKey := os.Getenv("WEATHERCASTER_API_KEY")
	if apiKey == "" {
		fmt.Fprintf(os.Stderr, "Please set the WEATHERCASTER_API_KEY environment variable to an OpenWeatherMap API key.\n")
fmt.Fprintf(os.Stderr, "To obtain an API key, see https://home.openweathermap.org/api_keys\n")
		os.Exit(1)
	}

	city := os.Getenv("WEATHERCASTER_CITY")
	if city == "" {
		city = "Great Neck Plaza,NY,US"
		fmt.Printf("Defaulting city to %s - to override, please set the WEATHERCASTER_CITY environment variable.\n", city)
	}

	wc := weather.NewClient(apiKey)

	json, err := wc.HttpGet(wc.FormUrl(city))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error calling API: %v\n", err)
		os.Exit(1)
	}

	err = wc.ParseJson(json)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing output from weather API: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("The brief weather forecast is %s\n", wc.GetDescription())

	fmt.Printf("The full API response is: %+v\n", wc.GetApiResponse())
}
