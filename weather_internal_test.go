package weather

import (
	"testing"
)

// Test formAPIUrl, as the test for Forecast() ignores the URI.
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
	wc := NewClient("DummyAPIKey", WithMeasurementSystem(tc.units))
		got, err := wc.formAPIUrl(tc.city)
		if err != nil {
			t.Errorf("Error while forming API URL for city %q and units %v: %v\n", tc.city, tc.units, err)
		}

		if tc.want != got {
			t.Errorf("Want %q, got %q, forming API Url for city %s and units %v)\n", tc.want, got, tc.city, tc.units)
		}
	}
}
