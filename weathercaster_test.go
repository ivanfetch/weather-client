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

	client := weatherCaster.NewClient(testApiKey)

	want := fmt.Sprintf("https://%s%s/?q=%s&appid=%s%s", client.ApiHost, client.ApiUri, url.QueryEscape(testCity), testApiKey, client.ApiQueryOptions)
	got := client.FormUrl(testCity)

	if want != got {
		t.Errorf("Want %q, got %q, testing formUrl(%s, %s)\n", want, got, testApiKey, testCity)
	}
}
