package socialapis

import (
	"reflect"
	"testing"
)

// FacebookScraper and InstagramScraper are TYPE ALIASES of Facebook
// and Instagram (defined with `type X = Y`). The assertions below
// verify the alias contract — same underlying type, same methods —
// so accidental decoupling fails CI.

func TestFacebookScraperIsFacebookAlias(t *testing.T) {
	fbType := reflect.TypeOf(Facebook{})
	scraperType := reflect.TypeOf(FacebookScraper{})
	if fbType != scraperType {
		t.Fatalf("FacebookScraper underlying type = %v, want %v", scraperType, fbType)
	}
	fb, err := NewFacebookScraper("t")
	if err != nil {
		t.Fatalf("NewFacebookScraper: %v", err)
	}
	if fb == nil {
		t.Fatal("NewFacebookScraper returned nil")
	}
	// Method set is identical because it's a type alias
	if reflect.TypeOf(fb).Elem() != reflect.TypeOf(Facebook{}) {
		t.Fatal("FacebookScraper does not share Facebook's method set")
	}
}

func TestInstagramScraperIsInstagramAlias(t *testing.T) {
	igType := reflect.TypeOf(Instagram{})
	scraperType := reflect.TypeOf(InstagramScraper{})
	if igType != scraperType {
		t.Fatalf("InstagramScraper underlying type = %v, want %v", scraperType, igType)
	}
	ig, err := NewInstagramScraper("t")
	if err != nil {
		t.Fatalf("NewInstagramScraper: %v", err)
	}
	if ig == nil {
		t.Fatal("NewInstagramScraper returned nil")
	}
}
