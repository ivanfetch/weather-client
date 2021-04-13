// Package weather eases getting and parsing weather data from OpenWeatherMap.org
package weather

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
)

// APIResponse matches fields from the OpenWeatherMap.org API `/2.5/forecast`.
// This does not fully mirror the API!
type APIResponse struct {
	List []struct {
		Weather []struct {
			Description string
		}
	}
}

// An OpenWeatherMap.org client
type Client struct {
	APIKey, APIHost, APIUri, APIQueryOptions string
}

// NewClient returns a pointer to a new weather client.
func NewClient(APIKey string) *Client {
	return &Client{
		APIKey:          APIKey,
		APIHost:         "api.openweathermap.org",
		APIUri:          "/data/2.5/forecast",
		APIQueryOptions: "&units=imperial&cnt=1",
	}
}

// formAPIUrl accepts a city and returns an OpenWeatherMap.org URL.
func (c Client) formAPIUrl(city string) string {
	// This will eventually vary the URL using client configuration,
	// E.G. temperature units.
	u := fmt.Sprintf("https://%s%s/?q=%s&appid=%s%s", c.APIHost, c.APIUri, url.QueryEscape(city), c.APIKey, c.APIQueryOptions)
	return u
}

// queryAPI accepts an OpenWeatherMap.org URL and queries its API.
func (c Client) queryAPI(url string) (APIResponse, error) {
	var apiRes APIResponse

	// This non-default client and its timeout is used
	// RE: https://medium.com/@nate510/don-t-use-go-s-default-http-client-4804cb19f779
	httpClient := http.Client{Timeout: time.Second * 3}
	httpRes, err := httpClient.Get(url)
	if err != nil {
		return apiRes, err
	}

	defer httpRes.Body.Close()

	// ioutil.ReadAll() returns a slice of bytes
	body, err := ioutil.ReadAll(httpRes.Body)
	if err != nil {
		return apiRes, err
	}

	if httpRes.StatusCode >= 400 {
		return apiRes, fmt.Errorf("HTTP %d returned from weather API: %v", httpRes.StatusCode, string(body))
	}

	jsonBytes := []byte(string(body))
	err = json.Unmarshal(jsonBytes, &apiRes)
	if err != nil {
		return apiRes, err
	}
	return apiRes, nil
}

// Forecast queries the weather API for a `city,state,country-code`,
// and stores the result in the Client object.
func (c *Client) Forecast(city string) (string, error) {
	res, err := c.queryAPI(c.formAPIUrl(city))
	if err != nil {
		return "", fmt.Errorf("Error querying weather API for city %q: %v", city, err)
	}

	// The formatForecast method returns its own error.
	return c.formatForecast(res)
}

// formatForecasts returns formatted output from an API response.
func (c *Client) formatForecast(ar APIResponse) (string, error) {
	if len(ar.List) == 0 {
		return "", fmt.Errorf("Empty response.List while formatting forecast")
	}

	if len(ar.List[0].Weather) == 0 {
		return "", fmt.Errorf("Empty response.List[0].Weather while formatting forecast")
	}

	return ar.List[0].Weather[0].Description, nil
}
