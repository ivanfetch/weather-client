package weather

import (
	"fmt"
	"github.com/google/go-cmp/cmp"
	"net/url"
	"testing"
)

// A non-working weather API key
const testApiKey = "abc123xyz"

// A city used in multiple tests.
const testCity = "Great Neck Plaza,NY,US"

func TestFormAPIUrl(t *testing.T) {
	t.Parallel()

	client := NewClient(testApiKey)

	want := fmt.Sprintf("https://%s%s/?q=%s&appid=%s%s", client.ApiHost, client.ApiUri, url.QueryEscape(testCity), testApiKey, client.ApiQueryOptions)
	got := client.formAPIUrl(testCity)

	if want != got {
		t.Errorf("Want %q, got %q, testing formUrl(%s, %s)\n", want, got, testApiKey, testCity)
	}
}
