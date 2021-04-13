package weather_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"weather"
)

func TestQueryAPI(t *testing.T) {
	const testCity = "Great Neck Plaza,NY,US"
	const testUnits = "imperial"
	const testJSON = `{"cod":"200","message":0,"cnt":1,"list":[{"dt":1616220000,"main":{"temp":34.47,"feels_like":23.59,"temp_min":33.94,"temp_max":34.47,"pressure":1031,"sea_level":1031,"grnd_level":1027,"humidity":38,"temp_kf":0.29},"weather":[{"id":800,"main":"Clear","description":"clear sky","icon":"01n"}],"clouds":{"all":1},"wind":{"speed":9.22,"deg":5},"visibility":10000,"pop":0,"sys":{"pod":"n"},"dt_txt":"2021-03-20 06:00:00"}],"city":{"id":5119226,"name":"Great Neck Plaza","coord":{"lat":40.7868,"lon":-73.7265},"country":"US","population":6707,"timezone":-14400,"sunrise":1616151573,"sunset":1616195137}}`
	const wantRequestURL = "/data/2.5/forecast/?q=Great+Neck+Plaza%2CNY%2CUS&appid=DummyAPIKey&units=imperial&cnt=1"

	t.Parallel()

	// Create a test HTTP server,
	// and populate it with JSON as though served by the weather API.
	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, testJSON)
		gotRequestURL := r.URL.String()
		if wantRequestURL != gotRequestURL {
			// t.FatalF will abort the http.HandlerFunc,
			// causing a second test to unnecesarily fail.
			t.Errorf("Want %q, got %q comparing API URI while getting formatted forecast for city %q using units %v\n", wantRequestURL, gotRequestURL, testCity, testUnits)
		}
	}))
	defer ts.Close()

	wc := weather.NewClient("DummyAPIKey", weather.WithUnits(testUnits))
	wc.HTTPClient = ts.Client()
	wc.APIHost = ts.URL

	want := "clear sky, temp 34.5 ºF (feels like 23.6 ºF), humidity 38.0%, wind 9.22 MPH"
	got, err := wc.Forecast(testCity)
	if err != nil {
		t.Fatalf("Error while getting formatted forecast for city %q using units %v: %v\n", testCity, testUnits, err)
	}

	if want != got {
		t.Errorf("Want %q, got %q, testing formatted forecast for city %q using units %v\n", want, got, testCity, testUnits)
	}
}

// Test FormAPIUrl more deeply than TestForecast().
func TestFormAPIUrl(t *testing.T) {
	// Define test cases
	testCases := []struct {
		city, units, want string
		errExpected       bool
	}{
		{
			city:  "Great Neck Plaza,NY,US",
			units: "standard",
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
		wc := weather.NewClient("DummyAPIKey", weather.WithUnits(tc.units))
		got, err := wc.FormAPIUrl(tc.city)
		if err != nil {
			t.Fatalf("Error while forming API URL for city %q and units %v: %v\n", tc.city, tc.units, err)
		}

		if tc.want != got {
			t.Errorf("Want %q, got %q, forming API Url for city %s and units %v)\n", tc.want, got, tc.city, tc.units)
		}
	}
}
