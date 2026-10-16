package socialapis

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// Mirrors the real envelope: payload under string key "0".
var samplePagePayload = map[string]any{
	"ad_page_id":              "206441436112629",
	"user_id":                 "100064888920170",
	"title":                   "Engen SA | Cape Town",
	"url":                     "https://www.facebook.com/EngenSA",
	"category":                []string{"Petroleum Service"},
	"bio":                     "Energy that drives Africa forward.",
	"followers_count":         119000,
	"likes_count":             1234567,
	"image":                   "https://scontent.fbcdn.net/profile.jpg",
	"is_business_page_active": false,
}

func samplePageResponse() []byte {
	body := map[string]any{
		"0":       samplePagePayload,
		"message": "OK",
		"meta":    map[string]any{"statusCode": 200, "creditsCharged": 1},
	}
	raw, _ := json.Marshal(body)
	return raw
}

// newTestFacebook spins up an httptest server with the given handler
// and returns a Facebook client configured to talk to it.
func newTestFacebook(t *testing.T, handler http.HandlerFunc) (*Facebook, *httptest.Server) {
	t.Helper()
	srv := httptest.NewServer(handler)
	fb, err := NewFacebook("test-token", WithBaseURL(srv.URL))
	if err != nil {
		srv.Close()
		t.Fatalf("NewFacebook: %v", err)
	}
	return fb, srv
}

func TestFacebookGetPageInfoReturnsTypedModel(t *testing.T) {
	fb, srv := newTestFacebook(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(samplePageResponse())
	})
	defer srv.Close()

	page, err := fb.GetPageInfo(context.Background(), "EngenSA", nil)
	if err != nil {
		t.Fatalf("GetPageInfo: %v", err)
	}
	if page.AdPageID != "206441436112629" {
		t.Errorf("AdPageID = %q, want 206441436112629", page.AdPageID)
	}
	if page.Title != "Engen SA | Cape Town" {
		t.Errorf("Title = %q", page.Title)
	}
	if page.FollowersCount != 119000 {
		t.Errorf("FollowersCount = %d, want 119000", page.FollowersCount)
	}
	if page.LikesCount != 1234567 {
		t.Errorf("LikesCount = %d", page.LikesCount)
	}
}

func TestFacebookGetPageInfoNormalisesSlug(t *testing.T) {
	var capturedQuery string
	fb, srv := newTestFacebook(t, func(w http.ResponseWriter, r *http.Request) {
		capturedQuery = r.URL.RawQuery
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(samplePageResponse())
	})
	defer srv.Close()

	if _, err := fb.GetPageInfo(context.Background(), "EngenSA", nil); err != nil {
		t.Fatalf("GetPageInfo: %v", err)
	}
	want := "link=" + "https%3A%2F%2Fwww.facebook.com%2FEngenSA"
	if !strings.Contains(capturedQuery, want) {
		t.Errorf("query = %q, want substring %q", capturedQuery, want)
	}
}

func TestFacebookSendsAuthHeader(t *testing.T) {
	var capturedToken string
	fb, srv := newTestFacebook(t, func(w http.ResponseWriter, r *http.Request) {
		capturedToken = r.Header.Get("x-api-token")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(samplePageResponse())
	})
	defer srv.Close()

	if _, err := fb.GetPageInfo(context.Background(), "EngenSA", nil); err != nil {
		t.Fatalf("GetPageInfo: %v", err)
	}
	if capturedToken != "test-token" {
		t.Errorf("x-api-token = %q, want test-token", capturedToken)
	}
}

func TestNewFacebookEmptyTokenFails(t *testing.T) {
	if _, err := NewFacebook(""); err == nil {
		t.Fatal("NewFacebook with empty token should fail, got nil error")
	}
}

func TestSearchAdsForwardsExtraParams(t *testing.T) {
	var capturedQuery string
	fb, srv := newTestFacebook(t, func(w http.ResponseWriter, r *http.Request) {
		capturedQuery = r.URL.RawQuery
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ads": []}`))
	})
	defer srv.Close()

	_, err := fb.SearchAds(context.Background(), "fitness", map[string]any{
		"country":      "US",
		"activeStatus": "Active",
	})
	if err != nil {
		t.Fatalf("SearchAds: %v", err)
	}
	for _, want := range []string{"query=fitness", "country=US", "activeStatus=Active"} {
		if !strings.Contains(capturedQuery, want) {
			t.Errorf("query = %q missing %q", capturedQuery, want)
		}
	}
}

func TestErrorMapping401(t *testing.T) {
	fb, srv := newTestFacebook(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error": "Invalid API token"}`))
	})
	defer srv.Close()

	_, err := fb.GetPageInfo(context.Background(), "EngenSA", nil)
	var authErr *AuthenticationError
	if !errors.As(err, &authErr) {
		t.Fatalf("expected *AuthenticationError, got %T: %v", err, err)
	}
	if authErr.StatusCode != 401 {
		t.Errorf("StatusCode = %d", authErr.StatusCode)
	}
}

func TestErrorMapping402(t *testing.T) {
	fb, srv := newTestFacebook(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusPaymentRequired)
		_, _ = w.Write([]byte(`{"error": "Out of credits"}`))
	})
	defer srv.Close()

	_, err := fb.GetPageInfo(context.Background(), "EngenSA", nil)
	var insufficient *InsufficientCreditsError
	if !errors.As(err, &insufficient) {
		t.Fatalf("expected *InsufficientCreditsError, got %T: %v", err, err)
	}
}

func TestErrorMapping429WithRetryAfter(t *testing.T) {
	fb, srv := newTestFacebook(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("retry-after", "12")
		w.WriteHeader(http.StatusTooManyRequests)
		_, _ = w.Write([]byte(`{"error": "Rate limited"}`))
	})
	defer srv.Close()

	_, err := fb.GetPageInfo(context.Background(), "EngenSA", nil)
	var rateErr *RateLimitError
	if !errors.As(err, &rateErr) {
		t.Fatalf("expected *RateLimitError, got %T: %v", err, err)
	}
	if rateErr.RetryAfterSeconds != 12 {
		t.Errorf("RetryAfterSeconds = %f, want 12", rateErr.RetryAfterSeconds)
	}
}

func TestErrorMapping400(t *testing.T) {
	fb, srv := newTestFacebook(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error": "Bad input"}`))
	})
	defer srv.Close()

	_, err := fb.GetPageInfo(context.Background(), "EngenSA", nil)
	var badReq *BadRequestError
	if !errors.As(err, &badReq) {
		t.Fatalf("expected *BadRequestError, got %T: %v", err, err)
	}
}

func TestGetGroupMetadataCallsGroupDetailsEndpoint(t *testing.T) {
	var gotPath string
	fb, srv := newTestFacebook(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"name":"Test Group","id":"123"}`))
	})
	defer srv.Close()

	out, err := fb.GetGroupMetadata(context.Background(), "https://www.facebook.com/groups/123", nil)
	if err != nil {
		t.Fatalf("GetGroupMetadata: %v", err)
	}
	if !strings.HasSuffix(gotPath, "/facebook/groups/details") {
		t.Errorf("path = %q, want suffix /facebook/groups/details", gotPath)
	}
	if out["name"] != "Test Group" {
		t.Errorf("name = %v, want Test Group", out["name"])
	}
}

func TestAlbumMethodsCallAlbumEndpoints(t *testing.T) {
	cases := []struct {
		name     string
		call     func(fb *Facebook) (Response, error)
		path     string
		param    string
		value    string
		extraKey string
		extraVal string
	}{
		{"GetPageAlbums", func(fb *Facebook) (Response, error) {
			return fb.GetPageAlbums(context.Background(), "100044395330646", map[string]any{"limit": 16})
		}, "/facebook/pages/albums", "profile_id", "100044395330646", "limit", "16"},
		{"GetGroupAlbums", func(fb *Facebook) (Response, error) {
			return fb.GetGroupAlbums(context.Background(), "187988788687356", nil)
		}, "/facebook/groups/albums", "group_id", "187988788687356", "", ""},
		{"GetAlbumItems", func(fb *Facebook) (Response, error) {
			return fb.GetAlbumItems(context.Background(), "826298618859975", map[string]any{"end_cursor": "abc"})
		}, "/facebook/albums/items", "album_id", "826298618859975", "end_cursor", "abc"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var gotPath string
			var gotQuery map[string][]string
			fb, srv := newTestFacebook(t, func(w http.ResponseWriter, r *http.Request) {
				gotPath = r.URL.Path
				gotQuery = r.URL.Query()
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{"items":[]}`))
			})
			defer srv.Close()

			if _, err := tc.call(fb); err != nil {
				t.Fatalf("%s: %v", tc.name, err)
			}
			if !strings.HasSuffix(gotPath, tc.path) {
				t.Errorf("path = %q, want suffix %s", gotPath, tc.path)
			}
			if got := gotQuery[tc.param]; len(got) != 1 || got[0] != tc.value {
				t.Errorf("%s = %v, want %s", tc.param, got, tc.value)
			}
			if tc.extraKey != "" {
				if got := gotQuery[tc.extraKey]; len(got) != 1 || got[0] != tc.extraVal {
					t.Errorf("%s = %v, want %s", tc.extraKey, got, tc.extraVal)
				}
			}
		})
	}
}

func TestGetGroupVideosSendsLink(t *testing.T) {
	var gotQuery map[string][]string
	fb, srv := newTestFacebook(t, func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.Query()
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"videos":[]}`))
	})
	defer srv.Close()

	if _, err := fb.GetGroupVideos(context.Background(), "187988788687356", nil); err != nil {
		t.Fatalf("GetGroupVideos: %v", err)
	}
	if got := gotQuery["link"]; len(got) != 1 || got[0] != "https://www.facebook.com/groups/187988788687356" {
		t.Errorf("link = %v", got)
	}
	if _, ok := gotQuery["group_id"]; ok {
		t.Errorf("group_id should not be sent")
	}
}
