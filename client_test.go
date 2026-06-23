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
