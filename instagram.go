package socialapis

import (
	"context"
	"encoding/json"
)

// Instagram is the synchronous client for the Instagram surface of the
// SocialAPIs.io REST API. Same shape as Facebook — every method takes
// ctx context.Context as the first argument.
//
//	ig, err := socialapis.NewInstagram("YOUR_API_TOKEN")
//	profile, err := ig.GetProfileDetails(ctx, "instagram", nil)
type Instagram struct {
	*baseConfig
}

// NewInstagram constructs an Instagram client.
func NewInstagram(apiToken string, opts ...Option) (*Instagram, error) {
	cfg, err := newBaseConfig(apiToken, opts...)
	if err != nil {
		return nil, err
	}
	return &Instagram{baseConfig: cfg}, nil
}

// =====================================================================
// PROFILES
// =====================================================================

// GetUserID returns the numeric Instagram user ID for a username or URL.
func (i *Instagram) GetUserID(ctx context.Context, profile string, extra map[string]any) (Response, error) {
	link, err := asInstagramURL(profile)
	if err != nil {
		return nil, err
	}
	out := Response{}
	if err := i.get(ctx, "/instagram/user/id", mergeParams(map[string]string{"link": link}, extra), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetProfileDetails returns typed ProfileInfo for an Instagram username.
// Unwraps the API's "data"-keyed envelope before populating.
func (i *Instagram) GetProfileDetails(ctx context.Context, username string, extra map[string]any) (*ProfileInfo, error) {
	var raw map[string]json.RawMessage
	if err := i.get(ctx, "/instagram/profile/details", mergeParams(map[string]string{"username": username}, extra), &raw); err != nil {
		return nil, err
	}
	payload, ok := raw["data"]
	if !ok {
		// Fall back to the raw envelope if upstream ever drops the wrapper.
		merged, _ := json.Marshal(raw)
		var pi ProfileInfo
		_ = json.Unmarshal(merged, &pi)
		return &pi, nil
	}
	var pi ProfileInfo
	if err := json.Unmarshal(payload, &pi); err != nil {
		return nil, &ConnectionError{Message: "failed to decode ProfileInfo", Cause: err}
	}
	return &pi, nil
}

// GetProfilePosts returns recent posts from an Instagram profile.
func (i *Instagram) GetProfilePosts(ctx context.Context, username string, extra map[string]any) (Response, error) {
	out := Response{}
	if err := i.get(ctx, "/instagram/profile/posts", mergeParams(map[string]string{"username": username}, extra), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetProfileReels returns Reels for an Instagram profile. Takes a
// numeric user_id (use GetUserID to resolve a username first).
func (i *Instagram) GetProfileReels(ctx context.Context, userID string, extra map[string]any) (Response, error) {
	out := Response{}
	if err := i.get(ctx, "/instagram/profile/reels", mergeParams(map[string]string{"user_id": userID}, extra), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetProfileHighlights returns all Story Highlights for a profile.
func (i *Instagram) GetProfileHighlights(ctx context.Context, userID string, extra map[string]any) (Response, error) {
	out := Response{}
	if err := i.get(ctx, "/instagram/profile/highlights", mergeParams(map[string]string{"user_id": userID}, extra), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetHighlightDetails returns all stories within a specific Highlight.
func (i *Instagram) GetHighlightDetails(ctx context.Context, highlightID string, extra map[string]any) (Response, error) {
	out := Response{}
	if err := i.get(ctx, "/instagram/highlight/details", mergeParams(map[string]string{"highlight_id": highlightID}, extra), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// =====================================================================
// POSTS
// =====================================================================

// GetPostID extracts the shortcode/ID from any Instagram post URL.
func (i *Instagram) GetPostID(ctx context.Context, post string, extra map[string]any) (Response, error) {
	out := Response{}
	if err := i.get(ctx, "/instagram/post/id", mergeParams(map[string]string{"link": post}, extra), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetPostDetails returns full Instagram post details.
func (i *Instagram) GetPostDetails(ctx context.Context, shortcode string, extra map[string]any) (Response, error) {
	out := Response{}
	if err := i.get(ctx, "/instagram/post/details", mergeParams(map[string]string{"shortcode": shortcode}, extra), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// =====================================================================
// REELS
// =====================================================================

// GetReelsFeed returns the trending Reels feed.
func (i *Instagram) GetReelsFeed(ctx context.Context, extra map[string]any) (Response, error) {
	out := Response{}
	if err := i.get(ctx, "/instagram/reels/feed", mergeParams(nil, extra), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetReelsByAudio returns all Reels using a specific audio track.
func (i *Instagram) GetReelsByAudio(ctx context.Context, audioID string, extra map[string]any) (Response, error) {
	out := Response{}
	if err := i.get(ctx, "/instagram/reels/audio", mergeParams(map[string]string{"audio_id": audioID}, extra), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// =====================================================================
// SEARCH + LOCATIONS
// =====================================================================

// Search returns popular Instagram results — users, hashtags, places.
func (i *Instagram) Search(ctx context.Context, keyword string, extra map[string]any) (Response, error) {
	out := Response{}
	if err := i.get(ctx, "/instagram/search", mergeParams(map[string]string{"keyword": keyword}, extra), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetLocationPosts returns posts tagged at a specific Instagram location.
// Pass tab="ranked" or tab="recent" via extra.
func (i *Instagram) GetLocationPosts(ctx context.Context, locationID string, extra map[string]any) (Response, error) {
	out := Response{}
	if err := i.get(ctx, "/instagram/location/posts", mergeParams(map[string]string{"location_id": locationID}, extra), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetNearbyLocations returns Instagram locations near a given location.
func (i *Instagram) GetNearbyLocations(ctx context.Context, locationID string, extra map[string]any) (Response, error) {
	out := Response{}
	if err := i.get(ctx, "/instagram/location/nearby", mergeParams(map[string]string{"location_id": locationID}, extra), &out); err != nil {
		return nil, err
	}
	return out, nil
}
