package weatherCaster_test

import (
	"fmt"
	"github.com/google/go-cmp/cmp"
	"net/url"
	"testing"
	"weatherCaster"
)

// A non-working weather API key
const testApiKey = "abc123xyz"

// A city used in multiple tests.
const testCity = "Great Neck Plaza,NY,US"

func TestFormUrl(t *testing.T) {
	t.Parallel()

	client := weatherCaster.NewClient(testApiKey)

	want := fmt.Sprintf("https://%s%s/?q=%s&appid=%s%s", client.ApiHost, client.ApiUri, url.QueryEscape(testCity), testApiKey, client.ApiQueryOptions)
	got := client.FormUrl(testCity)

	if want != got {
		t.Errorf("Want %q, got %q, testing formUrl(%s, %s)\n", want, got, testApiKey, testCity)
	}
}

func TestParseJson(t *testing.T) {
	t.Parallel()

	client := weatherCaster.NewClient(testApiKey)

	json := `{"cod":"200","message":0,"cnt":1,"list":[{"dt":1616220000,"main":{"temp":34.47,"feels_like":23.59,"temp_min":33.94,"temp_max":34.47,"pressure":1031,"sea_level":1031,"grnd_level":1027,"humidity":38,"temp_kf":0.29},"weather":[{"id":800,"main":"Clear","description":"clear sky","icon":"01n"}],"clouds":{"all":1},"wind":{"speed":9.22,"deg":5},"visibility":10000,"pop":0,"sys":{"pod":"n"},"dt_txt":"2021-03-20 06:00:00"}],"city":{"id":5119226,"name":"Great Neck Plaza","coord":{"lat":40.7868,"lon":-73.7265},"country":"US","population":6707,"timezone":-14400,"sunrise":1616151573,"sunset":1616195137}}`
	// Ideally I populate `want` more directly,
	// E.G. want := &weatherCaster.ApiResponse{List:[{Weather:[{Description:clear sky}]}]}
	// but I can't quite get that syntax correct, so...:
	//
	// Construct each layer of `want`.
	want := weatherCaster.ApiResponse{}
	want.List = make([]weatherCaster.ApiResponseList, 1)
	want.List[0].Weather = make([]weatherCaster.ApiResponseListWeather, 1)
	want.List[0].Weather[0].Description = "clear sky"

	err := client.ParseJson(json)
	if err != nil {
		t.Errorf("Error while calling ParseJson(%v): %v\n", json, err)
	}
	got := client.GetApiResponse()
	if !cmp.Equal(want, got) {
		t.Errorf("want %+v, got %+v, calling ParseJson(%v)\n", want, got, json)
	}
}
