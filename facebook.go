package socialapis

import (
	"context"
	"encoding/json"
)

// Facebook is the synchronous client for the Facebook surface of the
// SocialAPIs.io REST API. Construct with NewFacebook; every method
// takes a context.Context as the first argument (idiomatic Go).
//
//	fb, err := socialapis.NewFacebook("YOUR_API_TOKEN")
//	if err != nil { ... }
//	ctx := context.Background()
//	page, err := fb.GetPageInfo(ctx, "EngenSA", nil)
//
// FacebookScraper is an alias so users migrating from popular but
// abandoned scrapers can keep their import line greppable. See alias.go.
type Facebook struct {
	*baseConfig
}

// NewFacebook constructs a Facebook client. The apiToken is required.
//
// Get a free key (200 calls/month, no card) at
// https://socialapis.io/auth/signup.
func NewFacebook(apiToken string, opts ...Option) (*Facebook, error) {
	cfg, err := newBaseConfig(apiToken, opts...)
	if err != nil {
		return nil, err
	}
	return &Facebook{baseConfig: cfg}, nil
}

// =====================================================================
// PAGES
// =====================================================================

// GetPageID returns the numeric Facebook Page ID for a URL or slug.
// Backed by GET /facebook/pages/id.
func (f *Facebook) GetPageID(ctx context.Context, page string, extra map[string]any) (Response, error) {
	link, err := asFacebookURL(page)
	if err != nil {
		return nil, err
	}
	out := Response{}
	if err := f.get(ctx, "/facebook/pages/id", mergeParams(map[string]string{"link": link}, extra), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetPageInfo returns typed metadata for a Facebook Page.
// Backed by GET /facebook/pages/details. Unwraps the API's "0"-keyed
// envelope before populating PageInfo.
func (f *Facebook) GetPageInfo(ctx context.Context, page string, extra map[string]any) (*PageInfo, error) {
	link, err := asFacebookURL(page)
	if err != nil {
		return nil, err
	}
	var raw map[string]json.RawMessage
	if err := f.get(ctx, "/facebook/pages/details", mergeParams(map[string]string{"link": link}, extra), &raw); err != nil {
		return nil, err
	}
	// The API wraps the payload under string key "0". Fall back to the
	// raw envelope if upstream ever drops the wrapper.
	payload, ok := raw["0"]
	if !ok {
		// Re-serialize raw and decode into PageInfo as fallback
		merged, _ := json.Marshal(raw)
		var pi PageInfo
		_ = json.Unmarshal(merged, &pi)
		return &pi, nil
	}
	var pi PageInfo
	if err := json.Unmarshal(payload, &pi); err != nil {
		return nil, &ConnectionError{Message: "failed to decode PageInfo", Cause: err}
	}
	return &pi, nil
}

// GetPagePosts returns recent posts from a Facebook Page (cursor-paginated).
func (f *Facebook) GetPagePosts(ctx context.Context, page string, extra map[string]any) (Response, error) {
	link, err := asFacebookURL(page)
	if err != nil {
		return nil, err
	}
	out := Response{}
	if err := f.get(ctx, "/facebook/pages/posts", mergeParams(map[string]string{"link": link}, extra), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetPageReels returns Reels (short videos) from a Facebook Page.
func (f *Facebook) GetPageReels(ctx context.Context, page string, extra map[string]any) (Response, error) {
	link, err := asFacebookURL(page)
	if err != nil {
		return nil, err
	}
	out := Response{}
	if err := f.get(ctx, "/facebook/pages/reels", mergeParams(map[string]string{"link": link}, extra), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetPageVideos returns long-form videos from a Facebook Page.
func (f *Facebook) GetPageVideos(ctx context.Context, page string, extra map[string]any) (Response, error) {
	link, err := asFacebookURL(page)
	if err != nil {
		return nil, err
	}
	out := Response{}
	if err := f.get(ctx, "/facebook/pages/videos", mergeParams(map[string]string{"link": link}, extra), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// =====================================================================
// GROUPS
// =====================================================================

// GetGroupID returns the numeric Facebook Group ID.
func (f *Facebook) GetGroupID(ctx context.Context, group string, extra map[string]any) (Response, error) {
	link, err := asFacebookGroupURL(group)
	if err != nil {
		return nil, err
	}
	out := Response{}
	if err := f.get(ctx, "/facebook/groups/id", mergeParams(map[string]string{"link": link}, extra), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetGroupDetails returns typed GroupInfo. This endpoint has NO
// envelope wrapper — payload is at the top level.
func (f *Facebook) GetGroupDetails(ctx context.Context, group string, extra map[string]any) (*GroupInfo, error) {
	link, err := asFacebookGroupURL(group)
	if err != nil {
		return nil, err
	}
	var gi GroupInfo
	if err := f.get(ctx, "/facebook/groups/details", mergeParams(map[string]string{"link": link}, extra), &gi); err != nil {
		return nil, err
	}
	return &gi, nil
}

// GetGroupMetadata returns lightweight Group metadata (name, id, url, image).
func (f *Facebook) GetGroupMetadata(ctx context.Context, group string, extra map[string]any) (Response, error) {
	link, err := asFacebookGroupURL(group)
	if err != nil {
		return nil, err
	}
	out := Response{}
	if err := f.get(ctx, "/facebook/groups/metadata", mergeParams(map[string]string{"link": link}, extra), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetGroupPosts returns recent posts from a Facebook Group.
func (f *Facebook) GetGroupPosts(ctx context.Context, group string, extra map[string]any) (Response, error) {
	link, err := asFacebookGroupURL(group)
	if err != nil {
		return nil, err
	}
	out := Response{}
	if err := f.get(ctx, "/facebook/groups/posts", mergeParams(map[string]string{"link": link}, extra), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetGroupVideos returns videos posted to a Group (takes numeric group_id).
func (f *Facebook) GetGroupVideos(ctx context.Context, groupID string, extra map[string]any) (Response, error) {
	out := Response{}
	if err := f.get(ctx, "/facebook/groups/videos", mergeParams(map[string]string{"group_id": groupID}, extra), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// =====================================================================
// POSTS
// =====================================================================

// GetPostID extracts the numeric Facebook post ID from a post URL.
func (f *Facebook) GetPostID(ctx context.Context, post string, extra map[string]any) (Response, error) {
	out := Response{}
	if err := f.get(ctx, "/facebook/posts/id", mergeParams(map[string]string{"link": post}, extra), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetPostDetails returns full details of a Facebook post.
func (f *Facebook) GetPostDetails(ctx context.Context, post string, extra map[string]any) (Response, error) {
	out := Response{}
	if err := f.get(ctx, "/facebook/posts/details", mergeParams(map[string]string{"link": post}, extra), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetPostDetailsExtended returns extended post details (views, video URLs, music info).
func (f *Facebook) GetPostDetailsExtended(ctx context.Context, post string, extra map[string]any) (Response, error) {
	out := Response{}
	if err := f.get(ctx, "/facebook/posts/details/extended", mergeParams(map[string]string{"link": post}, extra), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetPostComments returns comments on a Facebook post or reel.
// Pass include_reply_info="true" via extra to get reply cursors.
func (f *Facebook) GetPostComments(ctx context.Context, post string, extra map[string]any) (Response, error) {
	out := Response{}
	if err := f.get(ctx, "/facebook/posts/comments", mergeParams(map[string]string{"link": post}, extra), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetCommentReplies returns replies to a specific comment. Both inputs
// come from GetPostComments when called with include_reply_info=true.
func (f *Facebook) GetCommentReplies(ctx context.Context, commentFeedbackID, expansionToken string, extra map[string]any) (Response, error) {
	out := Response{}
	primary := map[string]string{
		"comment_feedback_id": commentFeedbackID,
		"expansion_token":     expansionToken,
	}
	if err := f.get(ctx, "/facebook/posts/comments/replies", mergeParams(primary, extra), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetPostAttachments returns all media attachments from a post.
func (f *Facebook) GetPostAttachments(ctx context.Context, postID string, extra map[string]any) (Response, error) {
	out := Response{}
	if err := f.get(ctx, "/facebook/posts/attachments", mergeParams(map[string]string{"post_id": postID}, extra), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetVideoPostDetails returns title, reactions, and play counts for a video post.
func (f *Facebook) GetVideoPostDetails(ctx context.Context, videoID string, extra map[string]any) (Response, error) {
	out := Response{}
	if err := f.get(ctx, "/facebook/posts/video", mergeParams(map[string]string{"video_id": videoID}, extra), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// =====================================================================
// SEARCH
// =====================================================================

// SearchPages searches Facebook pages by keyword. Pass location_id via
// extra for geo filtering.
func (f *Facebook) SearchPages(ctx context.Context, query string, extra map[string]any) (Response, error) {
	out := Response{}
	if err := f.get(ctx, "/facebook/search/pages", mergeParams(map[string]string{"query": query}, extra), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// SearchPeople searches Facebook profiles by keyword.
func (f *Facebook) SearchPeople(ctx context.Context, query string, extra map[string]any) (Response, error) {
	out := Response{}
	if err := f.get(ctx, "/facebook/search/people", mergeParams(map[string]string{"query": query}, extra), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// SearchLocations searches Facebook for locations matching a keyword.
func (f *Facebook) SearchLocations(ctx context.Context, query string, extra map[string]any) (Response, error) {
	out := Response{}
	if err := f.get(ctx, "/facebook/search/locations", mergeParams(map[string]string{"query": query}, extra), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// SearchPosts searches Facebook posts by keyword.
func (f *Facebook) SearchPosts(ctx context.Context, query string, extra map[string]any) (Response, error) {
	out := Response{}
	if err := f.get(ctx, "/facebook/search/posts", mergeParams(map[string]string{"query": query}, extra), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// SearchVideos searches Facebook videos by keyword.
func (f *Facebook) SearchVideos(ctx context.Context, query string, extra map[string]any) (Response, error) {
	out := Response{}
	if err := f.get(ctx, "/facebook/search/videos", mergeParams(map[string]string{"query": query}, extra), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// =====================================================================
// META ADS LIBRARY
// =====================================================================

// GetAdsCountries returns all country codes supported by the Meta Ads Library.
func (f *Facebook) GetAdsCountries(ctx context.Context, extra map[string]any) (Response, error) {
	out := Response{}
	if err := f.get(ctx, "/facebook/ads/countries", mergeParams(nil, extra), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// SearchAds searches ads in the Meta Ad Library by keyword.
func (f *Facebook) SearchAds(ctx context.Context, query string, extra map[string]any) (Response, error) {
	out := Response{}
	if err := f.get(ctx, "/facebook/ads/search", mergeParams(map[string]string{"query": query}, extra), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetAdsPageDetails returns Ads-Library metadata for a Facebook Page.
func (f *Facebook) GetAdsPageDetails(ctx context.Context, pageID string, extra map[string]any) (Response, error) {
	out := Response{}
	if err := f.get(ctx, "/facebook/ads/page-details", mergeParams(map[string]string{"page_id": pageID}, extra), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetAdArchiveDetails returns detailed info for a specific archived ad.
func (f *Facebook) GetAdArchiveDetails(ctx context.Context, adArchiveID, pageID string, extra map[string]any) (Response, error) {
	out := Response{}
	primary := map[string]string{
		"ad_archive_id": adArchiveID,
		"page_id":       pageID,
	}
	if err := f.get(ctx, "/facebook/ads/archive-details", mergeParams(primary, extra), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// SearchAdsByKeywords searches ads in the Ad Library by keyword + country.
func (f *Facebook) SearchAdsByKeywords(ctx context.Context, query string, extra map[string]any) (Response, error) {
	out := Response{}
	if err := f.get(ctx, "/facebook/ads/keywords", mergeParams(map[string]string{"query": query}, extra), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// =====================================================================
// MARKETPLACE
// =====================================================================

// SearchMarketplace searches Facebook Marketplace listings.
func (f *Facebook) SearchMarketplace(ctx context.Context, query string, extra map[string]any) (Response, error) {
	out := Response{}
	if err := f.get(ctx, "/facebook/marketplace/search", mergeParams(map[string]string{"query": query}, extra), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetListingDetails returns full info for a Marketplace listing.
func (f *Facebook) GetListingDetails(ctx context.Context, listingID string, extra map[string]any) (Response, error) {
	out := Response{}
	if err := f.get(ctx, "/facebook/marketplace/listing", mergeParams(map[string]string{"listing_id": listingID}, extra), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetSellerDetails returns seller profile, ratings, reviews from Marketplace.
func (f *Facebook) GetSellerDetails(ctx context.Context, sellerID string, extra map[string]any) (Response, error) {
	out := Response{}
	if err := f.get(ctx, "/facebook/marketplace/seller", mergeParams(map[string]string{"seller_id": sellerID}, extra), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetMarketplaceCategories returns all Marketplace categories.
func (f *Facebook) GetMarketplaceCategories(ctx context.Context, extra map[string]any) (Response, error) {
	out := Response{}
	if err := f.get(ctx, "/facebook/marketplace/categories", mergeParams(nil, extra), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetCityCoordinates resolves a city name to GPS coordinates.
func (f *Facebook) GetCityCoordinates(ctx context.Context, city string, extra map[string]any) (Response, error) {
	out := Response{}
	if err := f.get(ctx, "/facebook/marketplace/city-coordinates", mergeParams(map[string]string{"city": city}, extra), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// SearchVehicles searches Marketplace vehicle listings. Requires lat/lng filters via extra.
func (f *Facebook) SearchVehicles(ctx context.Context, extra map[string]any) (Response, error) {
	out := Response{}
	if err := f.get(ctx, "/facebook/marketplace/vehicles", mergeParams(nil, extra), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// SearchRentals searches Marketplace rental-property listings.
func (f *Facebook) SearchRentals(ctx context.Context, extra map[string]any) (Response, error) {
	out := Response{}
	if err := f.get(ctx, "/facebook/marketplace/rentals", mergeParams(nil, extra), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// =====================================================================
// MEDIA
// =====================================================================

// DownloadMedia resolves a Facebook video/photo URL to a downloadable URL.
func (f *Facebook) DownloadMedia(ctx context.Context, mediaURL string, extra map[string]any) (Response, error) {
	out := Response{}
	if err := f.get(ctx, "/facebook/media/download", mergeParams(map[string]string{"url": mediaURL}, extra), &out); err != nil {
		return nil, err
	}
	return out, nil
}
