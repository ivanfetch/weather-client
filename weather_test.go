package weather_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"weather"
)

func TestQueryAPI(t *testing.T) {
	const testCity = "Great Neck Plaza,NY,US"
	const testUnits = "imperial"
	const testFileName = "testdata/greatneck.json"
	const wantRequestURL = "/data/2.5/forecast/?q=Great+Neck+Plaza%2CNY%2CUS&appid=DummyAPIKey&units=imperial&cnt=1"

	t.Parallel()

	f, err := os.Open(testFileName)
	if err != nil {
		t.Fatalf("unable to open test JSON file: %v", err)
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
			t.Errorf("Want %q, got %q comparing API URI while getting formatted forecast for city %q using units %v\n", wantRequestURL, gotRequestURL, testCity, testUnits)
		}
	}))
	defer ts.Close()

	wc, err := weather.NewClient("DummyAPIKey", weather.WithUnits(testUnits), weather.WithHTTPClient(ts.Client()), weather.WithAPIHost(ts.URL))
	if err != nil {
		t.Fatalf("Error while instanciating weather client to get formatted forecast for city %q using units %v: %v\n", testCity, testUnits, err)
	}

	want := "clear sky, temp 34.5 ºF (feels like 23.6 ºF), humidity 38.0%, wind 9.22 MPH"
	got, err := wc.Forecast(testCity)
	if err != nil {
		t.Fatalf("Error while getting formatted forecast for city %q using units %v: %v\n", testCity, testUnits, err)
	}

	if want != got {
		t.Errorf("Want %q, got %q, testing formatted forecast for city %q using units %v\n", want, got, testCity, testUnits)
	}
}

// Test FormAPIURL more deeply than TestForecast().
func TestFormAPIURL(t *testing.T) {
	// Define test cases
	testCases := []struct {
		city, units, want string
		errExpected       bool
	}{
		{
			city:  "Great Neck Plaza,NY,US",
			units: "si",
			want:  "https://api.openweathermap.org/data/2.5/forecast/?q=Great+Neck+Plaza%2CNY%2CUS&appid=DummyAPIKey&cnt=1",
		},
		{
			city:  "Great Neck Plaza,NY,US",
			units: "metric",
			want:  "https://api.openweathermap.org/data/2.5/forecast/?q=Great+Neck+Plaza%2CNY%2CUS&appid=DummyAPIKey&units=metric&cnt=1",
		},
		{
			city:  "Great Neck Plaza,NY,US",
			units: "imperial",
			want:  "https://api.openweathermap.org/data/2.5/forecast/?q=Great+Neck+Plaza%2CNY%2CUS&appid=DummyAPIKey&units=imperial&cnt=1",
		},
	}

	t.Parallel()

	for _, tc := range testCases {
		wc, err := weather.NewClient("DummyAPIKey", weather.WithUnits(tc.units))
		if err != nil {
			t.Fatalf("Error while instanciating weather client to form API URL for city %q and units %v: %v\n", tc.city, tc.units, err)
		}
		got, err := wc.FormAPIURL(tc.city)
		if err != nil {
			t.Fatalf("Error while forming API URL for city %q and units %v: %v\n", tc.city, tc.units, err)
		}

		if tc.want != got {
			t.Errorf("Want %q, got %q, forming API Url for city %s and units %v)\n", tc.want, got, tc.city, tc.units)
		}
	}
}
