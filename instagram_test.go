package socialapis

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// Mirrors the real envelope: payload under "data".
var sampleProfilePayload = map[string]any{
	"id":              "25025320",
	"pk":              "25025320",
	"fbid":            "17841400039600391",
	"username":        "instagram",
	"full_name":       "Instagram",
	"biography":       "Discover what's new on Instagram",
	"followers_count": 685_000_000,
	"following_count": 229,
	"media_count":     7900,
	"is_verified":     true,
	"is_private":      false,
	"profile_pic_url": "https://scontent.cdninstagram.com/profile.jpg",
}

func sampleProfileResponse() []byte {
	body := map[string]any{
		"success": true,
		"data":    sampleProfilePayload,
		"message": "OK",
		"meta":    map[string]any{"statusCode": 200},
	}
	raw, _ := json.Marshal(body)
	return raw
}

func newTestInstagram(t *testing.T, handler http.HandlerFunc) (*Instagram, *httptest.Server) {
	t.Helper()
	srv := httptest.NewServer(handler)
	ig, err := NewInstagram("test-token", WithBaseURL(srv.URL))
	if err != nil {
		srv.Close()
		t.Fatalf("NewInstagram: %v", err)
	}
	return ig, srv
}

func TestInstagramGetProfileDetailsReturnsTypedModel(t *testing.T) {
	ig, srv := newTestInstagram(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(sampleProfileResponse())
	})
	defer srv.Close()

	profile, err := ig.GetProfileDetails(context.Background(), "instagram", nil)
	if err != nil {
		t.Fatalf("GetProfileDetails: %v", err)
	}
	if profile.ID != "25025320" {
		t.Errorf("ID = %q", profile.ID)
	}
	if profile.Username != "instagram" {
		t.Errorf("Username = %q", profile.Username)
	}
	if profile.FullName != "Instagram" {
		t.Errorf("FullName = %q", profile.FullName)
	}
	if profile.FollowersCount != 685_000_000 {
		t.Errorf("FollowersCount = %d, want 685000000", profile.FollowersCount)
	}
	if profile.MediaCount != 7900 {
		t.Errorf("MediaCount = %d", profile.MediaCount)
	}
	if !profile.IsVerified {
		t.Errorf("IsVerified = false, want true")
	}
}

func TestInstagramSearchHitsSearchEndpoint(t *testing.T) {
	var capturedPath, capturedQuery string
	ig, srv := newTestInstagram(t, func(w http.ResponseWriter, r *http.Request) {
		capturedPath = r.URL.Path
		capturedQuery = r.URL.RawQuery
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"users": []}`))
	})
	defer srv.Close()

	if _, err := ig.Search(context.Background(), "travel", nil); err != nil {
		t.Fatalf("Search: %v", err)
	}
	if capturedPath != "/instagram/search" {
		t.Errorf("path = %q, want /instagram/search", capturedPath)
	}
	if !strings.Contains(capturedQuery, "keyword=travel") {
		t.Errorf("query = %q missing keyword=travel", capturedQuery)
	}
}

func TestInstagramGetLocationPostsForwardsTabKwarg(t *testing.T) {
	var capturedQuery string
	ig, srv := newTestInstagram(t, func(w http.ResponseWriter, r *http.Request) {
		capturedQuery = r.URL.RawQuery
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"posts": []}`))
	})
	defer srv.Close()

	if _, err := ig.GetLocationPosts(context.Background(), "454547536", map[string]any{"tab": "ranked"}); err != nil {
		t.Fatalf("GetLocationPosts: %v", err)
	}
	for _, want := range []string{"location_id=454547536", "tab=ranked"} {
		if !strings.Contains(capturedQuery, want) {
			t.Errorf("query = %q missing %q", capturedQuery, want)
		}
	}
}
