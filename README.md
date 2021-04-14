# Weather Command-line Client

A command-line weather client that produces a brief summary forecast by querying the [OpenWeatherMap.org forecast API](https://openweathermap.org/forecast5).

This is a learning project written in Go, that aims to provide both a useful weather package and a command-line interface - see more in the design section below.

```bash
$ ./weather -l Miami
clear sky, temp 81.1 ºF, feels like 82.7 ºF, humidity 57.0%, wind 9.9 mph
```

```bash
$ ./weather -l Miami -t celsius -s meters
clear sky, temp 26.7 ºC, feels like 26.4 ºC, humidity 34.0%, wind 4.4 m/s
```

## Usage

For now there are no pre-build releases, so you'll need to first [have Golang installed](https://golang.org/doc/install), then compile this.

* Create [an account on OpenWeathermap.org](https://home.openweathermap.org/users/sign_up), then get [your API key](https://home.openweathermap.org/api_keys).
* Clone this repository and change into its directory: `git clone https://github.com/ivanfetch/weather-client && cd weather-client`
* Build the weather client: `go build -o weather ./cmd/main.go`
* Set an environment variable with your OpenWeatherMap.org API key: `export OPENWEATHERMAP_API_KEY=YourActualAPIKey`
* `./weather -l "new york,ny,us"`

Run `./weather -h` for options.

## Design / Goals

This learning project is designed to be useful, represent good practices, and help me further my own Go standards and continue to learn.

* Continue learning Go, including interacting with an API, and solidifying what is "good," readable code?
* Describe key weather forecast conditions in a basic summary. Don't try to be the best full-featured weather interface in the world, but add features as the project and my process evolve. :)
* Provide a weather package that would be usable alone, and is not only an internal means to the accompanying command-line interface.
* Get my brain further wrapped around an approach to testing, and testing behaviors more than Functions - as described in [point #3 of John Arundel's "Ten commandments of Go" post](https://bitfieldconsulting.com/golang/commandments).
	* Thinking about code and testing in away that appropriately informs each other. Always write tests first? Write functions that turn out to be hard to test? What is worth testing, and what is not?
* Consider how much to decouple / abstract - a journey:
	* How much do I allow the weather API to help me? It facilitates returning data for different temperature and wind-speed units. Alternatively I can calculate those in my code.
	* How much do I allow the weather API to be reflected in the weather package user interface - is some abstraction of its API data structure a good thing, or unnecessary overhead / complexity?


## Future Possibilities

Some possible future expansion and to-do items:

* Use multiple forecast time-stamps? OpenWeatherMap.org returns a list of multiple time-stamps in its forecast results. I currently return data from a single time-stamp, but perhaps I should factor multiple timestamps into my forecast.
* Cache results, to avoid querying the weather API on each run. The OpenWeatherMap.org data is updated every ten minutes.
* Use additional weather data in my forecast. E.G. `temp_min` and `temp_max`, although API docs describe that as "min and max so far" instead of a forecasted range.
* Support using other weather APIs / providers. . .