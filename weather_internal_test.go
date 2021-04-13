package weather

import (
	"testing"
)

func TestFormAPIUrl(t *testing.T) {
	// THis may turn into a table-test to test multiple units,
	// if this test ends up being kept.
	const testCity = "Great Neck Plaza,NY,US"
	const testUnits = "f"

	t.Parallel()

	client := NewClient("DummyAPIKey")

	want := "https://api.openweathermap.org/data/2.5/forecast/?q=Great+Neck+Plaza%2CNY%2CUS&appid=DummyAPIKey&units=imperial&cnt=1"
	got, err := client.formAPIUrl(testCity, testUnits)

	if err != nil {
		t.Errorf("Errorwhile testing formAPIUrl(%s, %s): %v\n", testCity, testUnits, err)
	}

	if want != got {
		t.Errorf("Want %q, got %q, testing formAPIUrl(%s, %s)\n", want, got, testCity, testUnits)
	}
}
