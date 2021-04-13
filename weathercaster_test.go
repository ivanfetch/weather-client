package weatherCaster_test

import (
	"fmt"
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

	want := fmt.Sprintf("https://%s%s/?q=%s&appid=%s%s", weatherCaster.ApiHost, weatherCaster.ApiUri, url.QueryEscape(testCity), testApiKey, weatherCaster.ApiQueryOptions)
	got := weatherCaster.FormUrl(testApiKey, testCity)

	if want != got {
		t.Errorf("Want %q, got %q, testing formUrl(%s, %s)\n", want, got, testApiKey, testCity)
	}
}
