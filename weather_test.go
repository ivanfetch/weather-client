package weather_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"weather"
)

func TestForecast(t *testing.T) {
	t.Parallel()

	const testLocation = "Great Neck Plaza,NY,US"
	const testFileName = "testdata/greatneck.json"
	const wantRequestURL = "/data/2.5/forecast/?q=Great+Neck+Plaza%2CNY%2CUS&appid=DummyAPIKey&cnt=1"

	// Define test cases
	testCases := []struct {
		description       string
		setSpeedUnit      weather.SpeedUnit
		setTempUnit       weather.TempUnit
		want              string
		clientErrExpected bool
	}{
		{
			description:  "speed meters and temp kelvin",
			setSpeedUnit: weather.SpeedUnitMeters,
			setTempUnit:  weather.TempUnitKelvin,
			want:         "overcast clouds, temp 286.0K, feels like 285.7K, humidity 92.0%, wind 2.5 m/s",
		},
		{
			description:  "speed meters and temp celsius",
			setSpeedUnit: weather.SpeedUnitMeters,
			setTempUnit:  weather.TempUnitCelsius,
			want:         "overcast clouds, temp 12.9 ºC, feels like 12.6 ºC, humidity 92.0%, wind 2.5 m/s",
		},
		{
			description:  "speed miles and temp fahrenheit",
			setSpeedUnit: weather.SpeedUnitMiles,
			setTempUnit:  weather.TempUnitFahrenheit,
			want:         "overcast clouds, temp 55.4 ºF, feels like 54.9 ºF, humidity 92.0%, wind 5.6 mph",
		},
		{
			description:       "speed miles and invalid temp",
			setSpeedUnit:      weather.SpeedUnitMiles,
			setTempUnit:       30, // out of range int
			clientErrExpected: true,
		},
	}

	for _, tc := range testCases {
		f, err := os.Open(testFileName)
		if err != nil {
			t.Fatalf("%v", err)
		}
		defer f.Close()

		// Create a test HTTP server,
		// and populate it with JSON as though served by the weather API.
		// The `HandlerFunc` will be called when the test HTTP client
		// queries the test server.
		ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, err := io.Copy(w, f)
			if err != nil {
				t.Fatalf("unable to copy test JSON from file %s to test HTTP server: %v", testFileName, err)
			}
			gotRequestURL := r.URL.String()
			if wantRequestURL != gotRequestURL {
				// t.ErrorF is used because FatalF will abort the http.HandlerFunc,
				// causing the QueryAPI test to output failure.
				t.Errorf("Want %q, got %q comparing API URI", wantRequestURL, gotRequestURL)
			}
		}))
		defer ts.Close()

		wc, err := weather.NewClient("DummyAPIKey",
			weather.WithSpeedUnit(tc.setSpeedUnit),
			weather.WithTempUnit(tc.setTempUnit),
			weather.WithHTTPClient(ts.Client()),
			weather.WithAPIHost(ts.URL),
		)
		if !tc.clientErrExpected && err != nil {
			t.Fatalf("Error while instanciating weather client for test %v: %v", tc.description, err)
		}

		// Only get a forecast and compare results if the test-case did not expect
		// an error from the client constructor.
		if !tc.clientErrExpected {
			got, err := wc.Forecast(testLocation)
			if err != nil {
				t.Fatalf("Error while getting forecast for location %q: %v", testLocation, err)
			}

			if tc.want != got {
				t.Errorf("Want %q, got %q, testing %v", tc.want, got, tc.description)
			}
		}
	}
}
func TestProcessCLISpeedUnit(t *testing.T) {
	t.Parallel()

	// Define test cases
	testCases := []struct {
		userInput   string
		want        weather.SpeedUnit
		errExpected bool
	}{
		{
			userInput: "", // default case
			want:      weather.SpeedUnitMiles,
		},
		{
			userInput: "meters",
			want:      weather.SpeedUnitMeters,
		},
		{
			userInput: "miles",
			want:      weather.SpeedUnitMiles,
		},
		{
			userInput:   "feet",
			errExpected: true,
		},
	}

	for _, tc := range testCases {
		got, err := weather.ProcessCLISpeedUnit(tc.userInput)
		if !tc.errExpected && err != nil {
			t.Fatalf("error for user input %q: %v", tc.userInput, err)
		}

		if !tc.errExpected && tc.want != got {
			t.Fatalf("want %q, got %q, for user input %q", tc.want, got, tc.userInput)
		}
	}
}

func TestProcessCLITempUnit(t *testing.T) {
	t.Parallel()

	// Define test cases
	testCases := []struct {
		userInput   string
		want        weather.TempUnit
		errExpected bool
	}{
		{
			userInput: "", // default case
			want:      weather.TempUnitFahrenheit,
		},
		{
			userInput: "f",
			want:      weather.TempUnitFahrenheit,
		},
		{
			userInput: "c",
			want:      weather.TempUnitCelsius,
		},
		{
			userInput: "k",
			want:      weather.TempUnitKelvin,
		},
		{
			userInput:   "x",
			errExpected: true,
		},
	}

	for _, tc := range testCases {
		got, err := weather.ProcessCLITempUnit(tc.userInput)
		if !tc.errExpected && err != nil {
			t.Fatalf("error for user input %q: %v", tc.userInput, err)
		}

		if !tc.errExpected && tc.want != got {
			t.Fatalf("want %q, got %q, for user input %q", tc.want, got, tc.userInput)
		}
	}
}
